package parsers

import (
	"errors"
	"fmt"

	"github.com/flyx/tbc/data"

	"github.com/yhirose/go-peg"
)

type param struct {
	Name string
	Type data.VariableType
}

// HandlerSpec describes the content of a <tbc:handler> node.
type HandlerSpec struct {
	Name   string
	Params map[string]data.VariableType
}

var handlerParser *peg.Parser

func init() {
	var err error
	handlerParser, err = peg.NewParser(`
	ROOT        ← IDENTIFIER '(' PARAMLIST? ')'
	PARAMLIST   ← PARAM (',' PARAM)*
	PARAM       ← IDENTIFIER IDENTIFIER
	IDENTIFIER  ← < [a-zA-Z_][a-zA-Z_0-9]* >
	%whitespace ← [ \t]*
	`)
	if err != nil {
		panic(err)
	}
	g := handlerParser.Grammar
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["PARAM"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := param{Name: v.ToStr(0)}
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
	g["PARAMLIST"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make(map[string]data.VariableType)
		first := v.Vs[0].(param)
		ret[first.Name] = first.Type
		for i := 1; i < len(v.Vs); i++ {
			next := v.Vs[i].(param)
			_, ok := ret[next.Name]
			if ok {
				return nil, errors.New("duplicate param: " + next.Name)
			}
			ret[next.Name] = next.Type
		}
		return ret, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		var params map[string]data.VariableType
		if len(v.Vs) > 1 {
			params = v.Vs[1].(map[string]data.VariableType)
		}
		return HandlerSpec{Name: v.ToStr(0), Params: params}, nil
	}
}

// ParseHandler parses the content of a <tbc:handler> element.
func ParseHandler(s string) (HandlerSpec, error) {
	ret, err := handlerParser.ParseAndGetValue(s, nil)
	if err != nil {
		return HandlerSpec{}, err
	}
	return ret.(HandlerSpec), nil
}
