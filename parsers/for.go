package parsers

import (
	"github.com/flyx/askew/data"
	"github.com/yhirose/go-peg"
)

var forParser *peg.Parser

func init() {
	var err error
	forParser, err = peg.NewParser(`
	ROOT ← IDENTIFIER (',' IDENTIFIER)? ':=' 'range' EXPR
	` + exprSyntax)
	if err != nil {
		panic(err)
	}
	g := forParser.Grammar

	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["EXPR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.S, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := &data.ControlBlock{
			Kind: data.ForBlock, Index: v.ToStr(0)}
		if len(v.Vs) == 3 {
			ret.Variable = v.ToStr(1)
			ret.Expression = v.ToStr(2)
		} else {
			ret.Variable = "_"
			ret.Expression = v.ToStr(1)
		}
		return ret, nil
	}
}

// ParseFor parses the content of an `a:for` attribute.
func ParseFor(s string) (*data.ControlBlock, error) {
	ret, err := forParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.(*data.ControlBlock), nil
}
