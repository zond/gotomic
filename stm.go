
package gotomic

import (
	"sync/atomic"
	"fmt"
	"unsafe"
	"sort"
)

const (
	UNDECIDED = iota
	READ_CHECK
	SUCCESSFUL
	FAILED
)

var nextCommit uint64 = 0 

/*
 Clonable types can be handled by the transaction layer.
 */
type Clonable interface {
	Clone() Clonable
}

/*
 Handle wraps any type of data that is supposed to be handled by the transaction layer.
 */
type Handle struct {
	/*
	 Will point to a version.
	 */
	unsafe.Pointer
}
/*
 NewHandle will wrap a Clonable value to enable its use in the transaction layer.
 */
func NewHandle(c Clonable) *Handle {
	return &Handle{unsafe.Pointer(&version{0, nil, c})}
} 
func (self *Handle) getVersion() *version {
	return (*version)(atomic.LoadPointer(&self.Pointer))
}
func (self *Handle) replace(old, neu *version) bool {
	return atomic.CompareAndSwapPointer(&self.Pointer, unsafe.Pointer(old), unsafe.Pointer(neu))
}

type handles []*Handle
func (self handles) Len() int {
	return len(self)
}
func (self handles) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
func (self handles) Less(i, j int) bool {
	return uintptr(unsafe.Pointer(self[i])) < uintptr(unsafe.Pointer(self[j]))
}

type version struct {
	/*
	 The number of the transaction that created this version.
	 */
	commitNumber uint64
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

/*
 Transaction is based on "Concurrent Programming Without Locks" by Keir Fraser and Tim Harris <http://www.cl.cam.ac.uk/research/srg/netos/papers/2007-cpwl.pdf>
 
 It has a few tweaks that I think, but can not prove, doesn't break it:

 1) It has an ever increasing counter for the last transaction to commit. 

 It uses this counter to fail fast when trying to read a value that another transaction has changed since it began. 
 
 2) It copies the data not only on write opening, but also on read opening.
 
 These changes will make the transactions act more along the lines of "Sandboxing Transactional Memory" by Luke Dalessandro and Michael L. Scott <http://www.cs.rochester.edu/u/scott/papers/2012_TRANSACT_sandboxing.pdf> and will hopefully avoid the need to kill transactions exhibiting invalid behaviour due to inconsistent states.
 */
type Transaction struct {
	/*
	 Steadily incrementing number for each committed transaction.
	 */
	commitNumber uint64
	status int32
	readHandles map[*Handle]*snapshot
	writeHandles map[*Handle]*snapshot
}
func NewTransaction() *Transaction {
	return &Transaction{
		atomic.LoadUint64(&nextCommit),
		UNDECIDED, 
		make(map[*Handle]*snapshot), 
		make(map[*Handle]*snapshot),
	}
}
func (self *Transaction) getStatus() int32 {
	return atomic.LoadInt32(&self.status)
}
func (self *Transaction) objRead(h *Handle) (rval *version, err error) {
	version := h.getVersion()
	if version.commitNumber > self.commitNumber {
		return nil, fmt.Errorf("%v has changed", h.getVersion().content)
	}
	if version.lockedBy == nil {
		return version, nil
	}
	other := version.lockedBy
	if other.getStatus() == READ_CHECK {
		if self.getStatus() != READ_CHECK || self.commitNumber > other.commitNumber {
			other.Commit()
		} else {
			other.Abort()
		}
	}
	if other.getStatus() == SUCCESSFUL {
		if other.commitNumber > self.commitNumber {
			return nil, fmt.Errorf("%v has changed", other.writeHandles[h].neu.content)
		}
		return other.writeHandles[h].neu, nil
	}
	return version, nil
}
func (self *Transaction) sortedWrites() []*Handle {
	var rval handles
	for handle, _ := range self.writeHandles {
		rval = append(rval, handle)
	}
	sort.Sort(rval)
	return rval
}
func (self *Transaction) release() {
	stat := self.getStatus()
	if stat == SUCCESSFUL {
		self.commitNumber = atomic.AddUint64(&nextCommit, 1)
	}
	for _, handle := range self.sortedWrites() {
		current := handle.getVersion()
		if current.lockedBy == self { 
			snapshot := self.writeHandles[handle]
			wanted := snapshot.old
			if stat == SUCCESSFUL {
				wanted = snapshot.neu
				wanted.commitNumber = self.commitNumber
			}
			handle.replace(current, wanted)
		}
	}
}
func (self *Transaction) acquire() bool {
 	for _, handle := range self.sortedWrites() {
		for {
			snapshot, _ := self.writeHandles[handle]
			lockedVersion := snapshot.old.clone()
			lockedVersion.lockedBy = self
			if handle.replace(snapshot.old, lockedVersion) {
				break
			}
			current := handle.getVersion()
			if current.lockedBy == nil {
				return false
			}
			if current.lockedBy == self {
				break
			}
			current.lockedBy.Commit()
		}
	}
	return true
}
func (self *Transaction) readCheck() bool {
	for handle, snapshot := range self.readHandles {
		if handle.getVersion() != snapshot.old {
			return false
		}
	}
	return true
}
/*
 Commit the transaction. Will return whether the commit was successful or not.
 */
func (self *Transaction) Commit() bool {
	if !self.acquire() {
		self.Abort()
		return false
	} 
	atomic.CompareAndSwapInt32(&self.status, UNDECIDED, READ_CHECK)
	if !self.readCheck() {
		self.Abort()
		return false
	}
	atomic.CompareAndSwapInt32(&self.status, READ_CHECK, SUCCESSFUL)
	self.release()
	return self.getStatus() == SUCCESSFUL;
}
/*
 Abort the transaction.
 
 Unless the transaction is half-committed Abort isn't really necessary.
 */
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
/*
 Read will return a version of the data in h that is guaranteed to not have been changed since this Transaction started.

 Any changes made to the return value will *not* be saved when the Transaction commits.
 
 If another Transaction changes the data in h before this Transaction commits the commit will fail.
 */
func (self *Transaction) Read(h *Handle) (rval Clonable, err error)  {
	if self.getStatus() != UNDECIDED {
		return nil, fmt.Errorf("%v is not UNDECIDED", self)
	}
	if snapshot, ok := self.readHandles[h]; ok {
		return snapshot.neu.content, nil
	}
	if snapshot, ok := self.writeHandles[h]; ok {
		return snapshot.neu.content, nil
	}
	oldVersion, err := self.objRead(h)
	if err != nil {
		return nil, err
	}
	newVersion := oldVersion.clone()
	self.readHandles[h] = &snapshot{oldVersion, newVersion}
	return newVersion.content, nil
}
/*
 Write will return a version of the data in h that is guaranteed to not have been changed since this Transaction started.

 All changes made to the return value *will* be saved when the Transaction commits.

 If another Transaction changes the data in h before this Transaction commits the commit will fail.
 */
func (self *Transaction) Write(h *Handle) (rval Clonable, err error) {
	if self.getStatus() != UNDECIDED {
		return nil, fmt.Errorf("%v is not UNDECIDED", self)
	}
	if snapshot, ok := self.writeHandles[h]; ok {
		return snapshot.neu.content, nil
	}
	if snapshot, ok := self.readHandles[h]; ok {
		delete(self.readHandles, h)
		self.writeHandles[h] = snapshot
		return snapshot.neu.content, nil
	}
	oldVersion, err := self.objRead(h)
	if err != nil {
		return nil, err
	}
	newVersion := oldVersion.clone()
	self.writeHandles[h] = &snapshot{oldVersion, newVersion}
	return newVersion.content, nil
}
