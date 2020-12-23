package data

import (
	"github.com/flyx/net/html"
)

// File describes an .askew or .asite file.
type File struct {
	// map of alias -> relative package path.
	// package path is relative to module.
	// alias is by default the part of the package path after the last '.'
	Imports  map[string]string
	BaseName string
	Path     string
}

// AskewFile describes an .askew file.
type AskewFile struct {
	File
	Content    []*html.Node
	Macros     map[string]Macro
	Components map[string]*Component
}

// ASiteFile describes an .asite file.
type ASiteFile struct {
	File
	Unit
	Document, Descriptor *html.Node
	JSFile, HTMLFile     string
	VarName              *string
}

// RootNode returns the root node (<html>) of the file's HTML document
func (asf *ASiteFile) RootNode() *html.Node {
	// DocumentNode -[child]-> DoctypeNode -[sibling]-> html node
	return asf.Document.FirstChild.NextSibling
}

// Package describes a Go package.
type Package struct {
	Files []*AskewFile
	Site  *ASiteFile
	Path  string
}

// BaseDir describes the directory on which askew is executed
type BaseDir struct {
	Packages   map[string]*Package
	ImportPath string
}
