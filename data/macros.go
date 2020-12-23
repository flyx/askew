package data

import "github.com/flyx/net/html"

// Slot describes an <a:slot> inside a macro.
type Slot struct {
	Name string
	Node *html.Node
}

// Macro describes an <a:macro>.
type Macro struct {
	Slots       []Slot
	First, Last *html.Node
}
