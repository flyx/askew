package main

import (
	"golang.org/x/net/html"
)

type slot struct {
	name string
	node *html.Node
}

type macro struct {
	slots      []slot
	firstChild *html.Node
}

type macros map[string]macro

type macroProcessor struct {
	syms *symbols
}

type macroInstantiator struct {
	slots  []slot
	values []*html.Node
}

func (mi *macroInstantiator) instantiate(
	n *html.Node) (first *html.Node, last *html.Node) {
	if n.DataAtom == 0 && n.Data == "tbc:slot" {
		for i := range mi.slots {
			if mi.slots[i].node != n {
				continue
			}
			if mi.values[i] == nil {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					cFirst, cLast := mi.instantiate(c)
					if first == nil {
						first = cFirst
					} else if cFirst != nil {
						last.NextSibling = cFirst
						cFirst.PrevSibling = last
					}
					last = cLast
				}
			} else {
				first = mi.values[i]
				last = first
				first.Parent = nil
				first.NextSibling = nil
				first.PrevSibling = nil
			}
			return
		}
		panic("did not find matching slot (should never happen)")
	}

	first = &html.Node{
		Type: n.Type, DataAtom: n.DataAtom, Data: n.Data, Namespace: n.Namespace,
		Attr: append([]html.Attribute(nil), n.Attr...)}
	last = first
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		cFirst, cLast := mi.instantiate(c)
		if first.FirstChild == nil {
			first.FirstChild = cFirst
		} else if cFirst != nil {
			first.LastChild.NextSibling = cFirst
			cFirst.PrevSibling = first.LastChild
		}
		for d := cFirst; d != nil; d = d.NextSibling {
			d.Parent = first
		}
		first.LastChild = cLast
	}
	return
}

func (mp *macroProcessor) walk(n *html.Node, slots *[]slot, curPath []int,
	inSlot bool, pkgName string) (first *html.Node, last *html.Node) {
	if n.Type == html.ErrorNode {
		panic("encountered ErrorNode!")
	}
	if n.Type != html.ElementNode {
		return
	}
	if n.DataAtom == 0 {
		switch n.Data {
		case "tbc:slot":
			if slots == nil {
				panic("tbc:slot outside of tbc:macro")
			}
			if inSlot {
				panic("nested tbc:slot not allowed")
			}
			name := attrVal(n.Attr, "name")
			if name == "" {
				panic("tbc:slot misses `name` attribute")
			}
			*slots = append(*slots, slot{name: name, node: n})
			inSlot = true
		case "tbc:include":
			name := attrVal(n.Attr, "name")
			if name == "" {
				panic("tbc:slot misses `name` attribute")
			}
			m, err := mp.syms.findMacro(name)
			if err != nil {
				panic("cannot map: tbc:include: " + err.Error())
			}
			instantiator := macroInstantiator{
				slots: m.slots, values: make([]*html.Node, len(m.slots))}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type != html.ElementNode {
					continue
				}
				var iAttrs includeChildAttribs
				extractTbcAttribs(c, &iAttrs)

				if iAttrs.slot == "" {
					panic("child of tbc:include has no attribute `tbc:slot`")
				}
				found := false
				for i := range m.slots {
					if m.slots[i].name == iAttrs.slot {
						if instantiator.values[i] != nil {
							panic("dupicate value for slot `" + iAttrs.slot + "`")
						}
						instantiator.values[i] = c
						found = true
						break
					}
				}
				if !found {
					panic("unknown slot `" + iAttrs.slot + "`")
				}
			}

			for c := m.firstChild; c != nil; c = c.NextSibling {
				cFirst, cLast := instantiator.instantiate(c)
				if first == nil {
					first = cFirst
				} else if cFirst != nil {
					last.NextSibling = cFirst
					cFirst.PrevSibling = last
				}
				last = cLast
			}
			return
		case "tbc:embed", "tbc:handler":
			return n, n
		case "tbc:macro":
			panic("<tbc:macro> must be at top level")
		case "tbc:component":
			break
		default:
			panic("unknown element: <" + n.Data + ">")
		}
	}

	childPath := append(curPath, 0)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		cFirst, cLast := mp.walk(c, slots, childPath, inSlot, pkgName)
		if cFirst != nil {
			cFirst.PrevSibling = c.PrevSibling
			cLast.NextSibling = c.NextSibling
			if c.PrevSibling == nil {
				n.FirstChild = cFirst
			} else {
				c.PrevSibling.NextSibling = cFirst

			}
			if c.NextSibling == nil {
				n.LastChild = cLast
			} else {
				c.NextSibling.PrevSibling = cLast
			}
			for d := cFirst; d != nil; d = d.NextSibling {
				d.Parent = n
			}
		}
		childPath[len(childPath)-1]++
	}
	return nil, nil
}

func (mp *macroProcessor) processComponents(pkgName string, first *html.Node) {
	for n := first; n != nil; n = n.NextSibling {
		if n.Type != html.ElementNode {
			continue
		}
		if n.DataAtom == 0 && n.Data == "tbc:component" {
			mp.walk(n, nil, nil, false, pkgName)
			continue
		}
		if n.DataAtom == 0 && n.Data == "tbc:macro" {
			name := attrVal(n.Attr, "name")
			if name == "" {
				panic("tbc:macro without `name` attribute")
			}
			pkg, _ := mp.syms.packages[pkgName]
			_, ok := pkg.macros[name]
			if ok {
				panic("duplicate macro name: `" + name + "`")
			}
			slots := make([]slot, 0, 16)
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				cFirst, cLast := mp.walk(c, &slots, nil, false, pkgName)
				if cFirst != nil {
					if c.PrevSibling == nil {
						n.FirstChild = cFirst
					} else {
						c.PrevSibling.NextSibling = cFirst
					}
					if c.NextSibling == nil {
						n.LastChild = cLast
					} else {
						c.NextSibling.PrevSibling = cLast
					}
					for d := cFirst; d != nil; d = d.NextSibling {
						d.Parent = n
					}
				}
			}
			pkg.macros[name] = macro{slots: slots, firstChild: n.FirstChild}
		} else {
			panic("invalid top-level: <" + n.Data + ">")
		}
		if n.PrevSibling == nil {
			n.Parent.FirstChild = n.NextSibling
		} else {
			n.PrevSibling.NextSibling = n.NextSibling
		}
		if n.NextSibling == nil {
			n.Parent.LastChild = n.PrevSibling
		} else {
			n.NextSibling.PrevSibling = n.PrevSibling
		}
	}
}

func (mp *macroProcessor) process(nodes []*html.Node) {

	for _, n := range nodes {
		if n.Type != html.ElementNode {
			continue
		}
		if n.DataAtom != 0 || n.Data != "tbc:package" {
			panic("unexpected node <" + n.Data + "> at root level")
		}
		var attrs packageAttribs
		collectAttribs(n, &attrs)
		_, ok := mp.syms.packages[attrs.name]
		if ok {
			panic("duplicate package name: " + attrs.name)
		}
		mp.syms.curPkg = attrs.name
		mp.syms.packages[attrs.name] = &tbcPackage{macros: make(macros)}
		mp.processComponents(attrs.name, n.FirstChild)
	}
}
