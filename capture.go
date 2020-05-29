package main

import (
	"errors"

	peg "github.com/yhirose/go-peg"
)

type capture struct {
	event         string
	handler       string
	paramMappings map[string]boundValue
}

type captureParser struct {
	p *peg.Parser
}

type paramMapping struct {
	param    string
	supplier boundValue
}

func (cp *captureParser) init() {
	var err error
	cp.p, err = peg.NewParser(`
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
	registerBinders(cp.p)
	g := cp.p.Grammar
	g["VARIABLE"].Action = strToken
	g["EVENT"].Action = strToken
	g["HANDLER"].Action = strToken
	g["MAPPING"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return paramMapping{param: v.ToStr(0), supplier: v.Vs[1].(boundValue)}, nil
	}
	g["MAPPINGS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]boundValue)
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
		ret := capture{event: v.ToStr(0), handler: v.ToStr(1)}
		if v.Len() == 3 {
			ret.paramMappings = v.Vs[2].(map[string]boundValue)
		} else {
			ret.paramMappings = make(map[string]boundValue)
		}
		return ret, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]capture, v.Len())
		for i, c := range v.Vs {
			ret[i] = c.(capture)
		}
		return ret, nil
	}
}

func (cp *captureParser) parse(s string) ([]capture, error) {
	ret, err := cp.p.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]capture), nil
}

var cp captureParser

func init() {
	cp.init()
}
