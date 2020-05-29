package main

import (
	"fmt"

	"github.com/yhirose/go-peg"
)

type bindingsParser struct {
	p *peg.Parser
}

type valueKind int

const (
	autoVal valueKind = iota
	intVal
	stringVal
	boolVal
)

type targetVar struct {
	kind valueKind
	name string
}

type varBinding struct {
	value    boundValue
	variable targetVar
}

func (bp *bindingsParser) init() {
	var err error
	bp.p, err = peg.NewParser(`
	ROOT        ← BINDING (',' BINDING)*
	BINDING     ← BOUND ':' (AUTOVAR / TYPEDVAR)
	AUTOVAR     ← IDENTIFIER
	TYPEDVAR    ← '(' IDENTIFIER IDENTIFIER ')'
	IDENTIFIER  ← < [a-zA-Z_][a-zA-Z_0-9]* >
	%whitespace ← [ \t]*
	` + boundSyntax)
	if err != nil {
		panic(err)
	}
	registerBinders(bp.p)
	g := bp.p.Grammar
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["TYPEDVAR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		switch v.ToStr(1) {
		case "int":
			return targetVar{kind: intVal, name: v.ToStr(0)}, nil
		case "string":
			return targetVar{kind: stringVal, name: v.ToStr(0)}, nil
		case "bool":
			return targetVar{kind: boolVal, name: v.ToStr(0)}, nil
		default:
			return nil, fmt.Errorf("unsupported type: %s", v.ToStr(1))
		}
	}
	g["AUTOVAR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return targetVar{kind: autoVal, name: v.ToStr(0)}, nil
	}
	g["BINDING"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return varBinding{value: v.Vs[0].(boundValue), variable: v.Vs[1].(targetVar)}, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]varBinding, v.Len())
		for i, c := range v.Vs {
			ret[i] = c.(varBinding)
		}
		return ret, nil
	}
}

func (bp *bindingsParser) parse(s string) ([]varBinding, error) {
	ret, err := bp.p.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]varBinding), nil
}

var bp bindingsParser

func init() {
	bp.init()
}
