package main

import (
	"github.com/flyx/tbc/test/generated/ui"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	body := js.Global.Get("document").Get("body")
	forms := ui.NewNameForms()
	forms.Insert(body, nil)
	first := forms.Forms.Append()
	first.Heading.Set("First Form")
	first.Name.Set("First")
	first.Age.Set(42)

	second := forms.Forms.Append()
	second.Heading.Set("Second Form")
	second.Name.Set("Second")
	second.Age.Set(23)

	mt := ui.NewMacroTest()
	mt.Insert(body, nil)
	mt.MonospaceTitle.Set(true)
	mt.A.Set("AAA")
	mt.B.Set("BBB")
}
