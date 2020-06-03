package main

import (
	"errors"

	"github.com/flyx/askew/data"
	"golang.org/x/net/html"
)

type macroDiscovery struct {
	syms *data.Symbols
}

func (md *macroDiscovery) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	name := attrVal(n.Attr, "name")
	if name == "" {
		return false, nil, errors.New(": attribute `name` missing")
	}
	pkg, _ := md.syms.Packages[md.syms.CurPkg]
	_, ok := pkg.Macros[name]
	if ok {
		return false, nil, errors.New(": duplicate name `" + name + "`")
	}
	sd := slotDiscovery{slots: make([]data.Slot, 0, 16), syms: md.syms}
	w := walker{text: allow{}, stdElements: allow{},
		embed: allow{}, include: &includeProcessor{md.syms}, slot: &sd}

	first, last, err := w.walkChildren(n, &siblings{n.FirstChild})
	if err != nil {
		return false, nil, err
	}
	pkg.Macros[name] = data.Macro{Slots: sd.slots, First: first, Last: last}
	// removes the macro and stops parent walker from descending
	return false, &html.Node{Type: html.TextNode, Data: ""}, nil
}

type slotDiscovery struct {
	syms  *data.Symbols
	slots []data.Slot
}

func (sd *slotDiscovery) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	name := attrVal(n.Attr, "name")
	if name == "" {
		return false, nil, errors.New("missing attribute `name`")
	}
	sd.slots = append(sd.slots, data.Slot{Name: name, Node: n})

	w := walker{text: allow{}, stdElements: allow{}, include: &includeProcessor{sd.syms}}
	n.FirstChild, n.LastChild, err = w.walkChildren(n, &siblings{n.FirstChild})
	return false, nil, err
}

type includeProcessor struct {
	syms *data.Symbols
}

func (ip *includeProcessor) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	name := attrVal(n.Attr, "name")
	if name == "" {
		return false, nil, errors.New(": `name` attribute missing")
	}
	m, err := ip.syms.ResolveMacro(name)
	if err != nil {
		return false, nil, errors.New(": failed to process: " + err.Error())
	}

	vm := valueMapper{slots: m.Slots, values: make([]*html.Node, len(m.Slots)),
		syms: ip.syms}
	w := walker{text: whitespaceOnly{}, stdElements: &vm}
	n.FirstChild, n.LastChild, err = w.walkChildren(n, &siblings{n.FirstChild})
	if err != nil {
		return
	}

	instantiator := macroInstantiator{
		slots: m.Slots, values: vm.values}
	instantiator.w =
		walker{text: textCopier{}, stdElements: &elmCopier{&instantiator},
			slot: &slotReplacer{&instantiator}}
	replacement, _, err = instantiator.w.walkChildren(nil, &siblings{m.First})
	return
}

type valueMapper struct {
	syms   *data.Symbols
	slots  []data.Slot
	values []*html.Node
}

func (vm *valueMapper) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var iAttrs includeChildAttribs
	extractAskewAttribs(n, &iAttrs)

	if iAttrs.slot == "" {
		return false, nil, errors.New(": child of a:include has no attribute `a:slot`")
	}
	found := false
	for i := range vm.slots {
		if vm.slots[i].Name == iAttrs.slot {
			if vm.values[i] != nil {
				return false, nil, errors.New(": dupicate value for slot `" + iAttrs.slot + "`")
			}
			vm.values[i] = n
			n.PrevSibling = nil
			n.NextSibling = nil
			n.Parent = nil
			found = true
			break
		}
	}
	if !found {
		return false, nil, errors.New(": unknown slot `" + iAttrs.slot + "`")
	}
	w := walker{text: allow{}, stdElements: allow{}, include: &includeProcessor{vm.syms}}
	replacement, _, err = w.walkChildren(n, &siblings{n.FirstChild})
	return
}

type macroInstantiator struct {
	slots  []data.Slot
	values []*html.Node
	w      walker
}

type textCopier struct{}

func (textCopier) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	replacement = &html.Node{Type: n.Type, Data: n.Data}
	return
}

type elmCopier struct {
	mi *macroInstantiator
}

func (ec *elmCopier) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	replacement = &html.Node{
		Type: n.Type, DataAtom: n.DataAtom, Data: n.Data, Namespace: n.Namespace,
		Attr: append([]html.Attribute(nil), n.Attr...)}
	replacement.FirstChild, replacement.LastChild, err = ec.mi.w.walkChildren(
		replacement, &siblings{n.FirstChild})
	return
}

type slotReplacer struct {
	mi *macroInstantiator
}

func (sl *slotReplacer) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	for i := range sl.mi.slots {
		if sl.mi.slots[i].Node != n {
			continue
		}
		if sl.mi.values[i] == nil {
			replacement, _, err = sl.mi.w.walkChildren(nil, &siblings{n.FirstChild})
		} else {
			replacement = sl.mi.values[i]
		}
		return
	}
	return false, nil, errors.New("did not find matching slot (should never happen)")
}

type componentDescender struct {
	syms *data.Symbols
}

func (cd *componentDescender) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	w := walker{text: allow{}, stdElements: allow{}, include: &includeProcessor{cd.syms},
		handler: allow{}, embed: allow{}}
	n.FirstChild, n.LastChild, err = w.walkChildren(n, &siblings{n.FirstChild})
	return false, nil, err
}

type packageDiscovery struct {
	syms *data.Symbols
}

func (pd *packageDiscovery) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var attrs packageAttribs
	collectAttribs(n, &attrs)
	_, ok := pd.syms.Packages[attrs.name]
	if ok {
		return false, nil, errors.New("duplicate package name: " + attrs.name)
	}
	pd.syms.CurPkg = attrs.name
	pd.syms.Packages[attrs.name] = &data.Package{Macros: make(map[string]data.Macro)}

	w := walker{
		text: whitespaceOnly{}, component: &componentDescender{syms: pd.syms},
		macro: &macroDiscovery{syms: pd.syms}}
	n.FirstChild, n.LastChild, err = w.walkChildren(n, &siblings{n.FirstChild})
	return false, nil, err
}

func processMacros(nodes []*html.Node, syms *data.Symbols) (err error) {
	w := walker{text: whitespaceOnly{}, aPackage: &packageDiscovery{syms: syms}}
	_, _, err = w.walkChildren(nil, &nodeSlice{nodes, 0})
	return
}
