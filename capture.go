package main

import (
	"errors"

	peg "github.com/yhirose/go-peg"
)

type paramSupplierKind int

const (
	attrSupplier paramSupplierKind = iota
	propSupplier
)

type paramSupplier struct {
	kind paramSupplierKind
	id   string
}

type capture struct {
	event         string
	handler       string
	paramMappings map[string]paramSupplier
}

type captureParser struct {
	p *peg.Parser
}

type paramMapping struct {
	param    string
	supplier paramSupplier
}

func (cp *captureParser) init() {
	var err error
	cp.p, err = peg.NewParser(`
	ROOT        ← CAPTURE (',' CAPTURE)*
	CAPTURE     ← EVENT '=' HANDLER ('(' MAPPINGS? ')')?
	EVENT       ← < [a-z]+ >
	HANDLER     ← < [a-zA-Z_][a-zA-Z_0-9]* >
	MAPPINGS    ← MAPPING (',' MAPPING)*
	MAPPING     ← VARIABLE '=' SUPPLIER
	VARIABLE    ← < [a-zA-ZL_][a-zA-Z_0-9]* >
	SUPPLIER    ← ATTR / PROP
	ATTR        ← 'attr' '(' ID ')'
	PROP        ← 'prop' '(' ID ')'
	ID          ← < ([0-9a-zA-Z_] / '-')+ >
	%whitespace ← [ \t]*
	`)
	if err != nil {
		panic(err)
	}
	g := cp.p.Grammar
	strToken := func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["ID"].Action = strToken
	g["VARIABLE"].Action = strToken
	g["EVENT"].Action = strToken
	g["HANDLER"].Action = strToken
	g["PROP"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return paramSupplier{kind: propSupplier, id: v.ToStr(0)}, nil
	}
	g["ATTR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return paramSupplier{kind: attrSupplier, id: v.ToStr(0)}, nil
	}
	g["SUPPLIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Vs[0], nil
	}
	g["MAPPING"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return paramMapping{param: v.ToStr(0), supplier: v.Vs[1].(paramSupplier)}, nil
	}
	g["MAPPINGS"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]paramSupplier)
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
			ret.paramMappings = v.Vs[2].(map[string]paramSupplier)
		} else {
			ret.paramMappings = make(map[string]paramSupplier)
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
