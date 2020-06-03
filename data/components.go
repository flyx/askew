package data

import "golang.org/x/net/html"

// Embed describes a <a:embed> node.
type Embed struct {
	Path          []int
	Field, Pkg, T string
	List          bool
}

// Handler describes a <a:handler> node.
type Handler struct {
	Params map[string]VariableType
}

// Capture describe a `a:capture` attribute.
type Capture struct {
	Path     []int
	Mappings []EventMapping
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
	Template        *html.Node
	NeedsController bool
	NeedsList       bool
}
