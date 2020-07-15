package parsers

import (
	"errors"
	"fmt"

	"github.com/flyx/askew/data"

	"github.com/yhirose/go-peg"
)

// HandlerSpec describes the content of an <a:handler> node.
type HandlerSpec struct {
	Name        string
	Params      []data.GoValue
	ReturnsBool bool
}

var handlerParser *peg.Parser

func init() {
	var err error
	handlerParser, err = peg.NewParser(`
	ROOT        ← IDENTIFIER PARAMLIST IDENTIFIER?
	PARAMLIST   ← '(' (PARAM (',' PARAM)*)? ')'
  PARAM       ← IDENTIFIER IDENTIFIER` + identifierSyntax + whitespace)
	if err != nil {
		panic(err)
	}
	g := handlerParser.Grammar

	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["PARAM"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
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
	g["PARAMLIST"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
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

	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		params := v.Vs[1].([]data.GoValue)
		returnsBool := false
		if len(v.Vs) == 3 {
			retType := v.ToStr(2)
			if retType == "bool" {
				returnsBool = true
			}
			return nil, errors.New("unsupported return type: " + retType)
		}
		return HandlerSpec{
			Name: v.ToStr(0), Params: params, ReturnsBool: returnsBool}, nil
	}
}

// ParseHandler parses the content of a <a:handler> element.
func ParseHandler(s string) (HandlerSpec, error) {
	ret, err := handlerParser.ParseAndGetValue(s, nil)
	if err != nil {
		return HandlerSpec{}, err
	}
	return ret.(HandlerSpec), nil
}
