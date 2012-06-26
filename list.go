package gotomic

import (
	"fmt" 
	"sync/atomic"
	"unsafe"
)

type Comparable interface {
	Compare(a interface{}) int
}

type node struct {
	value Comparable
	next *nodeRef
	deleted bool
}
func (self *node) String() string {
	if self.next.node() == nil {
		return fmt.Sprintf("%#v", self.value)
	}
	return fmt.Sprintf("%#v -> %v", self.value, self.next)
}

type nodeRef struct {
	unsafe.Pointer
}
func (self *nodeRef) node() *node {
	return (*node)(atomic.LoadPointer(&self.Pointer))
}
func (self *nodeRef) push(c Comparable) {
	old_node := self.node()
	new_node := &node{c, &nodeRef{unsafe.Pointer(old_node)}, false}
	if !atomic.CompareAndSwapPointer(&self.Pointer, unsafe.Pointer(old_node), unsafe.Pointer(new_node)) {
		self.push(c)
	}
}
func (self *nodeRef) String() string {
	if atomic.LoadPointer(&self.Pointer) == nil {
		return "<nil/>"
	}
	return fmt.Sprint(self.node())
}
