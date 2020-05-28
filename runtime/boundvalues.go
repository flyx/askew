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
	root *js.Object, pName string, path ...int) *BoundProperty {
	return &BoundProperty{
		node: WalkPath(root, path...), pName: pName}
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

// BoundAttribute implements BoundValue for an attribute of an HTML node.
type BoundAttribute struct {
	node  *js.Object
	aName string
}

// NewBoundAttribute creates a BoundAttribtue for the node found at the given
// path (relative to root) and the given attribute name.
func NewBoundAttribute(
	root *js.Object, aName string, path ...int) *BoundAttribute {
	return &BoundAttribute{node: WalkPath(root, path...), aName: aName}
}

// Init initializes the bound attribute with the given node and attribute name.
func (ba *BoundAttribute) Init(node *js.Object, aName string) {
	ba.node, ba.aName = node, aName
}

func (ba *BoundAttribute) get() *js.Object {
	return ba.node.Call("getAttribute", ba.aName)
}

func (ba *BoundAttribute) set(value interface{}) {
	ba.node.Call("setAttribute", ba.aName, value)
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
func NewBoundClass(root *js.Object, className string, path ...int) *BoundClass {
	return &BoundClass{
		node: WalkPath(root, path...), className: className}
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
