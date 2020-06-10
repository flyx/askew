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
	ASSIGNMENT ← BOUND '=' EXPR
	` + boundSyntax + exprSyntax)
	if err != nil {
		panic(err)
	}
	registerBinders(assignmentParser)

	g := assignmentParser.Grammar
	g["IDENTIFIER"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.Token(), nil
	}
	g["EXPR"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return v.S, nil
	}
	g["ASSIGNMENT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		return data.Assignment{Expression: v.ToStr(1), Target: v.Vs[0].(data.BoundValue)}, nil
	}
	g["ROOT"].Action = func(v *peg.Values, d peg.Any) (peg.Any, error) {
		ret := make([]data.Assignment, len(v.Vs))
		for i := range v.Vs {
			ret[i] = v.Vs[i].(data.Assignment)
		}
		return ret, nil
	}
}

// ParseAssignments parses a list of assignments of expressions to attribute names
func ParseAssignments(s string) ([]data.Assignment, error) {
	ret, err := assignmentParser.ParseAndGetValue(s, nil)
	if err != nil {
		return nil, err
	}
	return ret.([]data.Assignment), nil
}
