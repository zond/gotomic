package gotomic

import (
	"fmt" 
	"sync/atomic"
	"unsafe"
)

type thing interface{}

type node struct {
	value thing
	next *nodeRef
	deleted bool
}
func (self *node) String() string {
	deleted := ""
	if self.deleted {
		deleted = " (x)"
	}
	return fmt.Sprintf("%#v%v -> %v", self.value, deleted, self.next)
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
func (self *nodeRef) push(c thing) {
	old_node := self.node()
	new_node := &node{c, &nodeRef{unsafe.Pointer(old_node)}, false}
	for !atomic.CompareAndSwapPointer(&self.Pointer, unsafe.Pointer(old_node), unsafe.Pointer(new_node)) {
		old_node = self.node()
		new_node.next.Pointer = unsafe.Pointer(old_node)
	}
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
