package main

import (
	"strconv"

	"github.com/flyx/askew/test/ui"

	"github.com/gopherjs/gopherjs/js"
)

type handler struct{}

func (*handler) Reset(foo string) bool {
	js.Global.Call("alert", "reset: "+foo)
	return true
}

func (*handler) Submit(name string, age int) bool {
	js.Global.Call("alert", "name="+name+", age="+strconv.FormatInt(int64(age), 10))
	return true
}

func main() {
	first := ui.NewNameForm(1)
	first.Heading.Set("First Form")
	first.Name.Set("First")
	first.Age.Set(42)
	first.Controller = &handler{}
	Forms.Forms.Append(first)

	second := ui.NewNameForm(2)
	second.Heading.Set("Second Form")
	second.Name.Set("Second")
	second.Age.Set(23)
	Forms.Forms.Append(second)

	Test.Content.MonospaceTitle.Set(true)
	Test.Content.A.Set("AAA")
	Test.Content.B.Set("BBB")
}
