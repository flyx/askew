package main

import (
	"errors"
	"fmt"

	"github.com/yhirose/go-peg"
)

type param struct {
	name string
	t    valueKind
}

type handlerSpec struct {
	name   string
	params map[string]valueKind
}

type handlerParser struct {
	p *peg.Parser
}

func (hp *handlerParser) init() {
	var err error
	hp.p, err = peg.NewParser(`
	ROOT        ← IDENTIFIER '(' PARAMLIST? ')'
	PARAMLIST   ← PARAM (',' PARAM)*
	PARAM       ← IDENTIFIER IDENTIFIER
	IDENTIFIER  ← < [a-zA-Z_][a-zA-Z_0-9]* >
	%whitespace ← [ \t]*
	`)
	if err != nil {
		panic(err)
	}
	g := hp.p.Grammar
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["PARAM"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := param{name: v.ToStr(0)}
		switch v.ToStr(1) {
		case "int":
			ret.t = intVal
		case "string":
			ret.t = stringVal
		case "bool":
			ret.t = boolVal
		default:
			return nil, fmt.Errorf("unsupported type: %s", v.ToStr(1))
		}
		return ret, nil
	}
	g["PARAMLIST"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]valueKind)
		first := v.Vs[0].(param)
		ret[first.name] = first.t
		for i := 1; i < len(v.Vs); i++ {
			next := v.Vs[i].(param)
			_, ok := ret[next.name]
			if ok {
				return nil, errors.New("duplicate param: " + next.name)
			}
			ret[next.name] = next.t
		}
		return ret, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		var params map[string]valueKind
		if len(v.Vs) > 1 {
			params = v.Vs[1].(map[string]valueKind)
		}
		return handlerSpec{name: v.ToStr(0), params: params}, nil
	}
}

func (hp *handlerParser) parse(s string) (handlerSpec, error) {
	ret, err := hp.p.ParseAndGetValue(s, nil)
	if err != nil {
		return handlerSpec{}, err
	}
	return ret.(handlerSpec), nil
}

var hp handlerParser

func init() {
	hp.init()
}
