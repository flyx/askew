package data

import "golang.org/x/net/html"

// Skeleton describes the main HTML file.
type Skeleton struct {
	EmbedHost
	Imports map[string]string
	Root    *html.Node
}
