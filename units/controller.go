package units

import (
	"errors"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers"
	"github.com/flyx/net/html"
)

type controllerProcessor struct {
	syms      *data.Symbols
	cmp       *data.Component
	indexList *[]int
}

func (cp *controllerProcessor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	if len(*cp.indexList) != 1 {
		return false, nil, errors.New(": must be defined as direct child of <a:component>")
	}
	def := n.FirstChild
	if def.Type != html.TextNode || def.NextSibling != nil {
		return false, nil, errors.New(": must have plain text as content and nothing else")
	}
	if cp.cmp.Controller != nil {
		return false, nil, errors.New(": only one <a:controller> allowed per <a:component>")
	}
	parsed, err := parsers.ParseHandlers(def.Data)
	if err != nil {
		return false, nil, errors.New(": unable to parse `" + def.Data + "`: " + err.Error())
	}
	cp.cmp.Controller = make(map[string]data.ControllerMethod)
	for _, raw := range parsed {
		_, ok := cp.cmp.Controller[raw.Name]
		if !ok && cp.cmp.Handlers != nil {
			_, ok = cp.cmp.Handlers[raw.Name]
		}
		if ok {
			return false, nil, errors.New(": duplicate handler name: " + raw.Name)
		}
		cp.cmp.Controller[raw.Name] =
			data.ControllerMethod{
				Handler:      data.Handler{Params: raw.Params, Returns: raw.Returns},
				CaptureError: canCapture(raw.Params)}
	}

	replacement = &html.Node{Type: html.CommentNode, Data: "controller"}
	return
}
