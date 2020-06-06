package parsers

import (
	"github.com/flyx/askew/data"
	"github.com/yhirose/go-peg"
)

var assignmentParser *peg.Parser

func init() {
	var err error
	assignmentParser, err = peg.NewParser(`
	ROOT ← ASSIGNMENT (',' ASSIGNMENT)*
	ASSIGNMENT ← IDENTIFIER '=' EXPR
	` + exprSyntax)
	if err != nil {
		panic(err)
	}

	g := assignmentParser.Grammar
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["EXPR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.SS, nil
	}
	g["ASSIGNMENT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.ParamAssignment{Expression: v.ToStr(1), AttributeName: v.ToStr(0)}, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]data.ParamAssignment, len(v.Vs))
		for i := range v.Vs {
			ret[i] = v.Vs[i].(data.ParamAssignment)
		}
		return ret, nil
	}
}

// ParseAssignments parses a list of assignments of expressions to attribute names
func ParseAssignments(s string) ([]data.ParamAssignment, error) {
	ret, err := assignmentParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]data.ParamAssignment), nil
}
