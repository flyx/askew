package runtime

import "github.com/gopherjs/gopherjs/js"

// ComponentData holds the content of an instance of a <a:component>.
//
// Initially, the data is contained in a DocumentFragment which resulted from instantiating a <template>.
// On InsertInto, that DocumentFragment is emptied as the nodes are moved into the given parent.
//
// After insertion, ComponentData keeps track of its content via a pointer to the first and the last node.
// This requires that those nodes can never change. Since they are typically TextNodes of the whitespace after
// <a:component> / before </a:component>, that is not a problem.
// This should eventually be enforced by the parser.
//
// Extracting the component moves the nodes from their current parent back into the DocumentFragment so that
// it can be inserted somewhere else. Extracting and re-inserting seems to be a rather exotic use-case, but
// can happen with components that are part of a list.
//
// The interface of this type is consumed by <a:embed>; the user does not need to use it directly if
// the website is defined with a skeleton.
type ComponentData struct {
	fragment, first, last *js.Object
}

// Init initializes the ComponentData with the given DocumentFragment node.
// Previous data is discarded. The Component will be in initial state afterwards.
func (cd *ComponentData) Init(frag *js.Object) {
	cd.fragment, cd.first, cd.last = frag, nil, nil

}

// DoInsert inserts the component into the given parent before the node before or at the end if before is nil.
// The ComponentData must be in initial state and transitions into inserted state.
//
// This is the backend for a Component's InsertInto, but the component may need to do additional things depending on its embeds.
func (cd *ComponentData) DoInsert(parent *js.Object, before *js.Object) {
	if cd.first != nil {
		panic("DoInsert called on ComponentData that is already in inserted state")
	}
	cd.first = cd.fragment.Get("firstChild")
	cd.last = cd.fragment.Get("lastChild")
	parent.Call("insertBefore", cd.fragment, before)
}

// DoExtract removes the component from its current parent.
// The ComponentData must be in inserted state and transitions to initial state.
//
// This is the backend for a Component's Extract, but the component may need to do additional things depending on its embeds.
func (cd *ComponentData) DoExtract() {
	if cd.first == nil {
		panic("Extract called on ComponentData that is in initial state")
	}
	cur := cd.first
	parent := cur.Get("parentNode")
	for {
		next := cur.Get("nextSibling")
		cd.fragment.Call("appendChild", parent.Call("removeChild", cur))
		if cur == cd.last {
			break
		}
		cur = next
	}
	cd.first, cd.last = nil, nil
}

// Walk descends into the DocumentFragment's children using the given list of indexes.
// This may only be done when the ComponentData is in initial state.
func (cd *ComponentData) Walk(path ...int) *js.Object {
	if cd.first != nil {
		panic("Walk called on ComponentData that is already inserted")
	}
	return WalkPath(cd.fragment, path...)
}

// First returns the first DOM node in this component
func (cd *ComponentData) First() *js.Object {
	if cd.first == nil {
		return cd.fragment.Get("firstChild")
	}
	return cd.first
}

// DocumentFragment returns the DocumentFragment the component uses to store its contents
// when it is in initial state.
func (cd *ComponentData) DocumentFragment() *js.Object {
	return cd.fragment
}

// Component is implemented by every type generated from <a:component>.
type Component interface {
	Data() *ComponentData
	InsertInto(parent *js.Object, before *js.Object)
	Extract()
}
