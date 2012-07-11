
package gotomic

import (
	"sync/atomic"
	"unsafe"
	"sort"
)

const (
	UNDECIDED = iota
	READ_CHECK
	SUCCESSFUL
	FAILED
)

var nextTransactionNumber uint64 = 0 

type Clonable interface {
	Clone() Clonable
}

type Handle struct {
	/*
	 Will point to a version.
	 */
	unsafe.Pointer
}
func NewHandle(c Clonable) *Handle {
	return &Handle{unsafe.Pointer(&version{0, nil, c})}
} 
func (self *Handle) getVersion() *version {
	return (*version)(atomic.LoadPointer(&self.Pointer))
}

type Handles []*Handle
func (self Handles) Len() int {
	return len(self)
}
func (self Handles) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
func (self Handles) Less(i, j int) bool {
	return uintptr(unsafe.Pointer(self[i])) < uintptr(unsafe.Pointer(self[j]))
}

type version struct {
	/*
	 The number of the transaction that created this version.
	 */
	transactionNumber uint64
	/*
	 The transaction (or nil) having locked this version.
	 */
	lockedBy *Transaction
	/*
	The content in this version.
	 */
	content Clonable
}
func (self *version) clone() *version {
	newVersion := *self
	newVersion.content = self.content.Clone()
	return &newVersion
}

type snapshot struct {
	old *version
	neu *version
}

type Transaction struct {
	/*
	 Steadily incrementing number for each created transaction.
	 */
	id uint64
	status int32
	readHandles map[*Handle]*snapshot
	writeHandles map[*Handle]*snapshot
}
func NewTransaction() *Transaction {
	return &Transaction{
		atomic.AddUint64(&nextTransactionNumber, 1), 
		UNDECIDED, 
		make(map[*Handle]*snapshot), 
		make(map[*Handle]*snapshot),
	}
}
func (self *Transaction) getStatus() int32 {
	return atomic.LoadInt32(&self.status)
}
func (self *Transaction) objRead(h *Handle) *version {
	version := h.getVersion()
	if version.lockedBy == nil {
		return version
	}
	other := version.lockedBy
	if other.getStatus() == READ_CHECK {
		if self.getStatus() != READ_CHECK || self.id > other.id {
			other.Commit()
		} else {
			other.Abort()
		}
	}
	if other.getStatus() == SUCCESSFUL {
		return other.writeHandles[h].neu
	}
	return other.writeHandles[h].old
}
func (self *Transaction) sortedWrites() []*Handle {
	var rval Handles
	for handle, _ := range self.writeHandles {
		rval = append(rval, handle)
	}
	sort.Sort(rval)
	return rval
}
func (self *Transaction) release() {
	stat := self.getStatus()
	for _, handle := range self.sortedWrites() {
		current := handle.getVersion()
		if current.lockedBy == self { 
			snapshot := self.writeHandles[handle]
			wanted := snapshot.old
			if stat == SUCCESSFUL {
				wanted = snapshot.neu
			}
			atomic.CompareAndSwapPointer(&handle.Pointer, unsafe.Pointer(current), unsafe.Pointer(wanted))
		}
	}
}
func (self *Transaction) Commit() bool {
	self.Abort()
	return false
}
func (self *Transaction) Abort() {
	for {
		current := self.getStatus()
		if current == FAILED {
			return
		}
		atomic.CompareAndSwapInt32(&self.status, current, FAILED)
	}
	self.release()
}
func (self *Transaction) Read(h *Handle) Clonable {
	if snapshot, ok := self.readHandles[h]; ok {
		return snapshot.neu.content
	}
	if snapshot, ok := self.writeHandles[h]; ok {
		return snapshot.neu.content
	}
	oldVersion := self.objRead(h)
	newVersion := oldVersion.clone()
	self.readHandles[h] = &snapshot{oldVersion, newVersion}
	return newVersion.content
}
func (self *Transaction) Write(h *Handle) Clonable {
	if snapshot, ok := self.writeHandles[h]; ok {
		return snapshot.neu.content
	}
	if snapshot, ok := self.readHandles[h]; ok {
		delete(self.readHandles, h)
		self.writeHandles[h] = snapshot
		return snapshot.neu.content
	}
	oldVersion := (*version)(atomic.LoadPointer(&h.Pointer))
	newVersion := oldVersion.clone()
	self.writeHandles[h] = &snapshot{oldVersion, newVersion}
	return newVersion.content
}
