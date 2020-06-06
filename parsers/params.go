package parsers

import (
	"github.com/flyx/askew/data"
	"github.com/yhirose/go-peg"
)

var paramParser *peg.Parser

func init() {
	var err error
	paramParser, err = peg.NewParser(`
	ROOT  ← PARAM (',' PARAM)*
	PARAM ← IDENTIFIER EXPR
	` + exprSyntax)
	if err != nil {
		panic(err)
	}
	g := paramParser.Grammar
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["EXPR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.S, nil
	}
	g["PARAM"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.ComponentParam{Name: v.ToStr(0), Type: v.ToStr(1)}, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]data.ComponentParam, len(v.Vs))
		for i := range v.Vs {
			ret[i] = v.Vs[i].(data.ComponentParam)
		}
		return ret, nil
	}
}

// ParseParameters parses component parameters from the given string.
func ParseParameters(s string) ([]data.ComponentParam, error) {
	ret, err := paramParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]data.ComponentParam), nil
}
