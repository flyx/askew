package main

import (
	"strconv"

	askew "github.com/flyx/askew/runtime"

	"github.com/flyx/askew/test/ui"

	"syscall/js"
)

type handler struct{}

func (*handler) Reset(foo string) bool {
	go func() {
		js.Global().Call("alert", "reset: "+foo)
	}()
	return true
}

func (*handler) Submit(name string, age int) {
	go func() {
		js.Global().Call("alert", "name="+name+", age="+strconv.FormatInt(int64(age), 10))
	}()
}

func main() {
	first := ui.NewNameForm(1)
	first.Heading.Set("First Form")
	first.Name.Set("First")
	first.Age.Set(42)
	Forms.Forms.Append(first)
	first.Controller = &handler{}

	second := ui.NewNameForm(2)
	second.Heading.Set("Second Form")
	second.Name.Set("Second")
	second.Age.Set(23)
	Forms.Forms.Append(second)

	Test.Content.MonospaceTitle.Set(true)
	Test.Content.A.Set("AAA")
	Test.Content.B.Set("BBB")

	Derp.Set(ui.NewHerp())
	Anything.Set(ui.NewHerp())

	askew.KeepAlive()
}
