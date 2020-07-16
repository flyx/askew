package parsers

import (
	"errors"

	"github.com/flyx/askew/data"

	"github.com/yhirose/go-peg"
)

// HandlerSpec describes the content of an <a:handler> node.
type HandlerSpec struct {
	Name    string
	Params  []data.Param
	Returns *data.ParamType
}

var handlerParser *peg.Parser

func init() {
	var err error
	handlerParser, err = peg.NewParser(`
	ROOT        ← < [\n;]* > HANDLER (< [\n;]+ > HANDLER)* < [\n;]* >
	HANDLER     ← IDENTIFIER PARAMLIST TYPE?
	PARAMLIST   ← '(' (PARAM (',' PARAM)*)? ')'
  PARAM       ← IDENTIFIER TYPE` + typeSyntax + identifierSyntax + whitespace)
	if err != nil {
		panic(err)
	}
	g := handlerParser.Grammar

	registerType(handlerParser)
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["PARAM"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.Param{Name: v.ToStr(0), Type: v.Vs[1].(*data.ParamType)}, nil
	}
	g["PARAMLIST"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		names := make(map[string]struct{})
		ret := make([]data.Param, len(v.Vs))
		for i, raw := range v.Vs {
			param := raw.(data.Param)
			_, ok := names[param.Name]
			if ok {
				return nil, errors.New("duplicate param name: " + param.Name)
			}
			names[param.Name] = struct{}{}
			ret[i] = param
		}
		return ret, nil
	}

	g["HANDLER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		params := v.Vs[1].([]data.Param)
		var returns *data.ParamType
		if len(v.Vs) == 3 {
			returns = v.Vs[2].(*data.ParamType)
		}
		return HandlerSpec{
			Name: v.ToStr(0), Params: params, Returns: returns}, nil
	}

	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		handlers := make([]HandlerSpec, len(v.Vs))
		for i := range v.Vs {
			handlers[i] = v.Vs[i].(HandlerSpec)
		}
		return handlers, nil
	}
}

// ParseHandlers parses the content of a <a:handlers> element.
func ParseHandlers(s string) ([]HandlerSpec, error) {
	ret, err := handlerParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]HandlerSpec), nil
}
