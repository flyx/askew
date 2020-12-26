package ui

import (
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

// SetTitle is called when input A changes.
func (o *MacroTest) SetTitle(value string) bool {
	o.Title.Set(value)
	return true
}

func (o *Herp) click() {
	o.count++
	js.Global.Call("alert", "Derp"+strconv.Itoa(o.count))
}

func (o *OneTwoThree) click(caption string) {
	js.Global.Call("alert", caption)
}

func (o *EventTest) click(e *js.Object) {
	button := e.Get("currentTarget")
	button.Set("innerText", button.Get("innerText").String()+"!")
}

func (o *ColorShuffler) click() {
	o.Color.Set((o.Color.Get() + 1) % 4)
}

func (o *ColorChooserByText) click(value string) {
	o.BgColor.Set(value)
}
