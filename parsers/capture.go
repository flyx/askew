package parsers

import (
	"errors"

	"github.com/flyx/tbc/data"
	peg "github.com/yhirose/go-peg"
)

type paramMapping struct {
	param    string
	supplier data.BoundValue
}

var captureParser *peg.Parser

func init() {
	var err error
	captureParser, err = peg.NewParser(`
	ROOT        ← CAPTURE (',' CAPTURE)*
	CAPTURE     ← EVENT ':' HANDLER ('(' MAPPINGS? ')')?
	EVENT       ← < [a-z]+ >
	HANDLER     ← < [a-zA-Z_][a-zA-Z_0-9]* >
	MAPPINGS    ← MAPPING (',' MAPPING)*
	MAPPING     ← VARIABLE '=' BOUND
	VARIABLE    ← < [a-zA-Z_][a-zA-Z_0-9]* >
	%whitespace ← [ \t]*
	` + boundSyntax)
	if err != nil {
		panic(err)
	}
	registerBinders(captureParser)
	g := captureParser.Grammar
	g["VARIABLE"].Action = strToken
	g["EVENT"].Action = strToken
	g["HANDLER"].Action = strToken
	g["MAPPING"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return paramMapping{param: v.ToStr(0), supplier: v.Vs[1].(data.BoundValue)}, nil
	}
	g["MAPPINGS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]data.BoundValue)
		first := v.Vs[0].(paramMapping)
		ret[first.param] = first.supplier
		for i := 1; i < len(v.Vs); i++ {
			next := v.Vs[i].(paramMapping)
			_, ok := ret[next.param]
			if ok {
				return nil, errors.New("duplicate param: " + next.param)
			}
			ret[next.param] = next.supplier
		}
		return ret, nil
	}
	g["CAPTURE"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := data.EventMapping{Event: v.ToStr(0), Handler: v.ToStr(1)}
		if v.Len() == 3 {
			ret.ParamMappings = v.Vs[2].(map[string]data.BoundValue)
		} else {
			ret.ParamMappings = make(map[string]data.BoundValue)
		}
		return ret, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]data.EventMapping, v.Len())
		for i, c := range v.Vs {
			ret[i] = c.(data.EventMapping)
		}
		return ret, nil
	}
}

// ParseCapture parses the content of a tbc:capture attribute.
func ParseCapture(s string) ([]data.EventMapping, error) {
	ret, err := captureParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]data.EventMapping), nil
}
