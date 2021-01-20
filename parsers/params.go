package parsers

import (
	"errors"

	"github.com/flyx/askew/data"
	"github.com/yhirose/go-peg"
)

var paramParser *peg.Parser

func init() {
	var err error
	paramParser, err = peg.NewParser(`
	ROOT  ← PARAM (',' PARAM)*
	PARAM ← 'var' IDENTIFIER TYPE / IDENTIFIER TYPE
	` + typeSyntax + identifierSyntax + whitespace)
	if err != nil {
		panic(err)
	}
	g := paramParser.Grammar
	registerType(paramParser)
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["PARAM"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.ComponentParam{Name: v.ToStr(0),
			Type: *v.Vs[1].(*data.ParamType), IsVar: v.Choice == 0}, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]data.ComponentParam, len(v.Vs))
		names := make(map[string]struct{})
		for i := range v.Vs {
			ret[i] = v.Vs[i].(data.ComponentParam)
			if _, ok := names[ret[i].Name]; ok {
				return nil, errors.New("duplicate parameter name: " + ret[i].Name)
			}
			names[ret[i].Name] = struct{}{}
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
