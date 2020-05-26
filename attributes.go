package main

import (
	"strings"

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

type templateAttribs struct {
	name string
}

func (t *templateAttribs) collect(name, val string) bool {
	if name == "name" {
		t.name = val
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
	interactive interactivity
	name        string
	classSwitch string
	capture     map[string]string
}

func (g *generalAttribs) collect(name, val string) bool {
	switch name {
	case "dynamic":
		if g.interactive != defaultInter {
			panic("cannot mix tbc:dynamic with dbc:ignore")
		}
		g.interactive = forceActive
	case "ignore":
		if g.interactive != defaultInter {
			panic("cannot mix tbc:dynamic with dbc:ignore")
		}
		g.interactive = inactive
	case "name":
		g.name = val
	case "classswitch":
		g.classSwitch = val
	case "capture":
		items := strings.Split(strings.TrimSpace(val), ",")
		g.capture = make(map[string]string)
		for i := range items {
			item := strings.TrimSpace(items[i])
			tmp := strings.Split(item, "=")
			if len(tmp) != 2 {
				panic("illegal capture item: " + item)
			}
			event := strings.TrimSpace(tmp[0])
			_, ok := g.capture[event]
			if ok {
				panic("duplicate event in capture: " + event)
			}
			g.capture[event] = strings.TrimSpace(tmp[1])
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
