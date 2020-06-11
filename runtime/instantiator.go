package runtime

import "github.com/gopherjs/gopherjs/js"

// InstantiateTemplate clones a template's content and returns it.
// The return value will always be a DOM DocumentFragment.
func InstantiateTemplate(tmpl *js.Object) *js.Object {
	return tmpl.Get("content").Call("cloneNode", true)
}

// InstantiateTemplateByID clones a template's content and returns it.
// the template is identified by its ID.
// The return value will always be a DOM DocumentFragment.
func InstantiateTemplateByID(id string) *js.Object {
	document := js.Global.Get("document")
	return InstantiateTemplate(document.Call("getElementById", id))
}
