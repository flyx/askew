package main

import peg "github.com/yhirose/go-peg"

type boundKind int

const (
	boundAttribute boundKind = iota
	boundProperty
	boundClass
)

type boundValue struct {
	kind boundKind
	id   string
}

var boundSyntax = `
BOUND ← ATTR / PROP / CLASS
ATTR  ← 'attr' '(' ID ')'
PROP  ← 'prop' '(' ID ')'
CLASS ← 'class' '(' ID ')'
ID    ← < ([0-9a-zA-Z_] / '-')+ >
`

func registerBinders(p *peg.Parser) {
	p.Grammar["PROP"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return boundValue{kind: boundProperty, id: v.ToStr(0)}, nil
	}
	p.Grammar["ATTR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return boundValue{kind: boundAttribute, id: v.ToStr(0)}, nil
	}
	p.Grammar["CLASS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return boundValue{kind: boundClass, id: v.ToStr(0)}, nil
	}
	p.Grammar["BOUND"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Vs[0], nil
	}
}
