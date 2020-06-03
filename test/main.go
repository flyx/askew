package main

import (
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
	first := Forms.Forms.Append(nil)
	first.Heading.Set("First Form")
	first.Name.Set("First")
	first.Age.Set(42)
	first.Controller = &handler{}

	second := Forms.Forms.Append(nil)
	second.Heading.Set("Second Form")
	second.Name.Set("Second")
	second.Age.Set(23)

	Test.Content.MonospaceTitle.Set(true)
	Test.Content.A.Set("AAA")
	Test.Content.B.Set("BBB")
}
