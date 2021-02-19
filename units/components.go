package units

import (
	"errors"

	"github.com/flyx/askew/attributes"
	"github.com/flyx/askew/data"
	"github.com/flyx/net/html"
)

type componentProcessor struct {
	unitProcessor
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

	replacement = &html.Node{Type: html.DocumentNode}
	cmp := &data.Component{Unit: data.Unit{}, Template: replacement,
		Name: cmpAttrs.Name, Parameters: cmpAttrs.Params,
		GenNewInit: cmpAttrs.GenNewInit}
	for _, param := range cmp.Parameters {
		if param.IsVar {
			t := param.Type
			v := param.Name
			cmp.Fields = append(cmp.Fields, &data.Field{Name: param.Name, Type: &t,
				DefaultValue: &v})
		}
	}

	err = p.processUnitContent(n, &cmp.Unit, cmp, replacement, true)

	curFile := p.syms.CurAskewFile()
	if curFile.Components == nil {
		curFile.Components = make(map[string]*data.Component)
	}
	curFile.Components[cmp.Name] = cmp

	return
}
