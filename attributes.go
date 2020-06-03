package main

import (
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers"
	"golang.org/x/net/html"
)

type interactivity int

const (
	defaultInter interactivity = iota
	forceActive
	inactive
)

type attribCollector interface {
	collect(name string, val string) bool
}

type packageAttribs struct {
	name string
}

func (p *packageAttribs) collect(name, val string) bool {
	if name == "name" {
		p.name = val
		return true
	}
	return false
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
	bindings []data.VariableMapping
	capture  []data.EventMapping
}

func (g *generalAttribs) collect(name, val string) bool {
	switch name {
	case "ignore":
		g.ignore = true
	case "bindings":
		var err error
		g.bindings, err = parsers.ParseBindings(val)
		if err != nil {
			panic("while parsing bindings `" + val + "`: " + err.Error())
		}
	case "capture":
		var err error
		g.capture, err = parsers.ParseCapture(val)
		if err != nil {
			panic("while parsing capture `" + val + "`: " + err.Error())
		}
	default:
		return false
	}
	return true
}

type askewAttribs struct {
	list        bool
	name        string
	interactive interactivity
}

func extractAskewAttribs(n *html.Node, target attribCollector) {
	seen := make(map[string]struct{})

	i := 0
	for i < len(n.Attr) {
		attr := n.Attr[i]
		if len(attr.Key) < 2 || attr.Key[0:2] != "a:" {
			i++
			continue
		}

		// erase attribute from token (won't be written out)
		copy(n.Attr[i:], n.Attr[i+1:])
		n.Attr = n.Attr[:len(n.Attr)-1]

		key := attr.Key[2:]

		if _, ok := seen[key]; ok {
			panic("duplicate attribute: " + attr.Key)
		}
		seen[key] = struct{}{}
		if !target.collect(key, attr.Val) {
			panic("element <" + n.Data + "> does not allow attribute " + attr.Key)
		}
	}
}

func collectAttribs(n *html.Node, target attribCollector) {
	seen := make(map[string]struct{})

	for _, attr := range n.Attr {

		if _, ok := seen[attr.Key]; ok {
			panic("duplicate attribute: " + attr.Key)
		}
		seen[attr.Key] = struct{}{}
		if !target.collect(attr.Key, attr.Val) {
			panic("element <" + n.Data + "> does not allow attribute " + attr.Key)
		}
	}
}

func attrVal(a []html.Attribute, name string) string {
	for i := range a {
		if a[i].Key == name {
			return a[i].Val
		}
	}
	return ""
}

func attrExists(a []html.Attribute, name string) bool {
	for i := range a {
		if a[i].Key == name {
			return true
		}
	}
	return false
}
