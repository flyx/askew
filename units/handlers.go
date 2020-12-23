package units

import (
	"errors"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers"
	"github.com/flyx/net/html"
)

func canCapture(params []data.Param) bool {
	for _, p := range params {
		switch p.Type.Kind {
		case data.IntType, data.StringType, data.BoolType:
			break
		default:
			return false
		}
	}
	return true
}

type handlersProcessor struct {
	syms      *data.Symbols
	cmp       *data.Component
	indexList *[]int
}

func (hp *handlersProcessor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	if len(*hp.indexList) != 1 {
		return false, nil, errors.New(": must be defined as direct child of <a:component>")
	}
	def := n.FirstChild
	if def.Type != html.TextNode || def.NextSibling != nil {
		return false, nil, errors.New(": must have plain text as content and nothing else")
	}
	parsed, err := parsers.ParseHandlers(def.Data)
	if err != nil {
		return false, nil, errors.New(": unable to parse `" + def.Data + "`: " + err.Error())
	}
	if hp.cmp.Handlers != nil {
		return false, nil, errors.New(": only one <a:handlers> allowed per <a:component>")
	}
	hp.cmp.Handlers = make(map[string]data.Handler)
	for _, raw := range parsed {
		_, ok := hp.cmp.Handlers[raw.Name]
		if !ok && hp.cmp.Controller != nil {
			_, ok = hp.cmp.Controller[raw.Name]
		}
		if ok {
			return false, nil, errors.New(": duplicate handler name: " + raw.Name)
		}
		if !canCapture(raw.Params) {
			return false, nil, errors.New(": handlers must only use int, string and bool as parameter types")
		}
		hp.cmp.Handlers[raw.Name] =
			data.Handler{Params: raw.Params, Returns: raw.Returns}
	}

	replacement = &html.Node{Type: html.CommentNode, Data: "handlers"}
	return
}
