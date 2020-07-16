package walker

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// NodeHandler is a handler that processes a specific kind of node.
type NodeHandler interface {
	// processes the node. if descend is true, the walker will descend into the
	// node's children afterwards.
	//
	// if replacement is not nil, the node will be replaced by the last of nodes,
	// starting at replacement and continuing through the NextSibling nodes, in
	// its current context (usually the node list of its current parent).
	//
	// replacement may only be not nil if descend if false.
	Process(n *html.Node) (descend bool, replacement *html.Node, err error)
}

// Walker walks over a node graph. for each askew-specific element, a
// NodeHandler must be registered for that element to be valid. for each
// standard HTML element, a NodeHandler may be registered.
type Walker struct {
	AImport     NodeHandler
	Macro       NodeHandler
	Component   NodeHandler
	Slot        NodeHandler
	Include     NodeHandler
	Embed       NodeHandler
	Handlers    NodeHandler
	Templates   NodeHandler
	StdElements NodeHandler
	Text        NodeHandler
	AText       NodeHandler
	IndexList   *[]int
}

func (w *Walker) walk(n *html.Node, nodesCount map[string]int) (replacement *html.Node, err error) {
	switch n.Type {
	case html.ErrorNode:
		return nil, errors.New(": encountered error node: " + n.Data)
	case html.TextNode:
		if w.Text == nil {
			return nil, errors.New(": text content not allowed here")
		}
		_, replacement, err = w.Text.Process(n)
		return
	case html.ElementNode:
		break
	case html.CommentNode, html.DoctypeNode:
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
func (w *Walker) processElement(n *html.Node) (replacement *html.Node, err error) {
	var h NodeHandler
	if n.DataAtom == 0 {
		switch n.Data {
		case "a:import":
			h = w.AImport
		case "a:macro":
			h = w.Macro
		case "a:component":
			h = w.Component
		case "a:slot":
			h = w.Slot
		case "a:include":
			h = w.Include
		case "a:embed":
			h = w.Embed
		case "a:handlers":
			h = w.Handlers
		case "a:templates":
			h = w.Templates
		case "a:text":
			h = w.AText
		default:
			return nil, errors.New(": unknown element")
		}
		if h == nil {
			return nil, errors.New(": element not allowed here")
		}
	} else {
		h = w.StdElements
	}
	if h != nil {
		var descend bool
		descend, replacement, err = h.Process(n)
		if err != nil || !descend {
			return
		}
	}
	n.FirstChild, n.LastChild, err = w.WalkChildren(n, &Siblings{n.FirstChild})
	return
}

// NodeList is a list of node that can be iterated over
type NodeList interface {
	next() *html.Node
}

// Siblings is a NodeList of sibling nodes
type Siblings struct {
	Cur *html.Node
}

func (s *Siblings) next() (ret *html.Node) {
	ret = s.Cur
	if ret != nil {
		s.Cur = s.Cur.NextSibling
	}
	return
}

// NodeSlice is a NodeList backed by a slice.
type NodeSlice struct {
	Items []*html.Node
	cur   int
}

func (ns *NodeSlice) next() (ret *html.Node) {
	if ns.cur == len(ns.Items) {
		return nil
	}
	ret = ns.Items[ns.cur]
	ns.cur++
	return
}

// WalkChildren walks over each node of the given list, treating them as
// children of the given parent (which may be nil)
func (w *Walker) WalkChildren(parent *html.Node,
	l NodeList) (repFirst, repLast *html.Node, err error) {
	nodesCount := make(map[string]int)
	if w.IndexList != nil {
		*w.IndexList = append(*w.IndexList, 0)
	}
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
				parent.FirstChild = repFirst
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
		if w.IndexList != nil {
			(*w.IndexList)[len(*w.IndexList)-1]++
		}
	}
	if w.IndexList != nil {
		*w.IndexList = (*w.IndexList)[:len(*w.IndexList)-1]
	}
	if parent != nil {
		parent.LastChild = repLast
	}
	return
}

// Allow can be used to signal that certain special elements are allowed to be
// encountered by the walker, but won't be processed.
type Allow struct{}

// Process alsways returns true
func (Allow) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	return true, nil, nil
}

// DontDescend can be used to signal that certain elements are allowed and
// should not be descended into.
type DontDescend struct{}

// Process always returns false
func (DontDescend) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	return false, nil, nil
}

// WhitespaceOnly can be used to force TextNodes to be empty.
type WhitespaceOnly struct{}

// Process asserts the text node contains only whitespace.
func (WhitespaceOnly) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	if strings.TrimSpace(n.Data) != "" {
		err = errors.New(": contains illegal text content")
	}
	return
}
