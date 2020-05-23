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

type tdcAttribs struct {
	kind        templateKind
	list        bool
	name        string
	interactive interactivity
}

func extractTdcAttribs(n *html.Node) (ret tdcAttribs) {
	i := 0
	for i < len(n.Attr) {
		attr := n.Attr[i]
		if attr.Namespace != "tdc" {
			i++
			continue
		}
		// erase attribute from token (won't be written out)
		copy(n.Attr[i:], n.Attr[i+1:])
		n.Attr = n.Attr[:len(n.Attr)-1]

		switch attr.Key {
		case "kind":
			if n.DataAtom != atom.Template {
				panic("tdc:kind on non-template element <" + n.Data + ">")
			}
			if ret.kind != noTmpl {
				panic("duplicate tdc:kind")
			}
			switch attr.Val {
			case "component":
				ret.kind = componentTmpl
			case "macro":
				ret.kind = macroTmpl
			default:
				panic("unknown tdc:kind: " + attr.Val)
			}
		case "list":
			if n.DataAtom == atom.Template {
				panic("tdc:list not allowed on <template>")
			}
			ret.list = true
		case "name":
			if len(ret.name) != 0 {
				panic("duplicate tdc:name: " + attr.Val)
			}
			ret.name = attr.Val
		case "ignore":
			if n.DataAtom == atom.Template {
				panic("tdc:ignore invalid on <template>")
			}
			ret.interactive = inactive
		case "dynamic":
			if n.DataAtom == atom.Template {
				panic("tdc:dynamic invalid on <template>")
			}
			ret.interactive = forceActive
		default:
			panic("unknown attribute: tdc:" + attr.Key)
		}
	}
	return
}
