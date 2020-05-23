package main

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type templateKind int

const (
	noTmpl templateKind = iota
	componentTmpl
	macroTmpl
)

type interactivity int

const (
	defaultInter interactivity = iota
	forceActive
	inactive
)

type tbcAttribs struct {
	kind        templateKind
	list        bool
	name        string
	interactive interactivity
}

func extractTbcAttribs(n *html.Node) (ret tbcAttribs) {
	i := 0
	for i < len(n.Attr) {
		attr := n.Attr[i]
		if len(attr.Key) < 4 || attr.Key[0:4] != "tbc:" {
			i++
			continue
		}
		// erase attribute from token (won't be written out)
		copy(n.Attr[i:], n.Attr[i+1:])
		n.Attr = n.Attr[:len(n.Attr)-1]

		key := attr.Key[4:]

		switch key {
		case "kind":
			if n.DataAtom != atom.Template {
				panic("tbc:kind on non-template element <" + n.Data + ">")
			}
			if ret.kind != noTmpl {
				panic("duplicate tbc:kind")
			}
			switch attr.Val {
			case "component":
				ret.kind = componentTmpl
			case "macro":
				ret.kind = macroTmpl
			default:
				panic("unknown tbc:kind: " + attr.Val)
			}
		case "list":
			if n.DataAtom == atom.Template {
				panic("tbc:list not allowed on <template>")
			}
			ret.list = true
		case "name":
			if len(ret.name) != 0 {
				panic("duplicate tbc:name: " + attr.Val)
			}
			ret.name = attr.Val
		case "ignore":
			if n.DataAtom == atom.Template {
				panic("tbc:ignore invalid on <template>")
			}
			ret.interactive = inactive
		case "dynamic":
			if n.DataAtom == atom.Template {
				panic("tbc:dynamic invalid on <template>")
			}
			ret.interactive = forceActive
		default:
			panic("unknown attribute: tbc:" + key)
		}
	}
	return
}
