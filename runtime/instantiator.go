package runtime

import "github.com/gopherjs/gopherjs/js"

// InstantiateTemplate clones a template's content and returns it.
func InstantiateTemplate(tmpl *js.Object) *js.Object {
	document := js.Global.Get("document")
	copied := document.Call("importNode", tmpl, true)
	return copied.Get("content")
}

// InstantiateTemplateByID clones a template's content and returns it.
// the template is identified by its ID.
func InstantiateTemplateByID(id string) *js.Object {
	document := js.Global.Get("document")
	elm := document.Call("getElementById", id)
	copied := document.Call("importNode", elm, true)
	return copied.Get("content")
}
