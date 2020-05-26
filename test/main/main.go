package main

import (
	"github.com/flyx/tbc/test/generated/ui"
	"github.com/gopherjs/gopherjs/js"
)

type handler struct{}

func (*handler) Reset(foo string) bool {
	js.Global.Call("alert", "reset: "+foo)
	return true
}

func (*handler) Submit() bool {
	js.Global.Call("alert", "submit!")
	return true
}

func main() {
	body := js.Global.Get("document").Get("body")
	forms := ui.NewNameForms()
	forms.InsertInto(body, nil)
	first := forms.Forms.Append()
	first.Heading.Set("First Form")
	first.Name.Set("First")
	first.Age.Set(42)
	first.SetController(&handler{})

	second := forms.Forms.Append()
	second.Heading.Set("Second Form")
	second.Name.Set("Second")
	second.Age.Set(23)

	et := ui.NewEmbedTest()
	et.InsertInto(body, nil)
	et.Content.MonospaceTitle.Set(true)
	et.Content.A.Set("AAA")
	et.Content.B.Set("BBB")
}
