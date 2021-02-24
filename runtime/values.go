package askew

import "syscall/js"

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
	raw := iv.get()
	switch raw.Type() {
	case js.TypeNumber:
		return raw.Int()
	case js.TypeString:
		return js.Global().Call("parseInt", raw, 10).Int()
	case js.TypeBoolean:
		if raw.Bool() {
			return 1
		}
		return 0
	}
	panic("Cannot retrieve int value from " + raw.String())
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
	raw := bv.get()
	switch raw.Type() {
	case js.TypeBoolean:
		return raw.Bool()
	case js.TypeString:
		str := raw.String()
		return len(str) > 0
	case js.TypeNumber:
		return raw.Int() != 0
	}
	panic("Cannot retrieve bool value from " + raw.String())
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
func (rv *RawValue) Get() js.Value {
	return rv.get()
}

// Set updates the underlying node with the given value.
func (rv *RawValue) Set(value js.Value) {
	rv.set(value)
}
