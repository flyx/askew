package units

import (
	"errors"
	"fmt"
	"strings"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/net/html"
	"github.com/flyx/net/html/atom"
)

type componentProcessor struct {
	unitProcessor
	counter *int
}

// Process reads the given component element.
func (p *componentProcessor) Process(n *html.Node) (descend bool,
	replacement *html.Node, err error) {
	var cmpAttrs attributes.Component
	err = attributes.Collect(n, &cmpAttrs)
	if err != nil {
		return
	}
	if len(cmpAttrs.Name) == 0 {
		return false, nil, errors.New(": attribute `name` missing")
	}

	replacement = &html.Node{Type: html.ElementNode, DataAtom: atom.Template,
		Data: "template"}
	cmp := &data.Component{Unit: data.Unit{}, Template: replacement,
		Name: cmpAttrs.Name, Parameters: cmpAttrs.Params, Init: cmpAttrs.Init,
		OnInclude: cmpAttrs.OnInclude, OnExclude: cmpAttrs.OnExclude}

	err = p.processUnitContent(n, &cmp.Unit, cmp, replacement, true)
	(*p.counter)++
	cmp.ID = fmt.Sprintf("askew-component-%d-%s", *p.counter, strings.ToLower(cmp.Name))
	replacement.Attr = []html.Attribute{{Key: "id", Val: cmp.ID}}

	curFile := p.syms.CurAskewFile()
	if curFile.Components == nil {
		curFile.Components = make(map[string]*data.Component)
	}
	curFile.Components[cmp.Name] = cmp

	return
}
