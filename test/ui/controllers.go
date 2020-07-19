package ui

import "github.com/gopherjs/gopherjs/js"

// SetTitle is called when input A changes.
func (o *MacroTest) SetTitle(value string) bool {
	o.Title.Set(value)
	return true
}

func (o *Herp) click() {
	js.Global.Call("alert", "Derp")
}
