package main

import (
	"bytes"
	"errors"
	"io/ioutil"

	"github.com/flyx/tbc/data"
	"golang.org/x/net/html"
)

type templateInjector struct {
	syms *data.Symbols
	seen bool
}

func (ti *templateInjector) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if ti.seen {
		return false, nil, errors.New("duplicate <tbc:templates> in document")
	}
	ti.seen = true
	var last *html.Node
	for _, pkg := range ti.syms.Packages {
		for _, cmp := range pkg.Components {
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
	return
}

func readSkeleton(syms *data.Symbols, file string) (*data.Skeleton, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	root, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil, errors.New(file + ": " + err.Error())
	}
	if root.Type != html.DocumentNode {
		panic(file + ": HTML document is not a DocumentNode")
	}

	indexList := make([]int, 0, 32)

	s := &data.Skeleton{EmbedHost: data.EmbedHost{Dependencies: make(map[string]struct{})}, Root: root}

	syms.CurHost = &s.EmbedHost
	syms.CurPkg = ""

	w := walker{text: allow{}, templates: &templateInjector{syms, false}, stdElements: allow{},
		embed: &embedProcessor{syms, &indexList}, indexList: &indexList}
	root.FirstChild, root.LastChild, err = w.walkChildren(root, &siblings{root.FirstChild})
	if err != nil {
		return nil, errors.New(file + err.Error())
	}

	tmp := make([]data.Embed, len(s.Embeds))
	for i, e := range s.Embeds {
		tmp[len(tmp)-i-1] = e
	}
	s.Embeds = tmp

	return s, nil
}
