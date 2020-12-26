package parsers

import (
	"github.com/flyx/askew/data"
	peg "github.com/yhirose/go-peg"
)

var boundSyntax = `
BOUND  ← DATA / PROP / STYLE / CLASS / FORM / EVENT
DATA   ← 'data' '(' HTMLID ')'
PROP   ← 'prop' '(' HTMLID ')'
STYLE  ← 'style' '(' HTMLID ')'
CLASS  ← 'class' '(' HTMLID (',' HTMLID)* ')'
FORM   ← 'form' '(' HTMLID ')'
EVENT  ← 'event' '(' JSID? ')'
HTMLID ← < ([0-9a-zA-Z_] / '-')+ >
JSID   ← < [a-zA-Z_] [0-9a-zA-Z_]* >
`

func strToken(v *peg.Values, d peg.Any) (peg.Any, error) {
	return v.Token(), nil
}

func registerBinders(p *peg.Parser) {
	p.Grammar["HTMLID"].Action = strToken
	p.Grammar["JSID"].Action = strToken
	p.Grammar["PROP"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.BoundValue{Kind: data.BoundProperty, IDs: []string{v.ToStr(0)}}, nil
	}
	p.Grammar["STYLE"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.BoundValue{Kind: data.BoundStyle, IDs: []string{v.ToStr(0)}}, nil
	}
	p.Grammar["DATA"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.BoundValue{Kind: data.BoundData, IDs: []string{v.ToStr(0)}}, nil
	}
	p.Grammar["CLASS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := data.BoundValue{Kind: data.BoundClass, IDs: make([]string, len(v.Vs))}
		for i := range v.Vs {
			ret.IDs[i] = v.ToStr(i)
		}
		return ret, nil
	}
	p.Grammar["FORM"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.BoundValue{Kind: data.BoundFormValue, IDs: []string{v.ToStr(0)}}, nil
	}
	p.Grammar["EVENT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := data.BoundValue{Kind: data.BoundEventValue}
		if len(v.Vs) > 0 {
			ret.IDs = []string{v.ToStr(0)}
		} else {
			ret.IDs = []string{""}
		}
		return ret, nil
	}
	p.Grammar["BOUND"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Vs[0], nil
	}
}
