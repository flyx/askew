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
