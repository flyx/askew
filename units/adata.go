package units

import (
	"errors"

	"github.com/flyx/askew/data"
	"github.com/flyx/askew/parsers"
	"github.com/flyx/net/html"
)

type aDataProcessor struct {
	cmp       *data.Component
	indexList *[]int
}

func (dp *aDataProcessor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {

	if dp.cmp.Fields != nil {
		return false, nil, errors.New(": duplicate a:data for component")
	}
	if len(*dp.indexList) != 1 {
		return false, nil, errors.New(": must be defined as direct child of <a:component>")
	}
	def := n.FirstChild
	if def.Type != html.TextNode || def.NextSibling != nil {
		return false, nil, errors.New(": must have plain text as content and nothing else")
	}
	dp.cmp.Fields, err = parsers.ParseFields(def.Data)
	if err != nil {
		return false, nil, errors.New(": unable to parse fields: " + err.Error())
	}
	replacement = &html.Node{Type: html.CommentNode, Data: "data"}
	return
}
