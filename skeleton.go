package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"

	"github.com/flyx/askew/components"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers"
	"github.com/flyx/askew/walker"
	"golang.org/x/net/html"
)

type templateInjector struct {
	syms *data.Symbols
	seen bool
}

func (ti *templateInjector) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if ti.seen {
		return false, nil, errors.New("duplicate <a:templates> in document")
	}
	ti.seen = true
	var last *html.Node
	for _, pkg := range ti.syms.Packages {
		for _, file := range pkg.Files {
			for _, cmp := range file.Components {
				if last == nil {
					last = cmp.Template
					last.PrevSibling = nil
					replacement = last
				} else {
					last.NextSibling = cmp.Template
					cmp.Template.PrevSibling = last
					last = last.NextSibling
				}
				last.NextSibling = nil
			}
		}
	}
	return
}

type importHandler struct {
	syms     *data.Symbols
	skeleton *data.Skeleton
}

func (ih *importHandler) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var raw string
	if n.FirstChild != nil {
		if n.LastChild != n.FirstChild || n.FirstChild.Type != html.TextNode {
			return false, nil, errors.New(": may only contain text content")
		}
		raw = n.FirstChild.Data
	}
	imports, err := parsers.ParseImports(raw)
	if err != nil {
		return false, nil, errors.New(": " + err.Error())
	}
	if ih.skeleton.Imports != nil {
		return false, nil, errors.New(": cannot have more than one <a:import> per file")
	}
	ih.skeleton.Imports = imports
	ih.syms.CurFile.Imports = imports
	return false, &html.Node{Type: html.CommentNode, Data: "import"}, nil
}

func readSkeleton(syms *data.Symbols, path string) (*data.Skeleton, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	os.Stdout.WriteString("[info] processing skeleton file " + path + "\n")
	root, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil, errors.New(path + ": " + err.Error())
	}
	if root.Type != html.DocumentNode {
		return nil, errors.New(path + ": HTML document is not a DocumentNode")
	}

	indexList := make([]int, 0, 32)

	s := &data.Skeleton{EmbedHost: data.EmbedHost{}, Root: root}

	syms.CurHost = &s.EmbedHost
	syms.CurPkg = ""
	syms.CurFile = &data.File{}

	w := walker.Walker{TextNode: walker.Allow{}, Templates: &templateInjector{syms, false},
		StdElements: walker.Allow{}, Import: &importHandler{syms, s},
		Embed: components.NewEmbedProcessor(syms, &indexList), IndexList: &indexList}
	root.FirstChild, root.LastChild, err = w.WalkChildren(root, &walker.Siblings{Cur: root.FirstChild})
	if err != nil {
		return nil, errors.New(path + ": " + err.Error())
	}

	tmp := make([]data.Embed, len(s.Embeds))
	for i, e := range s.Embeds {
		tmp[len(tmp)-i-1] = e
	}
	s.Embeds = tmp

	return s, nil
}
