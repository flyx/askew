package binding

import "github.com/flyx/askew/data"

//go:generate peg -switch grammar.peg

// ParseAssignments parses a list of assignments of expressions to attribute names
func ParseAssignments(s string) ([]data.Assignment, error) {
	p := BindingParser{Buffer: s}
	p.Init()
	if err := p.Parse(int(ruleassignments)); err != nil {
		return nil, err
	}
	p.Execute()
	return p.assignments, nil
}

// ParseBindings parses a list of bindings in an a:bindings attribute
func ParseBindings(s string) ([]data.VariableMapping, error) {
	p := BindingParser{Buffer: s}
	p.Init()
	if err := p.Parse(int(rulebindings)); err != nil {
		return nil, err
	}
	p.Execute()
	return p.varMappings, nil
}
