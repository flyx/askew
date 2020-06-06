package data

import "golang.org/x/net/html"

// Arguments is a list of arguments of an <a:embed>.
// The arguments will be mapped to the parameters of the referenced <a:component>.
// Askew will only check that the amount of arguments matches the amount of
// parameters. Apart from that, the argument string is passed through to the
// generated Go code.
type Arguments struct {
	Raw   string
	Count int
}

// Embed describes a <a:embed> node.
type Embed struct {
	Path          []int
	Field, Pkg, T string
	List          bool
	Args          Arguments
}

// Handler describes a <a:handler> node.
type Handler struct {
	Params []GoValue
}

// Capture describe a `a:capture` attribute.
type Capture struct {
	Path     []int
	Mappings []EventMapping
}

// ComponentParam is a component parameter whose type is not parsed or checked by
// Askew, but passed directly through to the Go code generator.
type ComponentParam struct {
	Name, Type string
}

// ParamAssignment assigns the value of a parameter to the attribute with the
// given name of the element at the given path. If AttributeName is empty, the
// element is replaced with a text node containing the parameter's value instead.
type ParamAssignment struct {
	Expression    string
	Path          []int
	AttributeName string
}

// Conditional includes or excludes an element from a component instance depending
// on a condition.
type Conditional struct {
	Condition string
	Path      []int
}

// Component describes a <a:component> node.
type Component struct {
	EmbedHost
	Name string
	// HTML id. internally generated.
	ID              string
	Variables       []VariableMapping
	Handlers        map[string]Handler
	Captures        []Capture
	Parameters      []ComponentParam
	Assignments     []ParamAssignment
	Conditionals    []Conditional
	Template        *html.Node
	NeedsController bool
	NeedsList       bool
}
