package data

// GoValue is a named, typed value that is exposed to a component's Go interface.
// It is used for dynamic variables and component parameters.
type GoValue struct {
	Type *ParamType
	Name string
}

// VariableMapping maps a Variable to a value in the DOM.
type VariableMapping struct {
	Variable GoValue
	Value    BoundValue
	Path     []int
}
