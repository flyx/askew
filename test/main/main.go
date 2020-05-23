package main

import (
	"github.com/flyx/tbc/test/generated/ui"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	body := js.Global.Get("document").Get("body")

	first := ui.NewNameForm()
	body.Call("appendChild", first.Root())

	first.Name.Set("First")
	first.Age.Set(42)

	second := ui.NewNameForm()
	body.Call("appendChild", second.Root())
	second.Name.Set("Second")
	second.Age.Set(23)
}
