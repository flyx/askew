package parsers

import "github.com/flyx/askew/data"

// ParseFor parses the content of an `a:for` attribute.
func ParseFor(s string) (*data.ControlBlock, error) {
	p := GeneralParser{Buffer: s}
	p.Init()
	if err := p.Parse(int(rulefor)); err != nil {
		return nil, err
	}
	p.Execute()
	ret := &data.ControlBlock{
		Kind: data.ForBlock, Index: p.names[0], Expression: p.expr}
	if len(p.names) == 2 {
		ret.Variable = p.names[1]
	} else {
		ret.Variable = "_"
	}
	return ret, nil
}
