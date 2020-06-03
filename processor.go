package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/output"
	"golang.org/x/net/html/atom"

	"golang.org/x/mod/modfile"
	"golang.org/x/net/html"
)

type processor struct {
	syms    data.Symbols
	counter int
	mod     *modfile.File
}

// dummy body node to be used for fragment parsing
var bodyEnv = html.Node{
	Type:     html.ElementNode,
	Data:     "body",
	DataAtom: atom.Body}

func (p *processor) init(outputPath string) bool {
	p.syms.Packages = make(map[string]*data.Package)
	raw, err := ioutil.ReadFile("go.mod")
	if err != nil {
		if os.IsNotExist(err) {
			os.Stderr.WriteString("[error] did not find go.mod.\n")
			os.Stderr.WriteString(
				"[error] askew must be run in the root directory of your module.\n")
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
	p.syms.PkgBasePath = filepath.Join(p.mod.Module.Mod.Path, outputPath)
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
				panic(file + ": non-whitespace text at top level: `" + text + "`")
			}
		case html.ErrorNode:
			panic(file + ": encountered ErrorNode: " + n.Data)
		case html.ElementNode:
			if n.DataAtom != 0 || n.Data != "a:package" {
				panic(file + ": only a:package is allowed at top level. found <" + n.Data + ">")
			}
			p.syms.CurPkg = attrVal(n.Attr, "name")
			if err := processComponents(&p.syms, n, &p.counter); err != nil {
				panic(file + err.Error())
			}
		default:
			panic(file + ": illegal node at top level: " + n.Data)
		}
	}
}

func (p *processor) dump(skeleton *data.Skeleton, htmlPath, packageParent string) {
	htmlFile, err := os.Create(htmlPath)
	if err != nil {
		panic("unable to write HTML output: " + err.Error())
	}
	if skeleton == nil {
		for _, pkg := range p.syms.Packages {
			for _, c := range pkg.Components {
				html.Render(htmlFile, c.Template)
			}
		}
	} else {
		html.Render(htmlFile, skeleton.Root)
	}

	htmlFile.Close()

	for pkgName, pkg := range p.syms.Packages {
		w := output.PackageWriter{Syms: &p.syms, PackageName: pkgName,
			PackagePath: filepath.Join(packageParent, pkgName)}
		if err := os.MkdirAll(w.PackagePath, os.ModePerm); err != nil {
			panic("failed to create package directory '" + w.PackagePath +
				"': " + err.Error())
		}
		for name, t := range pkg.Components {
			w.WriteComponent(name, t)
		}
	}

	if skeleton != nil {
		output.WriteSkeleton(&p.syms, filepath.Join(packageParent, "init.go"), skeleton)
	}
}
