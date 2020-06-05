package parsers

import (
	"github.com/flyx/askew/data"

	"github.com/yhirose/go-peg"
)

// HandlerSpec describes the content of an <a:handler> node.
type HandlerSpec struct {
	Name   string
	Params []data.GoValue
}

var handlerParser *peg.Parser

func init() {
	var err error
	handlerParser, err = peg.NewParser(`
	ROOT ← IDENTIFIER '(' PARAMLIST? ')'
	` + paramSyntax)
	if err != nil {
		panic(err)
	}
	initParamParsing(handlerParser)
	handlerParser.Grammar["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		var params []data.GoValue
		if len(v.Vs) > 1 {
			params = v.Vs[1].([]data.GoValue)
		}
		return HandlerSpec{Name: v.ToStr(0), Params: params}, nil
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
