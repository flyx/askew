package data

import "golang.org/x/net/html"

// Embed describes a <tbc:embed> node.
type Embed struct {
	Path          []int
	Field, Pkg, T string
	List          bool
}

// Handler describes a <tbc:handler> node.
type Handler struct {
	Params map[string]VariableType
}

// Capture describe a `tbc:capture` attribute.
type Capture struct {
	Path     []int
	Mappings []EventMapping
}

// Component describes a <tbc:component> node.
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
