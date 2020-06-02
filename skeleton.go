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
	for _, pkg := range ti.syms.Packages {
		for _, cmp := range pkg.Components {
			if replacement == nil {
				replacement = cmp.Template
			} else {
				replacement.NextSibling = cmp.Template
				cmp.Template.PrevSibling = replacement
				replacement = replacement.NextSibling
			}
		}
	}
	return
}

type globalEmbedProcessor struct {
	syms      *data.Symbols
	embeds    []data.Embed
	indexList []int
}

func (gep *globalEmbedProcessor) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	e, err := resolveEmbed(n, gep.syms, gep.indexList)
	if err != nil {
		return false, nil, err
	}
	gep.embeds = append(gep.embeds, e)
	return false, nil, nil
}

func readSkeleton(syms *data.Symbols, file string) (*data.Skeleton, error) {
	raw, err := ioutil.ReadFile(file)
	if err == nil {
		return nil, err
	}
	root, err := html.Parse(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if root.Type != html.DocumentNode {
		panic(file + ": HTML document is not a DocumentNode")
	}

	gep := globalEmbedProcessor{syms, make([]data.Embed, 0, 32), make([]int, 0, 32)}

	w := walker{templates: &templateInjector{syms, false}, stdElements: allow{},
		embed: &gep, indexList: &gep.indexList}
	_, _, err = w.walkChildren(root, &siblings{root.FirstChild})
	if err != nil {
		return nil, err
	}
	return &data.Skeleton{Root: root, Embeds: gep.embeds}, nil
}
