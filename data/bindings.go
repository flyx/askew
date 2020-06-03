package data

// BoundKind specifies the kind of a bound value.
type BoundKind int

const (
	// BoundData is a bound HTML dataset item.
	BoundData BoundKind = iota
	// BoundProperty is a bound property of DOM.Node.
	BoundProperty
	// BoundClass is a bound property of a node's classList
	BoundClass
)

// BoundValue specifies the target of a value binding.
type BoundValue struct {
	Kind BoundKind
	ID   string
}
