package data

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// BodyEnv is a dummy body node to be used for fragment parsing
var BodyEnv = html.Node{
	Type:     html.ElementNode,
	Data:     "body",
	DataAtom: atom.Body}
