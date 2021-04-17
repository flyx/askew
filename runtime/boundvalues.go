package askew

import "syscall/js"

// BoundValue is the interface for retrieving and setting bound values in
// the HTML DOM.
type BoundValue interface {
	get() js.Value
	set(value interface{})
}

// BoundProperty implements BoundValue for a single property of an
// HTML node, such as `textContent` or `value`.
type BoundProperty struct {
	node  js.Value
	pName string
}

// NewBoundProperty creates a BoundProperty for the node found at the
// given path (relative to root) and the given property name.
func NewBoundProperty(
	d *ComponentData, pName string, path ...int) *BoundProperty {
	return &BoundProperty{
		node: d.Walk(path...), pName: pName}
}

// BoundPropertyAt returns a BoundProperty for the given node and the given
// property name.
func BoundPropertyAt(node js.Value, pName string) *BoundProperty {
	return &BoundProperty{node: node, pName: pName}
}

func (bp *BoundProperty) get() js.Value {
	return bp.node.Get(bp.pName)
}

func (bp *BoundProperty) set(value interface{}) {
	bp.node.Set(bp.pName, value)
}

// BoundStyle implements BoundValue for a single property of the target node's
// `style` property.
type BoundStyle struct {
	node  js.Value
	sName string
}

// NewBoundStyle creates a BoundStyle for the node found at the
// given path (relative to root) and the given style name.
func NewBoundStyle(
	d *ComponentData, sName string, path ...int) *BoundStyle {
	return &BoundStyle{
		node: d.Walk(path...), sName: sName}
}

// BoundStyleAt returns a BoundStyle for the given node and the given style
// item name.
func BoundStyleAt(node js.Value, sName string) *BoundStyle {
	return &BoundStyle{node: node, sName: sName}
}

func (bs *BoundStyle) get() js.Value {
	return bs.node.Get("style").Get(bs.sName)
}

func (bs *BoundStyle) set(value interface{}) {
	bs.node.Get("style").Set(bs.sName, value)
}

// BoundDataset implements BoundValue for an item in the dataset of an HTML node.
type BoundDataset struct {
	node  js.Value
	dName string
}

// NewBoundDataset creates a BoundData for the node found at the given
// path (relative to root) and the given dataset item.
func NewBoundDataset(
	d *ComponentData, dName string, path ...int) *BoundDataset {
	return &BoundDataset{node: d.Walk(path...), dName: dName}
}

// BoundDatasetAt returns a bound dataset for the given node and the given
// dataset item name.
func BoundDatasetAt(node js.Value, dName string) *BoundDataset {
	return &BoundDataset{node: node, dName: dName}
}

func (ba *BoundDataset) get() js.Value {
	return ba.node.Get("dataset").Get(ba.dName)
}

func (ba *BoundDataset) set(value interface{}) {
	ba.node.Get("dataset").Set(ba.dName, value)
}

// BoundClasses implements BoundValue for a node's classList.
//
// Given a list of class names, it can shuffle between them, letting the
// classList contain at most one of the given names.
//
// BoundClasses supports a boolean interface, where containing any given
// class name equals `true` and not containing it equals `false`. This should
// not be used if multiple class names are given. Setting the value to true will
// always set it to the first class name.
//
// The interface also support integral values where `0` equals none of the
// given class names present, and positive numbers enumerate the given names
// beginning with `1` for the first name.
//
// Both the boolean and the integer interface can be used for both reading and
// writing.
type BoundClasses struct {
	node       js.Value
	classNames []string
}

// NewBoundClasses creates a BoundClasses for the node at the given path,
// which switches the class with the given name.
func NewBoundClasses(d *ComponentData, classNames []string, path ...int) *BoundClasses {
	return &BoundClasses{node: d.Walk(path...), classNames: classNames}
}

// BoundClassesAt returns a BoundClasses for the given node and the given
// list of class names.
func BoundClassesAt(node js.Value, classNames []string) *BoundClasses {
	return &BoundClasses{node: node, classNames: classNames}
}

func (bc *BoundClasses) get() js.Value {
	cList := bc.node.Get("classList")
	for i := range bc.classNames {
		if cList.Call("contains", bc.classNames[i]).Bool() {
			return js.Global().Call("Number", i+1)
		}
	}
	return js.Global().Call("Number", 0)
}

func (bc *BoundClasses) set(value interface{}) {
	var iVal int
	if bVal, ok := value.(bool); ok {
		if bVal {
			iVal = 0
		} else {
			iVal = -1
		}
	} else {
		iVal = value.(int) - 1
	}
	cList := bc.node.Get("classList")
	for i := range bc.classNames {
		if i == iVal {
			cList.Call("add", bc.classNames[i])
		} else {
			cList.Call("remove", bc.classNames[i])
		}
	}
}

// BoundFormValue implements BoundValue as a reference to an element supplying
// a value to the current form.
type BoundFormValue struct {
	form  js.Value
	name  string
	radio bool
}

// NewBoundFormValue creates a BoundFormValue for the from at the given path.
// radio must be true iff the input with the given name has type=radio.
func NewBoundFormValue(d *ComponentData, name string, radio bool, path ...int) *BoundFormValue {
	return BoundFormValueAt(d.Walk(path...), name, radio)
}

// BoundFormValueAt returns a BoundFormValue for the given node and given form
// input name. radio hints at whether the input is a radio button.
func BoundFormValueAt(form js.Value, name string, radio bool) *BoundFormValue {
	return &BoundFormValue{form: form, name: name, radio: radio}
}

func (bfv *BoundFormValue) get() js.Value {
	if bfv.radio {
		list := bfv.form.Get("elements").Get(bfv.name)
		for i := 0; i < list.Length(); i++ {
			item := list.Index(i)
			if item.Get("checked").Bool() {
				return item.Get("value")
			}
		}
		return js.Value{}
	}
	return bfv.form.Get("elements").Get(bfv.name).Get("value")
}

func (bfv *BoundFormValue) set(value interface{}) {
	elm := bfv.form.Get("elements").Get(bfv.name)
	if bfv.radio {
		if str, ok := value.(string); ok {
			for i := 0; i < elm.Length(); i++ {
				item := elm.Index(i)
				if item.Get("value").String() == str {
					item.Set("checked", true)
					return
				}
			}
		} else if iv, ok := value.(int); ok {
			for i := 0; i < elm.Length(); i++ {
				item := elm.Index(i)
				if item.Get("value").Int() == iv {
					item.Set("checked", true)
					return
				}
			}
		} else {
			panic("unsupported value type for BoundFormValue on radio button!")
		}
		panic("unknown radio value!")
	}
	elm.Set("value", value)
}

// BoundEventValue implements BoundValue as a reference to a value of the
// captured event, or the event itself.
type BoundEventValue struct {
	val js.Value
}

// BoundEventValueAt returns a BoundEventValue for the given event.
// If propName is not empty, it binds the event's property with the given name,
// else it binds the event itself.
func BoundEventValueAt(e js.Value, propName string) *BoundEventValue {
	if propName == "" {
		return &BoundEventValue{val: e}
	}
	return &BoundEventValue{val: e.Get(propName)}
}

func (bev *BoundEventValue) get() js.Value {
	return bev.val
}

func (bev *BoundEventValue) set(value interface{}) {
	panic("BoundEvent doesn't support set()")
}

// BoundSelf implements BoundValue as a reference to a DOM node.
// It retrieves the linked node when getting its value, and replaces it with the
// given node when setting a value.
type BoundSelf struct {
	node js.Value
}

// NewBoundSelf creates a BoundSelf for the node at the given path.
func NewBoundSelf(d *ComponentData, path ...int) *BoundSelf {
	return &BoundSelf{node: d.Walk(path...)}
}

// BoundSelfAt returns a BoundSelf with the given node as target.
func BoundSelfAt(node js.Value) *BoundSelf {
	return &BoundSelf{node: node}
}

func (bs *BoundSelf) get() js.Value {
	return bs.node
}

func (bs *BoundSelf) set(value interface{}) {
	if o, ok := value.(js.Value); ok {
		bs.node.Get("parentNode").Call("replaceChild", o, bs.node)
		bs.node = o
	} else {
		node := js.Global().Get("document").Call("createTextNode", value)
		bs.node.Get("parentNode").Call("replaceChild", node, bs.node)
		bs.node = node
	}
}

// Assign is low-level assignment of a value to a bound value.
func Assign(bv BoundValue, value interface{}) {
	bv.set(value)
}
