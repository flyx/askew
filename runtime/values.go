package runtime

import "github.com/gopherjs/gopherjs/js"

// StringValue provides access to a dynamic value of string type.
type StringValue struct {
	BoundValue
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
	BoundValue
}

// Get returns the current value of the linked node.
func (iv *IntValue) Get() int {
	return iv.get().Int()
}

// Set updates the underlying node with the given value.
func (iv *IntValue) Set(value int) {
	iv.set(value)
}

// BoolValue provides access to a dynamic value of bool type.
type BoolValue struct {
	BoundValue
}

// Get returns the current value of the linked node.
func (bv *BoolValue) Get() bool {
	return bv.get().Bool()
}

// Set updates the underlying node with the given value.
func (bv *BoolValue) Set(value bool) {
	bv.set(value)
}

// RawValue provides acces to the raw underlying value.
type RawValue struct {
	BoundValue
}

// Get returns the current value of the linked node.
func (rv *RawValue) Get() *js.Object {
	return rv.get()
}

// Set updates the underlying node with the given value.
func (rv *RawValue) Set(value *js.Object) {
	rv.set(value)
}
