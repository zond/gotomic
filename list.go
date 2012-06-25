package gotomic

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"unsafe"
)

type thing interface{}

type element struct {
	value   thing
	deleted int8
	next    unsafe.Pointer
}
func (self *element) unsafe() unsafe.Pointer {
	return unsafe.Pointer(self)
}

type iterator struct {
	list *List
	last *element
	next *element
}

func newIterator(l *List) *iterator {
	rval := &iterator{l, nil, (*element)(l.head)}
	rval.ff()
	return rval
}
func (self *iterator) Delete() {
	
}
func (self *iterator) ff() {
	for {
		if self.next == nil {
			return
		}
		if self.next.deleted != 0 {
			if unsafe.Pointer(self.next) == self.list.head {
				if atomic.CompareAndSwapPointer(&(self.list.head), self.next.unsafe(), self.next.next) {
					atomic.AddInt64(&(self.list.size), -1)
				}
			} else {
				if atomic.CompareAndSwapPointer(&(self.last.next), self.next.unsafe(), self.next.next) {
					atomic.AddInt64(&(self.list.size), -1)
				}
			}
			self.next = (*element)(self.next.next)
		} else {
			return
		}
	}
}
func (self *iterator) HasNext() bool {
	return self.next != nil
}
func (self *iterator) nextAny() thing {
	if self.HasNext() {
		el := self.next
		self.next = (*element)(el.next)
		return el.value
	}
	return nil
}
func (self *iterator) Next() thing {
	rval := self.nextAny()
	self.ff()
	return rval
}

type List struct {
	head unsafe.Pointer
	size int64
}

func NewList() *List {
	return &List{}
}
func (self *List) String() string {
	buffer := bytes.NewBufferString(fmt.Sprintf("<List/%v %p>", self.size, self))
	ptr := self.head
	for ptr != nil {
		el := (*element)(ptr)
		fmt.Fprint(buffer, el.value)
		ptr = el.next
		if ptr != nil {
			fmt.Fprint(buffer, ", ")
		}
	}
	fmt.Fprint(buffer, "</List>")
	return string(buffer.Bytes())
}
func (self *List) Size() int {
	return int(self.size)
}
func (self *List) Iterator() *iterator {
	return newIterator(self)
}
func (self *List) Push(t thing) {
	new_head := &element{t, 0, self.head}
	if atomic.CompareAndSwapPointer(&(self.head), new_head.next, new_head.unsafe()) {
		atomic.AddInt64(&(self.size), 1)
		return
	}
	self.Push(t)
}
func (self *List) Pop() thing {
	head := (*element)(self.head)
	if head == nil {
		return nil
	}
	if atomic.CompareAndSwapPointer(&(self.head), self.head, head.next) {
		atomic.AddInt64(&(self.size), -1)
		return head.value
	}
	return self.Pop()
}
