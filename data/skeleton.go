package data

import "golang.org/x/net/html"

// Skeleton describes the main HTML file.
type Skeleton struct {
	Root   *html.Node
	Embeds []Embed
}
