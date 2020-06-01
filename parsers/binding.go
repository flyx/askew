package parsers

import (
	"github.com/flyx/tbc/data"
	peg "github.com/yhirose/go-peg"
)

var boundSyntax = `
BOUND  ← ATTR / PROP / CLASS
ATTR   ← 'attr' '(' HTMLID ')'
PROP   ← 'prop' '(' HTMLID ')'
CLASS  ← 'class' '(' HTMLID ')'
HTMLID ← < ([0-9a-zA-Z_] / '-')+ >
`

func strToken(v *peg.Values, d peg.Any) (peg.Any, error) {
	return v.Token(), nil
}

func registerBinders(p *peg.Parser) {
	p.Grammar["HTMLID"].Action = strToken
	p.Grammar["PROP"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.BoundValue{Kind: data.BoundProperty, ID: v.ToStr(0)}, nil
	}
	p.Grammar["ATTR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.BoundValue{Kind: data.BoundAttribute, ID: v.ToStr(0)}, nil
	}
	p.Grammar["CLASS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.BoundValue{Kind: data.BoundClass, ID: v.ToStr(0)}, nil
	}
	p.Grammar["BOUND"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Vs[0], nil
	}
}
