package runtime

import "syscall/js"

// WalkPath starts at root, which is assumed to be an HTML node, and for each
// path item, selects the child node with the index corresponding to that path
// item. At the end, returns the target node.
func WalkPath(root js.Value, path ...int) js.Value {
	cur := root
	for _, i := range path {
		cur = cur.Get("childNodes").Index(i)
	}
	return cur
}
