package parsers

import "github.com/flyx/askew/data"

// AnalyseArguments analyses the given arguments in s.
func AnalyseArguments(s string) (data.Arguments, error) {
	p := GeneralParser{Buffer: s}
	p.Init()
	if err := p.Parse(int(ruleargs)); err != nil {
		return data.Arguments{}, err
	}
	p.Execute()
	return data.Arguments{Raw: s, Count: len(p.names)}, nil
}
