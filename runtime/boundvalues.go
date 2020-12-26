package runtime

import "github.com/gopherjs/gopherjs/js"

// BoundValue is the interface for retrieving and setting bound values in
// the HTML DOM.
type BoundValue interface {
	get() *js.Object
	set(value interface{})
}

// BoundProperty implements BoundValue for a single property of an
// HTML node, such as `textContent` or `value`.
type BoundProperty struct {
	node  *js.Object
	pName string
}

// NewBoundProperty creates a BoundProperty for the node found at the
// given path (relative to root) and the given property name.
func NewBoundProperty(
	d *ComponentData, pName string, path ...int) *BoundProperty {
	return &BoundProperty{
		node: d.Walk(path...), pName: pName}
}

// Init initializes the bound property with the given node and property name.
func (bp *BoundProperty) Init(node *js.Object, pName string) {
	bp.node, bp.pName = node, pName
}

func (bp *BoundProperty) get() *js.Object {
	return bp.node.Get(bp.pName)
}

func (bp *BoundProperty) set(value interface{}) {
	bp.node.Set(bp.pName, value)
}

// BoundData implements BoundValue for an item in the dataset of an HTML node.
type BoundData struct {
	node  *js.Object
	dName string
}

// NewBoundData creates a BoundData for the node found at the given
// path (relative to root) and the given dataset item.
func NewBoundData(
	d *ComponentData, dName string, path ...int) *BoundData {
	return &BoundData{node: d.Walk(path...), dName: dName}
}

// Init initializes the object with the given node and dataset item name.
func (ba *BoundData) Init(node *js.Object, dName string) {
	ba.node, ba.dName = node, dName
}

func (ba *BoundData) get() *js.Object {
	return ba.node.Get("dataset").Get(ba.dName)
}

func (ba *BoundData) set(value interface{}) {
	ba.node.Get("dataset").Set(ba.dName, value)
}

// BoundClass implements BoundValue for a node so that assigning boolean
// values to it switches a class with a given name on or off on the target node.
// may only be used with boolean values.
type BoundClass struct {
	node      *js.Object
	className string
}

// NewBoundClass creates a BoundClass for the node at the given path,
// which switches the class with the given name.
func NewBoundClass(d *ComponentData, className string, path ...int) *BoundClass {
	return &BoundClass{node: d.Walk(path...), className: className}
}

// Init initializes the BoundClass with the given node and class name.
func (bc *BoundClass) Init(node *js.Object, className string) {
	bc.node, bc.className = node, className
}

func (bc *BoundClass) get() *js.Object {
	return bc.node.Get("classList").Call("contains", bc.className)
}

func (bc *BoundClass) set(value interface{}) {
	if value.(bool) {
		bc.node.Get("classList").Call("add", bc.className)
	} else {
		bc.node.Get("classList").Call("remove", bc.className)
	}
}

// BoundFormValue implements BoundValue as a reference to an element supplying
// a value to the current form.
type BoundFormValue struct {
	form  *js.Object
	name  string
	radio bool
}

// NewBoundFormValue creates a BoundFormValue for the from at the given path.
// radio must be true iff the input with the given name has type=radio.
func NewBoundFormValue(d *ComponentData, name string, radio bool, path ...int) *BoundFormValue {
	ret := new(BoundFormValue)
	ret.Init(d.Walk(path...), name, radio)
	return ret
}

// Init initializes the BoundFormValue with the given form and element name.
func (bfv *BoundFormValue) Init(form *js.Object, name string, radio bool) {
	bfv.form, bfv.name, bfv.radio = form, name, radio
}

func (bfv *BoundFormValue) get() *js.Object {
	if bfv.radio {
		list := bfv.form.Get("elements").Get(bfv.name)
		for i := 0; i < list.Length(); i++ {
			item := list.Index(i)
			if item.Get("checked").Bool() {
				return item.Get("value")
			}
		}
		return nil
	}
	return bfv.form.Get("elements").Get(bfv.name).Get("value")
}

func (bfv *BoundFormValue) set(value interface{}) {
	elm := bfv.form.Get("elements").Get(bfv.name)
	if bfv.radio {
		for i := 0; i < elm.Length(); i++ {
			item := elm.Index(i)
			if item.Get("value") == value {
				item.Set("checked", true)
				return
			}
		}
		panic("unknown radio value!")
	}
	elm.Set("value", value)
}

// BoundEventValue implements BoundValue as a reference to a value of the
// captured event, or the event itself.
type BoundEventValue struct {
	val *js.Object
}

// Init initializes the BoundEventValue to return the given event's property
// with the given name, or the event itself if propName == ""
func (bev *BoundEventValue) Init(e *js.Object, propName string) {
	if propName == "" {
		bev.val = e
	} else {
		bev.val = e.Get(propName)
	}
}

func (bev *BoundEventValue) get() *js.Object {
	return bev.val
}

func (bev *BoundEventValue) set(value interface{}) {
	panic("BoundEvent doesn't support set()")
}

// BoundSelf implements BoundValue by replacing the referenced node with a text node
// containing the given value.
type BoundSelf struct {
	node *js.Object
}

// NewBoundSelf creates a BoundSelf for the node at the given path.
func NewBoundSelf(d *ComponentData, dummy string, path ...int) *BoundSelf {
	return &BoundSelf{node: d.Walk(path...)}
}

// Init initializes the BoundSelf with the given node.
func (bs *BoundSelf) Init(node *js.Object, dummy string) {
	bs.node = node
}

func (bs *BoundSelf) get() *js.Object {
	panic("BoundSelf doesn't support get()")
}

func (bs *BoundSelf) set(value interface{}) {
	node := js.Global.Get("document").Call("createTextNode", value)
	bs.node.Get("parentNode").Call("replaceChild", node, bs.node)
}

// Assign is low-level assignment of a value to a bound value.
func Assign(bv BoundValue, value interface{}) {
	bv.set(value)
}
