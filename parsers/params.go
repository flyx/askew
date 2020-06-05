package parsers

import (
	"errors"
	"fmt"

	"github.com/flyx/askew/data"
	"github.com/yhirose/go-peg"
)

var paramSyntax = `
PARAMLIST   ← PARAM (',' PARAM)*
PARAM       ← IDENTIFIER IDENTIFIER
IDENTIFIER  ← < [a-zA-Z_][a-zA-Z_0-9]* >
%whitespace ← [ \t]*
`

func initParamParsing(p *peg.Parser) {
	p.Grammar["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	p.Grammar["PARAM"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := data.GoValue{Name: v.ToStr(0)}
		switch v.ToStr(1) {
		case "int":
			ret.Type = data.IntVar
		case "string":
			ret.Type = data.StringVar
		case "bool":
			ret.Type = data.BoolVar
		default:
			return nil, fmt.Errorf("unsupported type: %s", v.ToStr(1))
		}
		return ret, nil
	}
	p.Grammar["PARAMLIST"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		names := make(map[string]struct{})
		ret := make([]data.GoValue, len(v.Vs))
		for i, raw := range v.Vs {
			param := raw.(data.GoValue)
			_, ok := names[param.Name]
			if ok {
				return nil, errors.New("duplicate param name: " + param.Name)
			}
			names[param.Name] = struct{}{}
			ret[i] = param
		}
		return ret, nil
	}
}

var paramListParser *peg.Parser

func init() {
	var err error
	paramListParser, err = peg.NewParser(paramSyntax)
	if err != nil {
		panic(err)
	}
	initParamParsing(paramListParser)
}

// ParseParamList parses a string containing a parameter list.
func ParseParamList(s string) ([]data.GoValue, error) {
	ret, err := paramListParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]data.GoValue), nil
}
