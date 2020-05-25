package main

import (
	"github.com/flyx/tbc/test/generated/ui"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	body := js.Global.Get("document").Get("body")

	first := ui.NewNameForm()
	body.Call("appendChild", first.Root())
	first.Heading.Set("First Form")
	first.Name.Set("First")
	first.Age.Set(42)

	second := ui.NewNameForm()
	body.Call("appendChild", second.Root())
	second.Heading.Set("Second Form")
	second.Name.Set("Second")
	second.Age.Set(23)

	mt := ui.NewMacroTest()
	body.Call("appendChild", mt.Root())
	mt.MonospaceTitle.Set(true)
	mt.A.Set("AAA")
	mt.B.Set("BBB")
}
