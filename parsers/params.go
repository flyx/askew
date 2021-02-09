package parsers

import "github.com/flyx/askew/data"

// ParseParameters parses component parameters from the given string.
func ParseParameters(s string) ([]data.ComponentParam, error) {
	p := GeneralParser{Buffer: s}
	p.Init()
	if err := p.Parse(int(rulecparams)); err != nil {
		return nil, err
	}
	p.Execute()
	return p.cParams, nil
}
