package data

import (
	"unicode"

	"github.com/flyx/net/html"
)

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

// ConstructorCallKind describes the kind of a nested constructor call.
type ConstructorCallKind int

const (
	// ConstructDirect creates exactly one instance.
	ConstructDirect ConstructorCallKind = iota
	// ConstructIf creates one instance if the expression evaluates to true
	ConstructIf
	// ConstructFor creates instances with a loop.
	ConstructFor
)

// ConstructorCall describes a <a:construct> node inside <a:embed>.
type ConstructorCall struct {
	ConstructorName string
	Args            Arguments
	Kind            ConstructorCallKind
	Index, Variable string // only for NestedFor
	Expression      string // only for NestedIf and NestedFor
}

// Embed describes a <a:embed> node.
type Embed struct {
	// is a constructor call if Kind == DirectEmbed.
	Args             Arguments
	Kind             EmbedKind
	Path             []int
	Field, Ns, T     string
	Control          bool
	ConstructorCalls []ConstructorCall
}

// Handler describes a <a:handler> node.
type Handler struct {
	Params  []Param
	Returns *ParamType
}

// ControllerMethod is a method of a controller declared with <a:controller>.
type ControllerMethod struct {
	Handler
	// CaptureError contains the error to be yielded if this method is used for
	// capturing events. It is nil if this method can be used to capture events.
	CaptureError error
}

// CanCapture returns true iff this method can be used to capture events.
func (cm ControllerMethod) CanCapture() bool {
	return cm.CaptureError == nil
}

// Capture describe a `a:capture` attribute.
type Capture struct {
	Path     []int
	Mappings []EventMapping
}

// ComponentParam is a component parameter whose type is not parsed or checked by
// Askew, but passed directly through to the Go code generator.
type ComponentParam struct {
	Name  string
	Type  ParamType
	IsVar bool
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
	Unit
	// HTML id. internally generated.
	ID              string
	Name            string
	Parameters      []ComponentParam
	Template        *html.Node
	Fields          []*Field
	Handlers        map[string]Handler
	Controller      map[string]ControllerMethod
	Captures        []Capture
	GenNewInit      bool
	GenList, GenOpt bool
}

// NewName returns the name of the component's new func.
func (c Component) NewName() string {
	runes := []rune(c.Name)
	if unicode.IsLower(runes[0]) {
		return "new" + string(unicode.ToUpper(runes[0])) + string(runes[1:])
	}
	return "New" + c.Name
}
