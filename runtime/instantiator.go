package runtime

import "syscall/js"

// InstantiateTemplate clones a template's content and returns it.
// The return value will always be a DOM DocumentFragment.
func InstantiateTemplate(tmpl js.Value) js.Value {
	return tmpl.Get("content").Call("cloneNode", true)
}

// InstantiateTemplateByID clones a template's content and returns it.
// the template is identified by its ID.
// The return value will always be a DOM DocumentFragment.
func InstantiateTemplateByID(id string) js.Value {
	document := js.Global().Get("document")
	return InstantiateTemplate(document.Call("getElementById", id))
}
