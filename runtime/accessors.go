package runtime

import (
	"github.com/gopherjs/gopherjs/js"
)

// ScalarAccessor is the interface for retrieving and setting a single value
// to an HTML node.
type ScalarAccessor interface {
	get() *js.Object
	set(value interface{})
}

// PropertyAccessor implements ScalarAccessor for a single property of an
// HTML node, such as `textContent` or `value`.
type PropertyAccessor struct {
	node  *js.Object
	pName string
}

// NewPropertyAccessor creates a PropertyAccessor for the node found at the
// given path (relative to root) and the given property name.
func NewPropertyAccessor(root *js.Object, path []int, pName string) *PropertyAccessor {
	return &PropertyAccessor{
		node: walkPath(root, path), pName: pName}
}

func (pa *PropertyAccessor) get() *js.Object {
	return pa.node.Get(pa.pName)
}

func (pa *PropertyAccessor) set(value interface{}) {
	pa.node.Set(pa.pName, value)
}

// StringValue provides access to a dynamic value of string type.
type StringValue struct {
	ScalarAccessor
}

// Get returns the current string value of the linked node.
func (sv *StringValue) Get() string {
	return sv.get().String()
}

// Set updates the underlying node with the given value.
func (sv *StringValue) Set(value string) {
	sv.set(value)
}

// IntValue provides access to a dynamic value of int type.
type IntValue struct {
	ScalarAccessor
}

// Get returns the current value of the linked node.
func (iv *IntValue) Get() int {
	return iv.get().Int()
}

// Set updates the underlying node with the given value.
func (iv *IntValue) Set(value int) {
	iv.set(value)
}
