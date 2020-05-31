package main

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// nodeHandler is a handler that processes a specific kind of node.
type nodeHandler interface {
	// processes the node. if descend is true, the walker will descend into the
	// node's children afterwards.
	//
	// if replacement is not nil, the node will be replaced by the last of nodes,
	// starting at replacement and continuing through the NextSibling nodes, in
	// its current context (usually the node list of its current parent).
	//
	// replacement may only be not nil if descend if false.
	process(n *html.Node) (descend bool, replacement *html.Node, err error)
}

// walker walks over a node graph. for each tbc-specific element, a nodeHandler
// must be registered for that element to be valid. for each standard HTML
// element, a nodeHandler may be registered.
type walker struct {
	tbcPackage  nodeHandler
	macro       nodeHandler
	component   nodeHandler
	slot        nodeHandler
	include     nodeHandler
	embed       nodeHandler
	handler     nodeHandler
	templates   nodeHandler
	stdElements nodeHandler
	text        nodeHandler
}

func (w *walker) walk(n *html.Node, nodesCount map[string]int) (replacement *html.Node, err error) {
	switch n.Type {
	case html.ErrorNode:
		return nil, errors.New(": encountered error node: " + n.Data)
	case html.TextNode:
		if w.text == nil {
			return nil, errors.New(": text content not allowed here")
		}
		_, replacement, err = w.text.process(n)
		return
	case html.ElementNode:
		break
	case html.CommentNode:
		return nil, nil
	default:
		return nil, errors.New(": unexpected node kind")
	}
	count, _ := nodesCount[n.Data]
	count++
	nodesCount[n.Data] = count
	replacement, err = w.processElement(n)
	if err != nil {
		return nil, fmt.Errorf("/%s[%d]%s", n.Data, count, err.Error())
	}
	return
}

// processElement walks over the subtree with node n, which must be an html.ElementNode.
// each time a child node is encountered for which a nodeHandler is available,
// that nodeHandler's process() func is called.
func (w *walker) processElement(n *html.Node) (replacement *html.Node, err error) {
	var h nodeHandler
	if n.DataAtom == 0 {
		switch n.Data {
		case "tbc:package":
			h = w.tbcPackage
		case "tbc:macro":
			h = w.macro
		case "tbc:component":
			h = w.component
		case "tbc:slot":
			h = w.slot
		case "tbc:include":
			h = w.include
		case "tbc:embed":
			h = w.embed
		case "tbc:handler":
			h = w.handler
		case "tbc:templates":
			h = w.templates
		default:
			return nil, errors.New(": unknown element")
		}
		if h == nil {
			return nil, errors.New(": element not allowed here")
		}
	} else {
		h = w.stdElements
	}
	if h != nil {
		var descend bool
		descend, replacement, err = h.process(n)
		if err != nil || !descend {
			return
		}
	}
	n.FirstChild, n.LastChild, err = w.walkChildren(n, &siblings{n.FirstChild})
	return
}

type nodeList interface {
	next() *html.Node
}

type siblings struct {
	cur *html.Node
}

func (s *siblings) next() (ret *html.Node) {
	ret = s.cur
	if ret != nil {
		s.cur = s.cur.NextSibling
	}
	return
}

type nodeSlice struct {
	items []*html.Node
	cur   int
}

func (ns *nodeSlice) next() (ret *html.Node) {
	if ns.cur == len(ns.items) {
		return nil
	}
	ret = ns.items[ns.cur]
	ns.cur++
	return
}

func (w *walker) walkChildren(parent *html.Node,
	l nodeList) (repFirst, repLast *html.Node, err error) {
	nodesCount := make(map[string]int)
	for c := l.next(); c != nil; c = l.next() {
		f, err := w.walk(c, nodesCount)
		if err != nil {
			return nil, nil, err
		}
		if repFirst == nil {
			if f == nil {
				repFirst = c
			} else {
				repFirst = f
			}
			if parent != nil {
				parent.FirstChild = repLast
			}
			repLast = repFirst
		} else {
			if f == nil {
				repLast.NextSibling = c
				c.PrevSibling = repLast
			} else {
				repLast.NextSibling = f
				f.PrevSibling = repLast
			}
			repLast = repLast.NextSibling
		}

		if f != nil && f != c {
			seen := make(map[*html.Node]struct{})
			for ; ; repLast = repLast.NextSibling {
				_, ok := seen[repLast]
				if ok {
					return nil, nil, errors.New(": cycle in siblings")
				}
				seen[repLast] = struct{}{}
				repLast.Parent = parent
				if repLast.NextSibling == nil {
					break
				}
			}
		}
	}
	if parent != nil {
		parent.LastChild = repLast
	}
	return
}

// allow can be used to signal that certain special elements are allowed to be
// encountered by the walker, but won't be processed.
type allow struct{}

func (allow) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	return true, nil, nil
}

// dontDescend can be used to signal that certain elements are allowed and
// should not be descended into.
type dontDescend struct{}

func (dontDescend) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	return false, nil, nil
}

// whitespaceOnly can be used to force TextNodes to be empty.
type whitespaceOnly struct{}

func (whitespaceOnly) process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if strings.TrimSpace(n.Data) != "" {
		err = errors.New(": contains illegal text content")
	}
	return
}
