package binding

import "github.com/flyx/askew/data"

//go:generate peg -switch grammar.peg

// ParseFields parses the content of a <a:data> element.
func ParseFields(s string) ([]*data.Field, error) {
	p := BindingParser{Buffer: s}
	p.Init()
	if err := p.Parse(int(rulefields)); err != nil {
		return nil, err
	}
	p.Execute()
	return p.fields, nil
}
