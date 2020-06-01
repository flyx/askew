package data

import "golang.org/x/net/html"

// Slot describes a <tbc:slot> inside a macro.
type Slot struct {
	Name string
	Node *html.Node
}

// Macro describes a <tbc:macro>.
type Macro struct {
	Slots       []Slot
	First, Last *html.Node
}
