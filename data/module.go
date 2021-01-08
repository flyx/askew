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
	Document         *html.Node
	JSFile, HTMLFile string
	VarName          *string
}

// RootNode returns the root node (<html>) of the file's HTML document
func (asf *ASiteFile) RootNode() *html.Node {
	// DocumentNode -[child]-> DoctypeNode -[sibling]-> html node
	return asf.Document.FirstChild.NextSibling
}

// Package describes a Go package with askew content in it.
type Package struct {
	// descriptors of all *.askew files inside the package
	Files []*AskewFile
	// if the directory contains a *.asite file, this is its descriptor.
	Site *ASiteFile
	// ImportPath can be used to import this package into other packages
	ImportPath string
	// Name is the package's name.
	Name string
}

// BaseDir describes the directory on which askew is executed
type BaseDir struct {
	// Package is a map tha maps relative paths to packages
	Packages map[string]*Package
	// ImportPath is the path with which the base directory can be imported.
	ImportPath string
}
