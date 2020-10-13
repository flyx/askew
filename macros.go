package main

import (
	"errors"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/walker"
	"golang.org/x/net/html"
)

type macroDiscovery struct {
	syms *data.Symbols
}

func (md *macroDiscovery) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	name := attributes.Val(n.Attr, "name")
	if name == "" {
		return false, nil, errors.New(": attribute `name` missing")
	}
	pkg, _ := md.syms.Packages[md.syms.CurPkg]
	for _, file := range pkg.Files {
		_, ok := file.Macros[name]
		if ok {
			return false, nil, errors.New(": duplicate name `" + name + "`")
		}
	}
	sd := slotDiscovery{slots: make([]data.Slot, 0, 16), syms: md.syms}
	w := walker.Walker{TextNode: walker.Allow{}, StdElements: walker.Allow{},
		Text: walker.Allow{}, Embed: walker.Allow{}, Construct: walker.Allow{},
		Include: &includeProcessor{md.syms}, Slot: &sd}

	first, last, err := w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})
	if err != nil {
		return false, nil, err
	}
	curFile := md.syms.CurAskewFile()
	if curFile.Macros == nil {
		curFile.Macros = make(map[string]data.Macro)
	}
	curFile.Macros[name] = data.Macro{Slots: sd.slots, First: first, Last: last}
	// removes the macro and stops parent walker from descending
	return false, &html.Node{Type: html.TextNode, Data: ""}, nil
}

type slotDiscovery struct {
	syms  *data.Symbols
	slots []data.Slot
}

func (sd *slotDiscovery) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	name := attributes.Val(n.Attr, "name")
	if name == "" {
		return false, nil, errors.New("missing attribute `name`")
	}
	sd.slots = append(sd.slots, data.Slot{Name: name, Node: n})

	w := walker.Walker{TextNode: walker.Allow{}, StdElements: walker.Allow{},
		Text: walker.Allow{}, Include: &includeProcessor{sd.syms}}
	n.FirstChild, n.LastChild, err = w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})
	return false, nil, err
}

type includeProcessor struct {
	syms *data.Symbols
}

func (ip *includeProcessor) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	name := attributes.Val(n.Attr, "name")
	if name == "" {
		return false, nil, errors.New(": `name` attribute missing")
	}
	m, err := ip.syms.ResolveMacro(name)
	if err != nil {
		return false, nil, errors.New(": failed to process: " + err.Error())
	}

	vm := valueMapper{slots: m.Slots, values: make([]*html.Node, len(m.Slots)),
		syms: ip.syms}
	w := walker.Walker{TextNode: walker.WhitespaceOnly{}, StdElements: &vm}
	n.FirstChild, n.LastChild, err = w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})
	if err != nil {
		return
	}

	instantiator := macroInstantiator{
		slots: m.Slots, values: vm.values}
	ec := elmCopier{&instantiator}
	instantiator.w =
		walker.Walker{TextNode: textCopier{}, StdElements: &ec, Text: &ec,
			Slot: &slotReplacer{&instantiator}}
	replacement, _, err = instantiator.w.WalkChildren(nil, &walker.Siblings{Cur: m.First})
	return
}

type valueMapper struct {
	syms   *data.Symbols
	slots  []data.Slot
	values []*html.Node
}

func (vm *valueMapper) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	var iAttrs attributes.IncludeChild
	attributes.ExtractAskewAttribs(n, &iAttrs)

	if iAttrs.Slot == "" {
		return false, nil, errors.New(": child of a:include has no attribute `a:slot`")
	}
	found := false
	for i := range vm.slots {
		if vm.slots[i].Name == iAttrs.Slot {
			if vm.values[i] != nil {
				return false, nil, errors.New(": dupicate value for slot `" + iAttrs.Slot + "`")
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
		return false, nil, errors.New(": unknown slot `" + iAttrs.Slot + "`")
	}
	w := walker.Walker{TextNode: walker.Allow{}, StdElements: walker.Allow{},
		Text: walker.Allow{}, Embed: walker.Allow{}, Construct: walker.Allow{},
		Include: &includeProcessor{vm.syms}}
	replacement, _, err = w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})
	return
}

type macroInstantiator struct {
	slots  []data.Slot
	values []*html.Node
	w      walker.Walker
}

type textCopier struct{}

func (textCopier) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	replacement = &html.Node{Type: n.Type, Data: n.Data}
	return
}

type elmCopier struct {
	mi *macroInstantiator
}

func (ec *elmCopier) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	replacement = &html.Node{
		Type: n.Type, DataAtom: n.DataAtom, Data: n.Data, Namespace: n.Namespace,
		Attr: append([]html.Attribute(nil), n.Attr...)}
	replacement.FirstChild, replacement.LastChild, err = ec.mi.w.WalkChildren(
		replacement, &walker.Siblings{Cur: n.FirstChild})
	return
}

type slotReplacer struct {
	mi *macroInstantiator
}

func (sl *slotReplacer) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	for i := range sl.mi.slots {
		if sl.mi.slots[i].Node != n {
			continue
		}
		if sl.mi.values[i] == nil {
			replacement, _, err = sl.mi.w.WalkChildren(nil, &walker.Siblings{Cur: n.FirstChild})
			if replacement == nil {
				replacement = &html.Node{Type: html.TextNode}
			}
		} else {
			replacement = sl.mi.values[i]
		}
		return
	}
	return false, nil, errors.New("did not find matching slot (should never happen)")
}

type unitDescender struct {
	syms *data.Symbols
}

func (cd *unitDescender) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	w := walker.Walker{TextNode: walker.Allow{}, StdElements: walker.Allow{}, Include: &includeProcessor{cd.syms},
		Handlers: walker.Allow{}, Controller: walker.Allow{}, Data: walker.Allow{},
		Embed: walker.Allow{}, Construct: walker.Allow{}, Text: walker.Allow{}}
	n.FirstChild, n.LastChild, err = w.WalkChildren(n, &walker.Siblings{Cur: n.FirstChild})
	return false, nil, err
}

type importRemover struct{}

func (importRemover) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	return false, &html.Node{Type: html.TextNode}, nil
}

func processMacros(nodes []*html.Node, syms *data.Symbols) (dummyParent *html.Node, err error) {
	dummyParent = &html.Node{Type: html.ElementNode}
	if len(nodes) > 0 {
		dummyParent.FirstChild = nodes[0]
		dummyParent.LastChild = nodes[len(nodes)-1]
		w := walker.Walker{TextNode: walker.WhitespaceOnly{},
			Component: &unitDescender{syms: syms},
			Site:      &unitDescender{syms: syms},
			Macro:     &macroDiscovery{syms: syms},
			Import:    importRemover{}}
		_, _, err = w.WalkChildren(dummyParent, &walker.NodeSlice{Items: nodes})
	}
	return
}
