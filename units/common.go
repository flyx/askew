package units

import (
	"errors"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/askew/walker"
	"golang.org/x/net/html"
)

type unitProcessor struct {
	syms *data.Symbols
}

func (p *unitProcessor) processUnitContent(
	n *html.Node, unit *data.Unit,
	component *data.Component, replacement *html.Node,
	allowNonWhitespace bool) (err error) {

	p.syms.CurUnit = unit

	var indexList []int
	w := walker.Walker{
		Text:      &aTextProcessor{&unit.Block, &indexList},
		Embed:     &embedProcessor{p.syms, &indexList},
		IndexList: &indexList}
	if component != nil {
		w.Data = &aDataProcessor{component, &indexList}
		w.Controller = &controllerProcessor{p.syms, component, &indexList}
		w.StdElements = &elementHandler{stdElementHandler{p.syms, &indexList, &unit.Block, -1, nil}, component}
		w.Handlers = &handlersProcessor{p.syms, component, &indexList}
	} else {
		w.StdElements = walker.Allow{}
	}
	if allowNonWhitespace {
		w.TextNode = walker.Allow{}
	} else {
		// root may only contain whitespace, but inside there can be text.
		w.TextNode = walker.Allow{}
	}
	replacement.FirstChild, replacement.LastChild, err = w.WalkChildren(
		replacement, &walker.Siblings{Cur: n.FirstChild})

	{
		// reverse Embed list so that they get embedded in reverse order.
		// this is necessary because embedding may change the number of elements in
		// a node, rendering the path of following embeds invalid.
		tmp := make([]data.Embed, len(unit.Embeds))
		for i, e := range unit.Embeds {
			tmp[len(tmp)-i-1] = e
		}
		unit.Embeds = tmp
	}

	{
		// reverse contained control blocks so that they are processed back to front,
		// ensuring that their paths are correct.
		tmp := make([]*data.ControlBlock, len(unit.Controlled))
		for i, e := range unit.Controlled {
			tmp[len(tmp)-i-1] = e
		}
		unit.Controlled = tmp
	}

	return
}

type aTextProcessor struct {
	b         *data.Block
	indexList *[]int
}

func (atp *aTextProcessor) Process(n *html.Node) (descend bool, replacement *html.Node, err error) {
	expr := attributes.Val(n.Attr, "expr")
	if expr == "" {
		return false, nil, errors.New(": missing attribute `expr`")
	}
	if n.FirstChild != nil {
		return false, nil, errors.New(": node may not have child nodes")
	}
	atp.b.Assignments = append(atp.b.Assignments, data.Assignment{
		Expression: expr, Path: append([]int(nil), *atp.indexList...), Target: data.BoundValue{Kind: data.BoundSelf}})
	return false, &html.Node{Type: html.CommentNode, Data: "a:text"}, nil
}
