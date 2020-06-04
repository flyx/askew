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
	// BoundFormValue is a reference to an element supplying a value to the
	// current <form> element and is only valid within a <form> element.
	//
	// The value is equal to the referenced element's `value` property.
	BoundFormValue
)

// BoundValue specifies the target of a value binding.
type BoundValue struct {
	Kind BoundKind
	ID   string
	// only used if Kind == BoundFormValue. States how many levels above the
	// subject item the <form> is located which is to be used for finding the
	// named element.
	FormDepth int
	// only used if Kind == BoundFormValue. States whether the target element
	// is an <input type=radio>.
	IsRadio bool
}
