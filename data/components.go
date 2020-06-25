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

// EmbedKind describes the type of embed
type EmbedKind int

const (
	// DirectEmbed embeds a component directly so that it is always available
	DirectEmbed EmbedKind = iota
	// ListEmbed embeds a list of components
	ListEmbed
	// OptionalEmbed embeds a component that may or may not exist
	OptionalEmbed
)

// Embed describes a <a:embed> node.
type Embed struct {
	Kind          EmbedKind
	Path          []int
	Field, Pkg, T string
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

// Assignment assigns an expression to a bound value during component instantiation.
// This is usually used with component parameters.
type Assignment struct {
	Expression string
	Path       []int
	Target     BoundValue
}

// Block is a subtree of a component.
type Block struct {
	Assignments []Assignment
	Controlled  []*ControlBlock
}

// ControlBlockKind describes the kind of a control block.
type ControlBlockKind int

const (
	// IfBlock removes the block if its expression evaluates to `false`.
	IfBlock ControlBlockKind = iota
	// ForBlock iterates over its expression and for each iteration, inserts
	// a copy of the original element at its place while evaluating the loop
	// variables in assignments in the original element.
	// The original element is removed from the structure.
	ForBlock
)

// ControlBlock is a block governed by some control structure.
// It is to be processed by its control structure in the component's constructor.
type ControlBlock struct {
	Block
	Kind            ControlBlockKind
	Index, Variable string // only for ForBlock
	Expression      string
	Path            []int
}

// Component describes a <a:component> node.
type Component struct {
	EmbedHost
	Block
	Name string
	// HTML id. internally generated.
	ID         string
	Variables  []VariableMapping
	Handlers   map[string]Handler
	Captures   []Capture
	Parameters []ComponentParam

	Template        *html.Node
	NeedsController bool
	NeedsList       bool
	NeedsOptional   bool
}
