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
	"reflect"
	"strings"

	"golang.org/x/net/html/atom"

	"golang.org/x/net/html"
)

type processor struct {
	templates templateSet
	counter   int
}

func (p *processor) processTemplate(n *html.Node) {
	var tmplAttrs templateAttribs
	extractTbcAttribs(n, &tmplAttrs)
	if len(tmplAttrs.name) == 0 {
		panic("<template> must have tbc:name!")
	}
	if attrExists(n.Attr, "id") {
		panic("<template> may not have id (id is generated by tbc)")
	}

	tmpl := &template{strippedHTML: n}
	p.counter++
	tmpl.id = fmt.Sprintf("tbc-component-%d-%s", p.counter, strings.ToLower(tmplAttrs.name))
	n.Attr = append(n.Attr, html.Attribute{Key: "id", Val: tmpl.id})
	indexList := make([]int, 1, 32)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		tmpl.process(p.templates, c, indexList)
		indexList[0]++
	}
	p.templates[tmplAttrs.name] = tmpl
}

// dummy body node to be used for fragment parsing
var bodyEnv = html.Node{
	Type:     html.ElementNode,
	Data:     "body",
	DataAtom: atom.Body}

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
		ip := includesProcessor{}
		ip.process(&nodes)

		// we need to write out the nodes and parse it again since text nodes may
		// be merged and additional elements may be created now with includes
		// processed. If we don't do this, paths to access the dynamic objects will
		// be wrong.
		b := strings.Builder{}
		for i := range nodes {
			html.Render(&b, nodes[i])
		}
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
			if n.DataAtom != atom.Template {
				panic("non-template element at top level: <" + n.Data + ">")
			}
			p.processTemplate(n)
		default:
			panic("illegal node at top level: " + n.Data)
		}
	}
}

func writePathLiteral(b *strings.Builder, path []int) {
	b.WriteString("[]int{")
	for i := range path {
		fmt.Fprintf(b, "%d, ", path[i])
	}
	b.WriteByte('}')
}

type goRenderer struct {
	templates   templateSet
	packageName string
	packagePath string
}

func (r *goRenderer) writeFileHeader(b *strings.Builder) {
	fmt.Fprintf(b, "package %s\n\n", r.packageName)
	b.WriteString("import (\n\"github.com/flyx/tbc/runtime\"\n\"github.com/gopherjs/gopherjs/js\"\n)\n")
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

	ioutil.WriteFile(file, []byte(stdout.String()), os.ModePerm)
}

func genPropertyAccessor(b *strings.Builder, path []int, propName string) {
	b.WriteString("runtime.NewPropertyAccessor(o.root, ")
	writePathLiteral(b, path)
	fmt.Fprintf(b, ", \"%s\")", propName)
}

func (r *goRenderer) writeComponentFile(name string, c *template) {
	b := strings.Builder{}
	r.writeFileHeader(&b)
	fmt.Fprintf(&b, "// %s is a DOM component autogenerated by TBC.\n", name)
	fmt.Fprintf(&b, "type %s struct {\n", name)
	b.WriteString("root *js.Object\n")
	for i := range c.objects {
		o := c.objects[i]
		b.WriteString(o.goName)
		b.WriteByte(' ')
		switch o.goType {
		case reflect.Int:
			b.WriteString("runtime.IntValue\n")
		case reflect.String:
			b.WriteString("runtime.StringValue\n")
		case reflect.Bool:
			b.WriteString("runtime.BoolValue\n")
		default:
			panic("unexpected type of dynamic object")
		}
	}
	for i := range c.embeds {
		e := c.embeds[i]
		if e.list {
			fmt.Fprintf(&b, "%s %sList\n", e.fieldName, e.goName)
		} else {
			fmt.Fprintf(&b, "%s *%s\n", e.fieldName, e.goName)
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
	for i := range c.objects {
		o := c.objects[i]
		fmt.Fprintf(&b, "o.%s.ScalarAccessor = ", o.goName)
		switch o.kind {
		case textContent:
			genPropertyAccessor(&b, o.path, "textContent")
		case inputValue:
			genPropertyAccessor(&b, o.path, "value")
		case classSwitch:
			b.WriteString("runtime.NewClassSwitcher(o.root, ")
			writePathLiteral(&b, o.path)
			fmt.Fprintf(&b, ", \"%s\")", o.className)
		}
		b.WriteString("\n")
	}
	for i := range c.embeds {
		e := c.embeds[i]
		b.WriteString("{\ncontainer := runtime.WalkPath(o.root, ")
		writePathLiteral(&b, e.path[:len(e.path)-1])
		fmt.Fprintf(&b, ")\no.%s.Init(container, %d)\n", e.fieldName, e.path[len(e.path)-1])
		b.WriteString("}\n")
	}

	b.WriteString("}\n// Insert inserts this component into the given object. This can only\n")
	b.WriteString("// be done once. The nodes will be inserted in front of `before`, or\n")
	b.WriteString("// at the end if `before` is `nil`.")
	fmt.Fprintf(&b, "\nfunc (o *%s) Insert(parent *js.Object, before *js.Object) {\n", name)
	b.WriteString("parent.Call(\"insertBefore\", o.root, before)\n")
	for i := range c.embeds {
		e := c.embeds[i]
		if e.list {
			fmt.Fprintf(&b, "o.%s.mgr.UpdateParent(o.root, parent, before)\n", e.fieldName)
		}
	}
	b.WriteString("}\n")

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

func (p *processor) dump(htmlPath, packagePath string) {
	htmlFile, err := os.Create(htmlPath)
	if err != nil {
		panic("unable to write HTML output: " + err.Error())
	}
	for i := range p.templates {
		html.Render(htmlFile, p.templates[i].strippedHTML)
	}
	htmlFile.Close()

	_, packageName := filepath.Split(packagePath)

	renderer := goRenderer{templates: p.templates, packageName: packageName,
		packagePath: packagePath}

	for name, t := range p.templates {
		renderer.writeComponentFile(name, t)
	}
}
