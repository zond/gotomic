
package gotomic

import (
	"unsafe"
	"bytes"
	"sync/atomic"
	"fmt"
)

type thing interface{}

type element struct {
	value thing
	next unsafe.Pointer
}

type iterator struct {
	next unsafe.Pointer
}
func (self *iterator) HasNext() bool {
	return self.next != nil
}
func (self *iterator) Next() thing {
	if self.HasNext() {
		el := (*element)(self.next)
		self.next = el.next
		return el.value
	}
	return nil
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
	return &iterator{self.head}
}
func (self *List) Push(t thing) {
	new_head := &element{t, self.head}
	if atomic.CompareAndSwapPointer(&(self.head), new_head.next, unsafe.Pointer(new_head)) {
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
