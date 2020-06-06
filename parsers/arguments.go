package parsers

import (
	"github.com/flyx/askew/data"
	"github.com/yhirose/go-peg"
)

var argumentParser *peg.Parser

func init() {
	var err error
	argumentParser, err = peg.NewParser(`
	ROOT ← ARGS?
	ARGS ← EXPR (',' EXPR)*
` + exprSyntax)
	if err != nil {
		panic(err)
	}
	g := argumentParser.Grammar
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return nil, nil
	}
	g["ARGS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return len(v.Vs), nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		if len(v.Vs) > 0 {
			return v.ToInt(0), nil
		}
		return 0, nil
	}
}

// AnalyseArguments analyses the given arguments in s.
func AnalyseArguments(s string) (data.Arguments, error) {
	ret, err := argumentParser.ParseAndGetValue(s, nil)
	if err != nil {
		return data.Arguments{}, err
	}
	return data.Arguments{Raw: s, Count: ret.(int)}, nil
}
