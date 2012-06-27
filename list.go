package gotomic

import (
	"fmt" 
	"sync/atomic"
	"unsafe"
)

type Comparable interface {
	Compare(thing) int
}

type thing interface{}

type node struct {
	value thing
	next *nodeRef
	deleted bool
}
func (self *node) val() thing {
	if self == nil {
		return nil
	}
	return self.value
}
func (self *node) String() string {
	deleted := ""
	if self.deleted {
		deleted = " (x)"
	}
	return fmt.Sprintf("%v%v -> %v", self.value, deleted, self.next)
}

type nodeRef struct {
	unsafe.Pointer
}
func (self *nodeRef) node() *node {
	current := (*node)(self.Pointer)
	next_ok := current
	for next_ok != nil && next_ok.deleted {
		next_ok = next_ok.next.node()
	}
	if current != next_ok {
		atomic.CompareAndSwapPointer(&self.Pointer, unsafe.Pointer(current), unsafe.Pointer(next_ok))
	}
	return next_ok
}
func (self *nodeRef) toSlice() []thing {
	var rval []thing
	current := self.node()
	for current != nil {
		rval = append(rval, current.value)
		current = current.next.node()
	}
	return rval
}
func (self *nodeRef) pushBefore(t thing, n *node) bool {
	if self.node() != n {
		return false
	}
	new_node := &node{t, &nodeRef{unsafe.Pointer(n)}, false}
	return atomic.CompareAndSwapPointer(&self.Pointer, unsafe.Pointer(n), unsafe.Pointer(new_node))
}
func (self *nodeRef) push(c thing) {
	for !self.pushBefore(c, self.node()) {}
}
func (self *nodeRef) inject(c Comparable) {
	for {
		b, m, a := self.search(c)
		if b == nil {
			if m == nil {
				if self.pushBefore(c, a) { break }
			} else {
				if self.pushBefore(c, m) { break }
			}
		} else {
			if m == nil {
				if b.next.pushBefore(c, a) { break }
			} else {
				if b.next.pushBefore(c, m) { break }
			}
		}
	}
}
func (self *nodeRef) search(c Comparable) (before, match, after *node) {
	before = nil
	match = self.node()
	after = nil
	for {
		if match == nil {
			return
		}
		after = match.next.node()
		switch cmp := c.Compare(match.value); {
		case cmp < 0:
			after = match
			match = nil
			return
		case cmp == 0:
			return
		}
		before = match
		match = match.next.node()
		after = nil
	}
	panic(fmt.Sprint("Unable to search for ", c, " in ", self))
}
func (self *nodeRef) pop() thing {
	old_node := self.node()
	if old_node == nil {
		return nil
	}
	deleted_node := &node{old_node.value, old_node.next, true}
	for !atomic.CompareAndSwapPointer(&self.Pointer, unsafe.Pointer(old_node), unsafe.Pointer(deleted_node)) {
		old_node = self.node()
		if old_node == nil {
			return nil 
		}
		deleted_node.value = old_node.value
		deleted_node.next = old_node.next
	}
	self.node()
	return old_node.value
}
func (self *nodeRef) String() string {
	return fmt.Sprint(self.node())
}
