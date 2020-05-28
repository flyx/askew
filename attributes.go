package main

import (
	"golang.org/x/net/html"
)

type interactivity int

const (
	defaultInter interactivity = iota
	forceActive
	inactive
)

type tbcAttribCollector interface {
	collect(name string, val string) bool
}

type componentAttribs struct {
	name       string
	controller bool
}

func (t *componentAttribs) collect(name, val string) bool {
	switch name {
	case "name":
		t.name = val
		return true
	case "controller":
		t.controller = true
		return true
	}
	return false
}

type includeChildAttribs struct {
	slot string
}

func (i *includeChildAttribs) collect(name, val string) bool {
	if name == "slot" {
		i.slot = val
		return true
	}
	return false
}

type embedAttribs struct {
	list bool
}

func (e *embedAttribs) collect(name, val string) bool {
	if name == "list" {
		e.list = true
		return true
	}
	return false
}

type generalAttribs struct {
	ignore   bool
	bindings []varBinding
	captures []capture
}

func (g *generalAttribs) collect(name, val string) bool {
	switch name {
	case "ignore":
		g.ignore = true
	case "bindings":
		var err error
		g.bindings, err = bp.parse(val)
		if err != nil {
			panic("while parsing bindings `" + val + "`: " + err.Error())
		}
	case "capture":
		var err error
		g.captures, err = cp.parse(val)
		if err != nil {
			panic("while parsing capture `" + val + "`: " + err.Error())
		}
	default:
		return false
	}
	return true
}

type tbcAttribs struct {
	list        bool
	name        string
	interactive interactivity
}

func extractTbcAttribs(n *html.Node, target tbcAttribCollector) {
	seen := make(map[string]struct{})

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

		if _, ok := seen[key]; ok {
			panic("duplicate attribute: " + attr.Key)
		}
		seen[key] = struct{}{}
		if !target.collect(key, attr.Val) {
			panic("element <" + n.Data + "> does not allow attribute " + attr.Key)
		}
	}
	return
}
