package main

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type slot struct {
	name string
	node *html.Node
}

type macro struct {
	slots      []slot
	firstChild *html.Node
}

type includesProcessor struct {
	macros map[string]macro
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

func (ip *includesProcessor) walk(n *html.Node, slots *[]slot, curPath []int,
	inSlot bool) (first *html.Node, last *html.Node) {
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
			m, ok := ip.macros[name]
			if !ok {
				panic("tbc:include references unknown macro `" + name + "`")
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
			break
		case "tbc:macro":
			panic("<tbc:macro> must be at top level")
		default:
			panic("unknown element: <" + n.Data + ">")
		}
	}

	childPath := append(curPath, 0)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		cFirst, cLast := ip.walk(c, slots, childPath, inSlot)
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
		childPath[len(childPath)-1]++
	}
	return nil, nil
}

func (ip *includesProcessor) process(nodes *[]*html.Node) {
	ip.macros = make(map[string]macro)

	for i := 0; i < len(*nodes); {
		n := (*nodes)[i]
		if n.Type != html.ElementNode {
			i++
			continue
		}
		if n.DataAtom == atom.Template {
			ip.walk(n, nil, nil, false)
			i++
			continue
		}
		if n.DataAtom == 0 && n.Data == "tbc:macro" {
			name := attrVal(n.Attr, "name")
			if name == "" {
				panic("tbc:macro without `name` attribute")
			}
			_, ok := ip.macros[name]
			if ok {
				panic("duplicate macro name: `" + name + "`")
			}
			slots := make([]slot, 0, 16)
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				cFirst, cLast := ip.walk(c, &slots, nil, false)
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
			ip.macros[name] = macro{slots: slots, firstChild: n.FirstChild}
		} else {
			panic("invalid top-level: <" + n.Data + ">")
		}
		copy((*nodes)[i:], (*nodes)[i+1:])
		*nodes = (*nodes)[:len(*nodes)-1]
	}
}
