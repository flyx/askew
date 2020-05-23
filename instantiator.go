package tdc

import "github.com/gopherjs/gopherjs/js"

// InstantiateTemplate clones a template's content and returns it.
func InstantiateTemplate(id string) *js.Object {
	document := js.Global.Get("document")
	elm := document.Call("getElementById", id)
	copied := document.Call("importNode", elm, true)
	return copied.Get("content")
}
