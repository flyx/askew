package ui

import (
	"strconv"

	"syscall/js"
)

// SetTitle is called when input A changes.
func (o *MacroTest) SetTitle(value string) bool {
	o.Title.Set(value)
	return true
}

// RandomizeText is called when the Randomize button is clicked.
func (o *MacroTest) RandomizeText() {
	o.TextContent.Set(js.Global().Get("Math").Call("random").String())
}

func (o *Herp) click() {
	o.count++
	js.Global().Call("alert", "Derp"+strconv.Itoa(o.count))
}

func (o *OneTwoThree) click(caption string) {
	js.Global().Call("alert", caption)
}

func (o *EventTest) click(e js.Value) {
	button := e.Get("currentTarget")
	button.Set("innerText", button.Get("innerText").String()+"!")
}

func (o *ColorShuffler) click() {
	o.Color.Set((o.Color.Get() + 1) % 4)
}

func (o *ColorChooserByText) click(value string) {
	o.BgColor.Set(value)
}

func (o *SelfTest) click() {
	js.Global().Call("alert", o.Button.Get().Get("dataset").Get("foo"))
}

func (o *AutoFieldTest) click() {
	js.Global().Call("alert", o.content)
}
