package gotomic

import (
	"fmt" 
	"sync/atomic"
	"bytes"
	"unsafe"
)

type hit struct {
	left *node
	node *node
	right *node
}
func (self *hit) String() string {
	return fmt.Sprintf("&hit{%v,%v,%v}", self.left.val(), self.node.val(), self.right.val())
}

/*
 * Comparable types can be kept sorted in a List.
 */
type Comparable interface {
	Compare(Thing) int
}

type Thing interface{}

var list_head = "LIST_HEAD"

/*
 * List is a singly linked list based on "A Pragmatic Implementation of Non-Blocking Linked-Lists" by Timothy L. Harris <http://www.timharris.co.uk/papers/2001-disc.pdf>
 *
 * It is thread safe and non-blocking, and supports ordered elements by using List#inject with values implementing Comparable.
 */
type List struct {
	*node
	size int64
}
func NewList() *List {
	return &List{&node{nil, &list_head}, 0}
}
/*
 * Push adds t to the top of the List.
 */
func (self *List) Push(t Thing) {
	self.node.add(t)
	atomic.AddInt64(&self.size, 1)
}
/*
 * Pop removes and returns the top of the List.
 */
func (self *List) Pop() (rval Thing, ok bool) {
	if rval, ok := self.node.remove(); ok {
		atomic.AddInt64(&self.size, -1)
		return rval, true
	}
	return nil, false
}
func (self *List) String() string {
	return fmt.Sprint(self.ToSlice())
}
/*
 * ToSlice returns a []Thing that is logically identical to the List.
 */
func (self *List) ToSlice() []Thing {
	return self.node.next().ToSlice()
}
/*
 * Search return the first element in the list that matches c (c.Compare(element) == 0)
 */
func (self *List) Search(c Comparable) Thing {
	if hit := self.node.search(c); hit.node != nil {
		return hit.node.val()
	}
	return nil
}
func (self *List) Size() int {
	return int(self.size)
}
/*
 * Inject c into the List at the first place where it is <= to all elements before it.
 */
func (self *List) Inject(c Comparable) {
	self.node.inject(c)
	atomic.AddInt64(&self.size, 1)
}



func isDeleted(p unsafe.Pointer) bool {
	return uintptr(p) & 1 == 1
}
func deleted(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) | 1)
}
func normal(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) &^ 1)
}

type node struct {
	unsafe.Pointer
	value Thing
}
func (self *node) next() *node {
	next := atomic.LoadPointer(&self.Pointer)
	for next != nil {
		if nextPointer := atomic.LoadPointer(&(*node)(normal(next)).Pointer); isDeleted(nextPointer) {
			if isDeleted(next) {
				atomic.CompareAndSwapPointer(&self.Pointer, next, nextPointer)
			} else {
				atomic.CompareAndSwapPointer(&self.Pointer, next, normal(nextPointer))
			}
			next = atomic.LoadPointer(&self.Pointer)
		} else {
			return (*node)(normal(next))
		}
	}
	return nil
}
func (self *node) val() Thing {
	if self == nil {
		return nil
	}
	return self.value
}
func (self *node) String() string {
	return fmt.Sprint(self.ToSlice())
}
func (self *node) Describe() string {
	if self == nil {
		return fmt.Sprint(nil)
	}
	deleted := ""
	if isDeleted(self.Pointer) {
		deleted = " (x)"
	}
	return fmt.Sprintf("%#v%v -> %v", self, deleted, self.next().Describe())
}
func (self *node) add(c Thing) {
	alloc := &node{}
	for !self.addBefore(c, alloc, self.next()) {}
}
func (self *node) addBefore(t Thing, allocatedNode, before *node) bool {
	if self.next() != before {
		return false
	}
	allocatedNode.value = t
	allocatedNode.Pointer = unsafe.Pointer(before)
	newPointer := unsafe.Pointer(allocatedNode)
	if isDeleted(self.Pointer) {
		newPointer = deleted(newPointer)
	}
	return atomic.CompareAndSwapPointer(&self.Pointer, unsafe.Pointer(before), newPointer)
}
/*
 * inject c into self either before the first matching value (c.Compare(value) == 0), before the first value
 * it should be before (c.Compare(value) < 0) or after the first value it should be after (c.Compare(value) > 0).
 */
func (self *node) inject(c Comparable) {
	alloc := &node{}
	for {
		hit := self.search(c)
		if hit.left != nil {
			if hit.node != nil {
				if hit.left.addBefore(c, alloc, hit.node) { break }
			} else {
				if hit.left.addBefore(c, alloc, hit.right) { break }
			}
		} else if hit.node != nil {
			if hit.node.addBefore(c, alloc, hit.right) { break }
		} else {
			panic(fmt.Errorf("Unable to inject %v properly into %v, it ought to be first but was injected into the first node of the list!", c, self))
		}
	}
}
func (self *node) ToSlice() []Thing {
	rval := make([]Thing, 0)
	current := self
	for current != nil {
		rval = append(rval, current.value)
		current = current.next()
	}
	return rval
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
func (self *node) search(c Comparable) (rval *hit) {
	rval = &hit{nil, self, nil}
	for {
		if rval.node == nil {
			return
		}
		rval.right = rval.node.next()
		if rval.node.value != &list_head {
			switch cmp := c.Compare(rval.node.value); {
			case cmp < 0:
				rval.right = rval.node
				rval.node = nil
				return
			case cmp == 0:
				return
			}
		}
		rval.left = rval.node
		rval.node = rval.left.next()
		rval.right = nil
	}
	panic(fmt.Sprint("Unable to search for ", c, " in ", self))
}
/*
 * Verify that all Comparable values in this list are after values they should be after (c.Compare(last) >= 0).
 */
func (self *node) verify() (err error) {
	current := self
	var last Thing
	var bad [][]Thing
	seen := make(map[*node]bool)
	for current != nil {
		if _, ok := seen[current]; ok {
			return fmt.Errorf("%#v is circular!", self)
		}
		value := current.value
		if last != &list_head {
			if comp, ok := value.(Comparable); ok {
				if comp.Compare(last) < 0 {
					bad = append(bad, []Thing{last,value})
				}
			}
		}
		seen[current] = true
		last = value
		current = current.next()
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
func (self *node) doRemove() bool {
	ptr := self.Pointer
	return atomic.CompareAndSwapPointer(&self.Pointer, normal(ptr), deleted(ptr))
}
func (self *node) remove() (rval Thing, ok bool) {
	n := self.next()
	for n != nil && !n.doRemove() {
		n = self.next()
	}
	if n != nil {
		return n.value, true
	}
	return nil, false
}
