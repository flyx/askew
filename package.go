package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (p *tbcPackage) process(syms *symbols, n *html.Node, counter *int) {
	p.components = make(map[string]*component)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			text := strings.TrimSpace(c.Data)
			if len(text) > 0 {
				panic("non-whitespace text at top level: `" + text + "`")
			}
		case html.ErrorNode:
			panic("encountered ErrorNode: " + c.Data)
		case html.ElementNode:
			if c.DataAtom != 0 || c.Data != "tbc:component" {
				panic("only tbc:macro and tbc:component are allowed at top level. found <" + c.Data + ">")
			}

			var cmpAttrs componentAttribs
			collectAttribs(c, &cmpAttrs)
			if len(cmpAttrs.name) == 0 {
				panic("<tbc:component> must have name!")
			}
			c.DataAtom = atom.Template
			c.Data = "template"

			cmp := &component{processedHTML: c, needsController: cmpAttrs.controller,
				dependencies: make(map[string]struct{})}
			syms.curComponent = cmp
			*counter++
			cmp.id = fmt.Sprintf("tbc-component-%d-%s", *counter, strings.ToLower(cmpAttrs.name))
			c.Attr = []html.Attribute{html.Attribute{Key: "id", Val: cmp.id}}
			indexList := make([]int, 1, 32)
			for child := c.FirstChild; child != nil; child = child.NextSibling {
				cmp.walk(syms, child, indexList)
				indexList[0]++
			}
			p.components[cmpAttrs.name] = cmp
		default:
			panic("illegal node at top level: " + c.Data)
		}
	}
}
