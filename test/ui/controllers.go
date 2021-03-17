package ui

import (
	"fmt"
	"strconv"

	"syscall/js"
)

func (o *row) foo() {}

// SetTitle is called when input A changes.
func (o *MacroTest) SetTitle(value string) bool {
	o.Title.Set(value)
	return true
}

// RandomizeText is called when the Randomize button is clicked.
func (o *MacroTest) RandomizeText() {
	o.TextContent.Set(fmt.Sprintf("%f", js.Global().Get("Math").Call("random").Float()))
}

func (o *Herp) click() {
	o.count++
	go func() {
		js.Global().Call("alert", "Derp"+strconv.Itoa(o.count))
	}()
}

func (o *OneTwoThree) click(caption string) {
	go func() {
		js.Global().Call("alert", caption)
	}()
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
	go func() {
		js.Global().Call("alert", o.Button.Get().Get("dataset").Get("foo"))
	}()
}

func (o *AutoFieldTest) click() {
	go func() {
		js.Global().Call("alert", o.content)
	}()
}
