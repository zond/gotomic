package gotomic

import (
	"fmt" 
	"sync/atomic"
	"bytes"
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
	Compare(Thing) int
}

type Thing interface{}

type node struct {
	value Thing
	next *nodeRef
	deleted bool
}
func (self *node) val() Thing {
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

type List struct {
	*nodeRef
	size int64
}
func NewList() *List {
	return &List{new(nodeRef), 0}
}
func (self *List) Push(t Thing) {
	self.nodeRef.push(t)
	atomic.AddInt64(&self.size, 1)
}
func (self *List) Pop() Thing {
	atomic.AddInt64(&self.size, -1)
	return self.nodeRef.pop()
}
func (self *List) Search(c Comparable) Thing {
	if hit := self.nodeRef.search(c); hit.node != nil {
		return hit.node.value
	}
	return nil
}
func (self *List) Inject(c Comparable) {
	self.nodeRef.inject(c)
	atomic.AddInt64(&self.size, 1)
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
func (self *nodeRef) toSlice() []Thing {
	rval := make([]Thing, 0)
	current := self.node()
	for current != nil {
		rval = append(rval, current.value)
		current = current.next.node()
	}
	return rval
}
func (self *nodeRef) pushBefore(t Thing, n *node) bool {
	if self.node() != n {
		return false
	}
	new_node := &node{t, &nodeRef{unsafe.Pointer(n)}, false}
	return atomic.CompareAndSwapPointer(&self.Pointer, new_node.next.Pointer, unsafe.Pointer(new_node))
}
func (self *nodeRef) push(c Thing) {
	for !self.pushBefore(c, self.node()) {}
}
/*
 * inject c into self either before the first matching value (c.Compare(value) == 0), before the first value
 * it should be before (c.Compare(value) < 0) or after the first value it should be after (c.Compare(value) > 0).
 */
func (self *nodeRef) inject(c Comparable) {
	for {
		hit := self.search(c)
		if hit.ref != nil {
			if hit.ref.pushBefore(c, hit.node) { break }
		} else if hit.rightRef != nil {
			if hit.rightRef.pushBefore(c, hit.rightNode) { break }
		} else if hit.leftRef != nil {
			if hit.leftNode.next.pushBefore(c, hit.rightNode) { break }
		} else {
			panic(fmt.Sprintf("Expected some kind of result from %#v.search(%v), but got %+v", self, c, hit))
		}
	}
}
/*
 * Verify that all Comparable values in this list are after values they should be after (c.Compare(last) >= 0).
 */
func (self *nodeRef) verify() error {
	node := self.node()
	if node == nil {
		return nil
	}
	last := node.val()
	node = node.next.node()
	var bad [][]Thing
	for node != nil {
		value := node.val()
		if comp, ok := value.(Comparable); ok {
			if comp.Compare(last) < 0 {
				bad = append(bad, []Thing{last,value})
			}
		}
		last = node.val()
		node = node.next.node()
	}
	if len(bad) == 0 {
		return nil
	}
	buffer := new(bytes.Buffer)
	for index, pair := range bad {
		fmt.Fprint(buffer, pair[0], ",", pair[1])
		if index < len(bad) - 1 {
			fmt.Fprint(buffer, "; ")
		}
	}
	return fmt.Errorf("%v is badly ordered. The following nodes are in the wrong order: %v", self, string(buffer.Bytes()));
	
}
/*
 * search for c in self.
 *
 * Will stop searching when finding nil or an element that should be after c (c.Compare(element) < 0).
 *
 * Will return a hit containing the last nodeRef and node before a match (if no match, the last nodeRef and node before
 * it stops searching), the nodeRef and node for the match (if a match) and the last nodeRef and node after the match
 * (if no match, the first nodeRef and node, or nil/nil if at the end of the list).
 */
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
func (self *nodeRef) pop() Thing {
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
