package main

import (
	"errors"

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

func invalidAttribute(name string) error {
	return errors.New("element does not allow attribute `" + name + "`")
}

type attribCollector interface {
	collect(name string, val string) error
}

type packageAttribs struct {
	name string
}

func (p *packageAttribs) collect(name, val string) error {
	if name == "name" {
		p.name = val
		return nil
	}
	return invalidAttribute(name)
}

type componentAttribs struct {
	name       string
	controller bool
	params     []data.ComponentParam
}

func (t *componentAttribs) collect(name, val string) error {
	switch name {
	case "name":
		t.name = val
		return nil
	case "controller":
		t.controller = true
		return nil
	case "params":
		var err error
		t.params, err = parsers.ParseParameters(val)
		return err
	}
	return invalidAttribute(name)
}

type includeChildAttribs struct {
	slot string
}

func (i *includeChildAttribs) collect(name, val string) error {
	if name == "slot" {
		i.slot = val
		return nil
	}
	return invalidAttribute(name)
}

type embedAttribs struct {
	list, optional bool
	t, name        string
	args           data.Arguments
}

func (e *embedAttribs) collect(name, val string) error {
	switch name {
	case "list":
		e.list = true
		return nil
	case "optional":
		e.optional = true
		return nil
	case "type":
		e.t = val
		return nil
	case "name":
		e.name = val
		return nil
	case "args":
		var err error
		e.args, err = parsers.AnalyseArguments(val)
		return err
	}
	return invalidAttribute(name)
}

type generalAttribs struct {
	bindings  []data.VariableMapping
	capture   []data.EventMapping
	_if, _for *data.ControlBlock
	assign    []data.Assignment
}

func (g *generalAttribs) collect(name, val string) error {
	switch name {
	case "bindings":
		var err error
		g.bindings, err = parsers.ParseBindings(val)
		if err != nil {
			return errors.New(": invalid bindings: " + err.Error())
		}
	case "capture":
		var err error
		g.capture, err = parsers.ParseCapture(val)
		if err != nil {
			return errors.New(": invalid capture: " + err.Error())
		}
	case "if":
		g._if = &data.ControlBlock{Kind: data.IfBlock, Expression: val}
	case "for":
		var err error
		g._for, err = parsers.ParseFor(val)
		if err != nil {
			return errors.New(": invalid for: " + err.Error())
		}
	case "assign":
		var err error
		g.assign, err = parsers.ParseAssignments(val)
		if err != nil {
			return errors.New(": invalid assign: " + err.Error())
		}
	default:
		return invalidAttribute(name)
	}
	return nil
}

type askewAttribs struct {
	list        bool
	name        string
	interactive interactivity
}

func extractAskewAttribs(n *html.Node, target attribCollector) error {
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
		if err := target.collect(key, attr.Val); err != nil {
			return err
		}
	}
	return nil
}

func collectAttribs(n *html.Node, target attribCollector) error {
	seen := make(map[string]struct{})

	for _, attr := range n.Attr {

		if _, ok := seen[attr.Key]; ok {
			panic("duplicate attribute: " + attr.Key)
		}
		seen[attr.Key] = struct{}{}
		if err := target.collect(attr.Key, attr.Val); err != nil {
			return err
		}
	}
	return nil
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
