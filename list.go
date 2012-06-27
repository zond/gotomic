package gotomic

import (
	"fmt" 
	"sync/atomic"
	"unsafe"
)

type hit struct {
	leftRef *nodeRef
	leftNode *node
	ref *nodeRef
	node *node
	rightRef *nodeRef
	rightNode *node
}
func (self *hit) String() string {
	return fmt.Sprintf("&hit{%p(%v),%p(%v),%p(%v)}", self.leftRef, self.leftNode.val(), self.ref, self.node.val(), self.rightRef, self.rightNode.val())
}

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
	return atomic.CompareAndSwapPointer(&self.Pointer, new_node.next.Pointer, unsafe.Pointer(new_node))
}
func (self *nodeRef) push(c thing) {
	for !self.pushBefore(c, self.node()) {}
}
func (self *nodeRef) inject(c Comparable) {
	for {
		hit := self.search(c)
		if hit.ref != nil {
			if hit.ref.pushBefore(c, hit.node) { break }
		} else if hit.rightRef != nil {
			if hit.rightRef.pushBefore(c, hit.rightNode) { break }
		} else if hit.leftRef != nil {
			if hit.leftRef.pushBefore(c, hit.leftNode) { break }
		} else {
			panic(fmt.Sprintf("Expected some kind of result from %#v.search(%v), but got %+v", self, c, hit))
		}
	}
}
func (self *nodeRef) search(c Comparable) (rval *hit) {
	rval = &hit{nil, nil, self, self.node(), nil, nil}
	for {
		if rval.node == nil {
			return
		}
		rval.rightRef = rval.node.next
		rval.rightNode = rval.rightRef.node()
		switch cmp := c.Compare(rval.node.value); {
		case cmp < 0:
			rval.rightRef = rval.ref
			rval.rightNode = rval.node
			rval.ref = nil
			rval.node = nil
			return
		case cmp == 0:
			return
		}
		rval.leftRef = rval.ref
		rval.leftNode = rval.leftRef.node()
		rval.ref = rval.leftNode.next
		rval.node = rval.ref.node()
		rval.rightRef = nil
		rval.rightNode = nil
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
