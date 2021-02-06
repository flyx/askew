package arguments

import "github.com/flyx/askew/data"

//go:generate peg -switch grammar.peg

// Analyse analyses the given arguments in s.
func Analyse(s string) (data.Arguments, error) {
	p := ArgumentParser{Buffer: s}
	p.Init()
	if err := p.Parse(); err != nil {
		return data.Arguments{}, err
	}
	p.Execute()
	return data.Arguments{Raw: s, Count: p.count}, nil
}
