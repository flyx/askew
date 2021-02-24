package askew

import (
	"syscall/js"
)

// GenericList is a list of Components whose manipulation methods auto-update
// the corresponding nodes in the document.
type GenericList struct {
	mgr   ListManager
	items []Component
}

// Init initializes the list, discarding previous data.
// The list's items will be placed in the given container, starting at the
// given index.
func (l *GenericList) Init(container js.Value, index int) {
	l.mgr = CreateListManager(container, index)
	l.items = nil
}

// Len returns the number of items in the list.
func (l *GenericList) Len() int {
	return len(l.items)
}

// Item returns the item at the current index.
func (l *GenericList) Item(index int) Component {
	return l.items[index]
}

// Append appends the given item to the list.
func (l *GenericList) Append(item Component) {
	if item == nil {
		panic("cannot append nil to list")
	}
	l.items = append(l.items, item)
	l.mgr.Append(item)
	return
}

// Insert inserts the given item at the given index into the list.
func (l *GenericList) Insert(index int, item Component) {
	var prev js.Value
	if index < len(l.items) {
		prev = l.items[index].FirstNode()
	}
	if item == nil {
		panic("cannot insert nil into list")
	}
	l.items = append(l.items, nil)
	copy(l.items[index+1:], l.items[index:])
	l.items[index] = item
	l.mgr.Insert(item, prev)
	return
}

// Remove removes the item at the given index from the list and returns it.
func (l *GenericList) Remove(index int) Component {
	item := l.items[index]
	item.Extract()
	copy(l.items[index:], l.items[index+1:])
	l.items = l.items[:len(l.items)-1]
	return item
}

// Destroy destroys the item at the given index in the list.
func (l *GenericList) Destroy(index int) {
	l.items[index].Destroy()
	copy(l.items[index:], l.items[index+1:])
	l.items = l.items[:len(l.items)-1]
}

// DestroyAll destroys all items in the list and empties it.
func (l *GenericList) DestroyAll() {
	for _, item := range l.items {
		item.Destroy()
	}
	l.items = l.items[:0]
}

// DoUpdateParent calls the underlying list manager's UpdateParent.
// This is an implementation detail and should not be called from user code.
func (l *GenericList) DoUpdateParent(oldParent, newParent, newEnd js.Value) {
	l.mgr.UpdateParent(oldParent, newParent, newEnd)
}

// GenericOptional is a container that may optionally hold one arbitrary component.
type GenericOptional struct {
	mgr ListManager
	cur Component
}

// Init initializes the container to be empty.
// The contained item, if any, will be placed in the given container at the
// given index.
func (o *GenericOptional) Init(container js.Value, index int) {
	o.mgr = CreateListManager(container, index)
	o.cur = nil
}

// Item returns the current item, or nil if no item is assigned
func (o *GenericOptional) Item() Component {
	return o.cur
}

// Set sets the contained item destroying the current one.
// Give nil as value to simply destroy the current item.
func (o *GenericOptional) Set(value Component) {
	if o.cur != nil {
		o.cur.Destroy()
	}
	o.cur = value
	if value != nil {
		o.mgr.Append(value)
	}
}

// Remove removes the contained item and returns it.
// Returns nil if no item is currently contained.
func (o *GenericOptional) Remove() Component {
	if o.cur != nil {
		o.cur.Extract()
		ret := o.cur
		o.cur = nil
		return ret
	}
	return nil
}

// DoUpdateParent calls the underlying list manager's UpdateParent.
// This is an implementation detail and should not be called from user code.
func (o *GenericOptional) DoUpdateParent(oldParent, newParent, newEnd js.Value) {
	o.mgr.UpdateParent(oldParent, newParent, newEnd)
}
