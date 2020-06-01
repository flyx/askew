package data

// VariableType is the type of a variable inside a component that is bound to
// the DOM.
type VariableType int

const (
	// AutoVar is not a valid type but a transient value for types that are
	// inferred.
	AutoVar VariableType = iota
	// IntVar specifies the variable is an int.
	IntVar
	// StringVar specifies the variable is a string.
	StringVar
	// BoolVar specifies the variable is a bool.
	BoolVar
)

// Variable is a dynamic variable of a component.
type Variable struct {
	Type VariableType
	Name string
}

// VariableMapping maps a Variable to a value in the DOM.
type VariableMapping struct {
	Variable
	Value BoundValue
	Path  []int
}
