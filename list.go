package gotomic

import (
	"fmt" 
	"sync/atomic"
	"bytes"
	"unsafe"
)

type hit struct {
	left *element
	element *element
	right *element
}
func (self *hit) String() string {
	return fmt.Sprintf("&hit{%v,%v,%v}", self.left.val(), self.element.val(), self.right.val())
}

/*
 Comparable types can be kept sorted in a List.
 */
type Comparable interface {
	Compare(Thing) int
}

type Thing interface{}

var list_head = "LIST_HEAD"

/*
 List is a singly linked list based on "A Pragmatic Implementation of Non-Blocking Linked-Lists" by Timothy L. Harris <http://www.timharris.co.uk/papers/2001-disc.pdf>
 
 It is thread safe and non-blocking, and supports ordered elements by using List#inject with values implementing Comparable.
 */
type List struct {
	*element
	size int64
}
func NewList() *List {
	return &List{&element{nil, &list_head}, 0}
}
/*
 Push adds t to the top of the List.
 */
func (self *List) Push(t Thing) {
	self.element.add(t)
	atomic.AddInt64(&self.size, 1)
}
/*
 Pop removes and returns the top of the List.
 */
func (self *List) Pop() (rval Thing, ok bool) {
	if rval, ok := self.element.remove(); ok {
		atomic.AddInt64(&self.size, -1)
		return rval, true
	}
	return nil, false
}
func (self *List) String() string {
	return fmt.Sprint(self.ToSlice())
}
/*
 ToSlice returns a []Thing that is logically identical to the List.
 */
func (self *List) ToSlice() []Thing {
	return self.element.next().ToSlice()
}
/*
 Search return the first element in the list that matches c (c.Compare(element) == 0)
 */
func (self *List) Search(c Comparable) Thing {
	if hit := self.element.search(c); hit.element != nil {
		return hit.element.val()
	}
	return nil
}
func (self *List) Size() int {
	return int(self.size)
}
/*
 Inject c into the List at the first place where it is <= to all elements before it.
 */
func (self *List) Inject(c Comparable) {
	self.element.inject(c)
	atomic.AddInt64(&self.size, 1)
}


/*
 Note that if the pointer has the deleted flag set, it is the element containing the pointer that is deleted!
 */
func isDeleted(p unsafe.Pointer) bool {
	return uintptr(p) & 1 == 1
}
func deleted(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) | 1)
}
func normal(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) &^ 1)
}

type element struct {
	/*
	 The next element in the list. If this pointer has the deleted flag set it means THIS element, not the next one, is deleted.
	 */
	unsafe.Pointer
	value Thing
}
func (self *element) next() *element {
	next := atomic.LoadPointer(&self.Pointer)
	for next != nil {
		/*
		 If the pointer of the next element is marked as deleted, that means the next element is supposed to be GONE
		 */
		if nextPointer := atomic.LoadPointer(&(*element)(normal(next)).Pointer); isDeleted(nextPointer) {
			/*
			 If OUR pointer is marked as deleted, that means WE are supposed to be gone
			 */
			if isDeleted(next) {
				/*
				 .. which means that we can steal the pointer of the next element right away,
				 it points to the right place AND it is marked as deleted.
				 */
				atomic.CompareAndSwapPointer(&self.Pointer, next, nextPointer)
			} else {
				/*
				 .. if not, we have to remove the marking on the pointer before we steal it.
				 */
				atomic.CompareAndSwapPointer(&self.Pointer, next, normal(nextPointer))
			}
			next = atomic.LoadPointer(&self.Pointer)
		} else {
			/*
			 If the next element is NOT deleted, then we simply return a pointer to it, and make
			 damn sure that the pointer is a working one even if we are deleted (and, therefore,
			 our pointer is marked as deleted).
			 */
			return (*element)(normal(next))
		}
	}
	return nil
}
func (self *element) val() Thing {
	if self == nil {
		return nil
	}
	return self.value
}
func (self *element) String() string {
	return fmt.Sprint(self.ToSlice())
}
func (self *element) Describe() string {
	if self == nil {
		return fmt.Sprint(nil)
	}
	deleted := ""
	if isDeleted(self.Pointer) {
		deleted = " (x)"
	}
	return fmt.Sprintf("%#v%v -> %v", self, deleted, self.next().Describe())
}
func (self *element) add(c Thing) {
	alloc := &element{}
	for !self.addBefore(c, alloc, self.next()) {}
}
func (self *element) addBefore(t Thing, allocatedElement, before *element) bool {
	if self.next() != before {
		return false
	}
	allocatedElement.value = t
	allocatedElement.Pointer = unsafe.Pointer(before)
	newPointer := unsafe.Pointer(allocatedElement)
	if isDeleted(self.Pointer) {
		newPointer = deleted(newPointer)
	}
	return atomic.CompareAndSwapPointer(&self.Pointer, unsafe.Pointer(before), newPointer)
}
/*
 inject c into self either before the first matching value (c.Compare(value) == 0), before the first value
 it should be before (c.Compare(value) < 0) or after the first value it should be after (c.Compare(value) > 0).
 */
func (self *element) inject(c Comparable) {
	alloc := &element{}
	for {
		hit := self.search(c)
		if hit.left != nil {
			if hit.element != nil {
				if hit.left.addBefore(c, alloc, hit.element) { break }
			} else {
				if hit.left.addBefore(c, alloc, hit.right) { break }
			}
		} else if hit.element != nil {
			if hit.element.addBefore(c, alloc, hit.right) { break }
		} else {
			panic(fmt.Errorf("Unable to inject %v properly into %v, it ought to be first but was injected into the first element of the list!", c, self))
		}
	}
}
func (self *element) ToSlice() []Thing {
	rval := make([]Thing, 0)
	current := self
	for current != nil {
		rval = append(rval, current.value)
		current = current.next()
	}
	return rval
}
/*
 search for c in self.
 
 Will stop searching when finding nil or an element that should be after c (c.Compare(element) < 0).
 
 Will return a hit containing the last elementRef and element before a match (if no match, the last elementRef and element before
 it stops searching), the elementRef and element for the match (if a match) and the last elementRef and element after the match
 (if no match, the first elementRef and element, or nil/nil if at the end of the list).
 */
func (self *element) search(c Comparable) (rval *hit) {
	rval = &hit{nil, self, nil}
	for {
		if rval.element == nil {
			return
		}
		rval.right = rval.element.next()
		if rval.element.value != &list_head {
			switch cmp := c.Compare(rval.element.value); {
			case cmp < 0:
				rval.right = rval.element
				rval.element = nil
				return
			case cmp == 0:
				return
			}
		}
		rval.left = rval.element
		rval.element = rval.left.next()
		rval.right = nil
	}
	panic(fmt.Sprint("Unable to search for ", c, " in ", self))
}
/*
 Verify that all Comparable values in this list are after values they should be after (c.Compare(last) >= 0).
 */
func (self *element) verify() (err error) {
	current := self
	var last Thing
	var bad [][]Thing
	seen := make(map[*element]bool)
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
	return fmt.Errorf("%v is badly ordered. The following elements are in the wrong order: %v", self, string(buffer.Bytes()));
	
}
func (self *element) doRemove() bool {
	ptr := self.Pointer
	return atomic.CompareAndSwapPointer(&self.Pointer, normal(ptr), deleted(ptr))
}
func (self *element) remove() (rval Thing, ok bool) {
	n := self.next()
	for n != nil && !n.doRemove() {
		n = self.next()
	}
	if n != nil {
		return n.value, true
	}
	return nil, false
}
