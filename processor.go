package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/net/html/atom"

	"golang.org/x/mod/modfile"
	"golang.org/x/net/html"
)

type processor struct {
	syms    symbols
	counter int
	mod     *modfile.File
}

// dummy body node to be used for fragment parsing
var bodyEnv = html.Node{
	Type:     html.ElementNode,
	Data:     "body",
	DataAtom: atom.Body}

func (p *processor) init(outputPath string) bool {
	p.syms.packages = make(map[string]*tbcPackage)
	raw, err := ioutil.ReadFile("go.mod")
	if err != nil {
		if os.IsNotExist(err) {
			os.Stderr.WriteString("[error] did not find go.mod.\n")
			os.Stderr.WriteString(
				"[error] tbc must be run in the root directory of your module.\n")
		} else {
			os.Stderr.WriteString("[error] while reading go.mod: ")
			os.Stderr.WriteString(err.Error() + "\n")
		}
		return false
	}
	p.mod, err = modfile.Parse("go.mod", raw, nil)
	if err != nil {
		os.Stderr.WriteString("[error] unable to read go.mod:\n")
		fmt.Fprintf(os.Stderr, "[error] %s\n", err.Error())
		return false
	}
	p.syms.pkgBasePath = filepath.Join(p.mod.Module.Mod.Path, outputPath)
	return true
}

func (p *processor) process(file string) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(file + ": unable to read file, skipping.")
		return
	}
	nodes, err := html.ParseFragment(bytes.NewReader(contents), &bodyEnv)
	if err != nil {
		log.Printf("%s: failed to parse with error(s):\n  %s\n", file, err.Error())
		return
	}
	{
		if err = processMacros(nodes, &p.syms); err != nil {
			panic(file + err.Error())
		}

		// we need to write out the nodes and parse it again since text nodes may
		// be merged and additional elements may be created now with includes
		// processed. If we don't do this, paths to access the dynamic objects will
		// be wrong.
		b := strings.Builder{}
		for i := range nodes {
			html.Render(&b, nodes[i])
		}
		log.Println("after macro processing:\n" + b.String())
		nodes, err = html.ParseFragment(strings.NewReader(b.String()), &bodyEnv)
		if err != nil {
			panic(err)
		}
	}

	for i := range nodes {
		n := nodes[i]
		switch n.Type {
		case html.TextNode:
			text := strings.TrimSpace(n.Data)
			if len(text) > 0 {
				panic("non-whitespace text at top level: `" + text + "`")
			}
		case html.ErrorNode:
			panic("encountered ErrorNode: " + n.Data)
		case html.ElementNode:
			if n.DataAtom != 0 || n.Data != "tbc:package" {
				panic("only tbc:package is allowed at top level. found <" + n.Data + ">")
			}
			p.syms.curPkg = attrVal(n.Attr, "name")
			pkg, _ := p.syms.packages[p.syms.curPkg]
			pkg.process(&p.syms, n, &p.counter)
		default:
			panic("illegal node at top level: " + n.Data)
		}
	}
}

func writePathLiteral(b *strings.Builder, path []int) {
	for i := range path {
		if i != 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%d", path[i])
	}
}

type goRenderer struct {
	syms        *symbols
	packageName string
	packagePath string
}

func writeQuotedLine(b *strings.Builder, name string) {
	fmt.Fprintf(b, "\"%s\"\n", name)
}

func (r *goRenderer) writeFileHeader(b *strings.Builder, deps map[string]struct{}) {
	fmt.Fprintf(b, "package %s\n\n", r.packageName)
	b.WriteString("import (\n")
	writeQuotedLine(b, "github.com/flyx/tbc/runtime")
	writeQuotedLine(b, "github.com/gopherjs/gopherjs/js")
	for dep := range deps {
		writeQuotedLine(b, dep)
	}
	b.WriteString(")\n")
}

func (r *goRenderer) writeFormatted(goCode string, file string) {
	fmtcmd := exec.Command("gofmt")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fmtcmd.Stdout = &stdout
	fmtcmd.Stderr = &stderr

	stdin, err := fmtcmd.StdinPipe()
	if err != nil {
		panic("unable to create stdin pipe: " + err.Error())
	}
	io.WriteString(stdin, goCode)
	stdin.Close()

	if err := fmtcmd.Run(); err != nil {
		log.Println("error while formatting: " + err.Error())
		log.Println("stderr output:")
		log.Println(stderr.String())
		log.Println("input:")
		log.Println(goCode)
		panic("failed to format Go code")
	}

	if err := ioutil.WriteFile(file, []byte(stdout.String()), os.ModePerm); err != nil {
		panic("failed to write file '" + file + "': " + err.Error())
	}
}

func genAccessor(b *strings.Builder, path []int, bv boundValue) {
	switch bv.kind {
	case boundProperty:
		fmt.Fprintf(b, "runtime.NewBoundProperty(o.root, \"%s\", ", bv.id)
	case boundAttribute:
		fmt.Fprintf(b, "runtime.NewBoundAttribute(o.root, \"%s\", ", bv.id)
	case boundClass:
		fmt.Fprintf(b, "runtime.NewBoundClass(o.root, \"%s\", ", bv.id)
	}
	writePathLiteral(b, path)
	b.WriteString(")")
}

func wrapperForType(k valueKind) string {
	switch k {
	case stringVal:
		return "StringValue"
	case intVal:
		return "IntValue"
	case boolVal:
		return "BoolValue"
	default:
		panic("unsupported type")
	}
}

func nameForType(k valueKind) string {
	switch k {
	case stringVal:
		return "string"
	case intVal:
		return "int"
	case boolVal:
		return "bool"
	default:
		panic("unsupported type")
	}
}

func nameForBound(b boundKind) string {
	switch b {
	case boundAttribute:
		return "BoundAttribute"
	case boundProperty:
		return "BoundProperty"
	case boundClass:
		return "BoundClass"
	default:
		panic("unknown boundKind")
	}
}

func (r *goRenderer) writeComponentFile(name string, c *component) {
	b := strings.Builder{}
	r.writeFileHeader(&b, c.dependencies)
	if c.needsList && c.handlers != nil {
		fmt.Fprintf(&b, "// %sController is the interface for handling events captured from %s\n", name, name)
		fmt.Fprintf(&b, "type %sController interface {\n", name)
		for hName, h := range c.handlers {
			fmt.Fprintf(&b, "%s(", hName)
			first := true
			for pName, pType := range h.params {
				if first {
					first = false
				} else {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%s %s", pName, nameForType(pType))
			}
			b.WriteString(") bool\n")
		}
		b.WriteString("}\n")
	}

	fmt.Fprintf(&b, "// %s is a DOM component autogenerated by TBC.\n", name)
	fmt.Fprintf(&b, "type %s struct {\n", name)
	b.WriteString("root *js.Object\n")
	if c.needsController && c.handlers != nil {
		fmt.Fprintf(&b, "c %sController\n", name)
	}
	for _, a := range c.accessors {
		fmt.Fprintf(&b, "%s runtime.%s\n", a.variable.name, wrapperForType(a.variable.kind))
	}
	for _, e := range c.embeds {
		fmt.Fprintf(&b, "%s ", e.field)
		if !e.list {
			b.WriteRune('*')
		}
		if e.pkg != "" {
			fmt.Fprintf(&b, "%s.", e.pkg)
		}
		if e.list {
			fmt.Fprintf(&b, "%sList\n", e.t)
		} else {
			fmt.Fprintf(&b, "%s\n", e.t)
		}
	}

	fmt.Fprintf(&b, "}\n// New%s creates a new component and initializes it with Init.\n", name)
	fmt.Fprintf(&b, "func New%s() *%s {\n", name, name)
	fmt.Fprintf(&b, "ret := new(%s)\n", name)
	b.WriteString("ret.Init()\nreturn ret\n}\n")
	b.WriteString("// Init initializes the component, discarding all previous information.\n")
	b.WriteString("// The component is initially a DocumentFragment until it gets inserted into\n")
	b.WriteString("// the main document. It can be manipulated both before and after insertion.\n")
	fmt.Fprintf(&b, "func (o *%s) Init() {\n", name)
	fmt.Fprintf(&b, "o.root = runtime.InstantiateTemplateByID(\"%s\")\n", c.id)
	for _, a := range c.accessors {
		fmt.Fprintf(&b, "o.%s.BoundValue = ", a.variable.name)
		genAccessor(&b, a.path, a.target)
		b.WriteString("\n")
	}
	for _, e := range c.embeds {
		b.WriteString("{\ncontainer := runtime.WalkPath(o.root, ")
		writePathLiteral(&b, e.path[:len(e.path)-1])
		if e.list {
			fmt.Fprintf(&b, ")\no.%s.Init(container, %d)\n", e.field, e.path[len(e.path)-1])
		} else {
			fmt.Fprintf(&b, ")\no.%s = ", e.field)
			if e.pkg != "" {
				fmt.Fprintf(&b, "%s.", e.pkg)
			}
			fmt.Fprintf(&b, "New%s()\n", e.t)
			fmt.Fprintf(&b, "o.%s.InsertInto(container, container.Get(\"childNodes\").Index(%d))\n",
				e.field, e.path[len(e.path)-1])
		}
		b.WriteString("}\n")
	}
	for _, src := range c.captureSources {
		b.WriteString("{\nsrc := runtime.WalkPath(o.root, ")
		writePathLiteral(&b, src.path)
		b.WriteString(")\n")
		for _, cap := range src.captures {
			b.WriteString("{\nwrapper := js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {\n")
			for pName, bVal := range cap.paramMappings {
				fmt.Fprintf(&b, "var p%s runtime.%s\n", pName, nameForBound(bVal.kind))
				fmt.Fprintf(&b, "p%s.Init(this, \"%s\")\n", pName, bVal.id)
			}
			fmt.Fprintf(&b, "if o.call%s(", cap.handler)
			first := true
			for pName := range cap.paramMappings {
				if first {
					first = false
				} else {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "&p%s", pName)
			}
			b.WriteString(") {\narguments[0].Call(\"preventDefault\")\n}\nreturn nil\n})\n")
			fmt.Fprintf(&b, "src.Call(\"addEventListener\", \"%s\", wrapper)\n", cap.event)
			b.WriteString("}\n")
		}
		b.WriteString("}\n")
	}

	b.WriteString("}\n// InsertInto inserts this component into the given object. This can only\n")
	b.WriteString("// be done once. The nodes will be inserted in front of `before`, or\n")
	b.WriteString("// at the end if `before` is `nil`.")
	fmt.Fprintf(&b, "\nfunc (o *%s) InsertInto(parent *js.Object, before *js.Object) {\n", name)
	b.WriteString("parent.Call(\"insertBefore\", o.root, before)\n")
	for _, e := range c.embeds {
		if e.list {
			fmt.Fprintf(&b, "o.%s.mgr.UpdateParent(o.root, parent, before)\n", e.field)
		}
	}
	b.WriteString("}\n")

	if c.handlers != nil {
		if c.needsController {
			b.WriteString("// SetController defines which object handles the captured events\n")
			b.WriteString("// of this component. If set to nil, default behavior will take over.\n")
			fmt.Fprintf(&b, "func (o *%s) SetController(c %sController) {\n", name, name)
			b.WriteString("o.c = c\n}\n")
		}
		for hName, h := range c.handlers {
			fmt.Fprintf(&b, "func (o *%s) call%s(", name, hName)
			first := true
			for pName := range h.params {
				if first {
					first = false
				} else {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%s runtime.BoundValue", pName)
			}
			b.WriteString(") bool {\n")
			if c.needsController {
				b.WriteString("if o.c == nil {\nreturn false\n}\n")
			}
			for pName, pType := range h.params {
				fmt.Fprintf(&b, "_%s := runtime.%s{BoundValue: %s}\n",
					pName, wrapperForType(pType), pName)
			}
			if c.needsController {
				fmt.Fprintf(&b, "return o.c.%s(", hName)
			} else {
				fmt.Fprintf(&b, "return o.%s(", hName)
			}
			first = true
			for pName := range h.params {
				if first {
					first = false
				} else {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "_%s.Get()", pName)
			}
			b.WriteString(")\n}\n")
		}
	}

	if c.needsList {
		fmt.Fprintf(&b, "// %sList is a list of %s whose manipulation methods auto-update\n", name, name)
		b.WriteString("// the corresponding nodes in the document.\n")
		fmt.Fprintf(&b, "type %sList struct {\n", name)
		b.WriteString("mgr runtime.ListManager\n")
		fmt.Fprintf(&b, "items []*%s\n", name)
		b.WriteString("}\n")
		b.WriteString("// Init initializes the list, discarding previous data.\n")
		b.WriteString("// The list is initially a DocumentFragment until it gets inserted into\n")
		b.WriteString("// the main document. It can be manipulated both before and after insertion.\n")
		fmt.Fprintf(&b, "func (l *%sList) Init(container *js.Object, index int) {\n", name)
		b.WriteString("l.mgr = runtime.CreateListManager(container, index)\n")
		b.WriteString("l.items = nil\n}\n")
		b.WriteString("// Len returns the number of items in the list.\n")
		fmt.Fprintf(&b, "func (l *%sList) Len() int {\n", name)
		b.WriteString("return len(l.items)\n}\n")
		b.WriteString("// Item returns the item at the current index.\n")
		fmt.Fprintf(&b, "func (l *%sList) Item(index int) *%s{\n", name, name)
		b.WriteString("return l.items[index]\n}\n")
		b.WriteString("// Append initializes a new item, appends it to the list and returns it.\n")
		fmt.Fprintf(&b, "func (l *%sList) Append() (ret *%s) {\n", name, name)
		fmt.Fprintf(&b, "ret = New%s()\n", name)
		b.WriteString("l.items = append(l.items, ret)\n")
		b.WriteString("l.mgr.Append(ret.root)\n")
		b.WriteString("return\n}\n")
		b.WriteString("// Insert initializes a new item, inserts it into the list and returns it.\n")
		fmt.Fprintf(&b, "func (l *%sList) Insert(index int) (ret *%s) {\n", name, name)
		b.WriteString("var prev *js.Object\n")
		b.WriteString("if index < len(l.items) {\nprev = l.items[index].root\n}\n")
		fmt.Fprintf(&b, "ret = New%s()\n", name)
		b.WriteString("l.items = append(l.items, nil)\n")
		b.WriteString("copy(l.items[index+1:], l.items[index:])\n")
		b.WriteString("l.items[index] = ret\n")
		b.WriteString("l.mgr.Insert(ret.root, prev)\n")
		b.WriteString("return\n}\n")
		b.WriteString("// Remove removes the item at the given index from the list.\n")
		fmt.Fprintf(&b, "func (l *%sList) Remove(index int) {\n", name)
		b.WriteString("l.mgr.Remove(l.items[index].root)\n")
		b.WriteString("copy(l.items[index:], l.items[index+1:])\n")
		b.WriteString("l.items = l.items[:len(l.items)-1]\n")
		b.WriteString("}\n")
	}

	r.writeFormatted(b.String(), filepath.Join(r.packagePath, strings.ToLower(name)+".go"))
}

func (p *processor) dump(htmlPath, packageParent string) {
	htmlFile, err := os.Create(htmlPath)
	if err != nil {
		panic("unable to write HTML output: " + err.Error())
	}
	for _, pkg := range p.syms.packages {
		for _, c := range pkg.components {
			html.Render(htmlFile, c.processedHTML)
		}
	}

	htmlFile.Close()

	for pkgName, pkg := range p.syms.packages {
		renderer := goRenderer{syms: &p.syms, packageName: pkgName,
			packagePath: filepath.Join(packageParent, pkgName)}
		if err := os.MkdirAll(renderer.packagePath, os.ModePerm); err != nil {
			panic("failed to create package directory '" + renderer.packagePath +
				"': " + err.Error())
		}
		for name, t := range pkg.components {
			renderer.writeComponentFile(name, t)
		}
	}
}
