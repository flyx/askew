package tdc

import "github.com/gopherjs/gopherjs/js"

// walkPath starts at root, which is assumed to be an HTML node, and for each
// path item, selects the child node with the index corresponding to that path
// item. At the end, returns the target node.
func walkPath(root *js.Object, path []int) *js.Object {
	cur := root.Index(path[0])
	for i := 1; i < len(path); i++ {
		cur = cur.Get("childNodes").Index(path[i])
	}
	return cur
}
