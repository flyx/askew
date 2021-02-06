package units

import (
	"errors"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers/binding"
	"github.com/flyx/net/html"
)

type aDataProcessor struct {
	cmp       *data.Component
	indexList *[]int
}

func (dp *aDataProcessor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {

	if len(*dp.indexList) != 1 {
		return false, nil, errors.New(": must be defined as direct child of <a:component>")
	}
	def := n.FirstChild
	if def.Type != html.TextNode || def.NextSibling != nil {
		return false, nil, errors.New(": must have plain text as content and nothing else")
	}
	fields, err := binding.ParseFields(def.Data)
	if err != nil {
		return false, nil, errors.New(": unable to parse fields: " + err.Error())
	}
	if dp.cmp.Fields != nil {
		names := make(map[string]struct{})
		for _, f := range dp.cmp.Fields {
			names[f.Name] = struct{}{}
		}
		for _, f := range fields {
			if _, ok := names[f.Name]; ok {
				return false, nil, errors.New(": duplicate field name: " + f.Name)
			}
		}
	}
	dp.cmp.Fields = append(dp.cmp.Fields, fields...)

	replacement = &html.Node{Type: html.CommentNode, Data: "data"}
	return
}
