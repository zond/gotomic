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
	deleted int32
	next    unsafe.Pointer
}
func (self *element) unsafe() unsafe.Pointer {
	return unsafe.Pointer(self)
}

type iterator struct {
	list *List
	last unsafe.Pointer
	current unsafe.Pointer
	next unsafe.Pointer
}
func newIterator(l *List) *iterator {
	rval := &iterator{l, nil, nil, l.head}
	rval.ff()
	return rval
}
func (self *iterator) Delete() {
	if self.current == nil {
		return
	}
	atomic.StoreInt32(&((*element)(self.current).deleted), 1) 
	if self.last == nil {
		if atomic.CompareAndSwapPointer(&(self.list.head), self.current, self.next) {
			atomic.AddInt64(&(self.list.size), -1)
		}
	} else {
		if atomic.CompareAndSwapPointer(&((*element)(self.last).next), self.current, self.next) {
			atomic.AddInt64(&(self.list.size), -1)
		}
	}
}
func (self *iterator) ff() {
	for {
		if self.next == nil {
			return
		}
		if (*element)(atomic.LoadPointer(&(self.next))).deleted == 0 {
			return 
		}
		self.step()
		self.Delete()
	}
}
func (self *iterator) step() {
	self.last = self.current
	self.current = self.next
	self.next = (*element)(self.next).next
}
func (self *iterator) HasNext() bool {
	return self.next != nil
}
func (self *iterator) nextAny() thing {
	if self.HasNext() {
		el := (*element)(self.next)
		self.step()
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
	if self.head == nil {
		return nil
	}
	head := self.head
	if atomic.CompareAndSwapPointer(&(self.head), head, (*element)(head).next) {
		atomic.AddInt64(&(self.size), -1)
		return (*element)(head).value
	}
	return self.Pop()
}
