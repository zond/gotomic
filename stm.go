package gotomic

import (
	"fmt"
	"sort"
	"sync/atomic"
	"unsafe"
)

const (
	undecided = iota
	read_check
	successful
	failed
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
	newVersion.lockedBy = nil
	return &newVersion
}

type snapshot struct {
	old *version
	neu *version
}

type write struct {
	handle *Handle
	snapshot *snapshot
}
type writes []write
func (self writes) Len() int {
	return len(self)
}
func (self writes) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
func (self writes) Less(i, j int) bool {
	return uintptr(unsafe.Pointer(self[i].handle)) < uintptr(unsafe.Pointer(self[j].handle))
}



/*
 Transaction is based on "Concurrent Programming Without Locks" by Keir Fraser and Tim Harris <http://www.cl.cam.ac.uk/research/srg/netos/papers/2007-cpwl.pdf>

 It has a few tweaks that I don't believe break it (but I haven't even tried proving it):

 1) It has an ever increasing counter for the last transaction to commit. 

 It uses this counter to fail transactions fast when they try to read a value that another 
 transaction has changed since the first transaction began. 

 2) It copies the data not only on write opening, but also on read opening.

 These changes will make the transactions act more along the lines of "Sandboxing Transactional Memory" by Luke Dalessandro and Michael L. Scott <http://www.cs.rochester.edu/u/scott/papers/2012_TRANSACT_sandboxing.pdf> and will hopefully avoid the need to kill transactions exhibiting invalid behaviour due to inconsistent states.
*/
type Transaction struct {
	/*
	 Steadily incrementing number for each committed transaction.
	*/
	commitNumber uint64
	status       int32
	readHandles  map[*Handle]*snapshot
	writeHandles map[*Handle]*snapshot
	sortedWrites writes
}

func NewTransaction() *Transaction {
	return &Transaction{
		atomic.LoadUint64(&nextCommit),
		undecided,
		make(map[*Handle]*snapshot),
		make(map[*Handle]*snapshot),
		nil,
	}
}
func (self *Transaction) getStatus() int32 {
	return atomic.LoadInt32(&self.status)
}
func (self *Transaction) objRead(h *Handle) (rval *version, err error) {
	version := h.getVersion()
	for {
		if version.commitNumber > self.commitNumber {
			err = fmt.Errorf("%v has changed", h.getVersion().content)
			break
		}
		if version.lockedBy == nil {
			rval = version
			break
		}
		other := version.lockedBy
		if other.getStatus() == read_check {
			if self.getStatus() != read_check || self.commitNumber > other.commitNumber {
				other.commit()
			} else {
				other.Abort()
			}
		}
		if other.getStatus() != successful {
			rval = version
			break
		}
		version = h.getVersion()
	}
	return
}
func (self *Transaction) sortWrites() {
	if self.sortedWrites == nil && self.writeHandles != nil {
		for handle, _ := range self.writeHandles {
			self.sortedWrites = append(self.sortedWrites, write{handle, self.writeHandles[handle]})
		}
		sort.Sort(self.sortedWrites)
		self.writeHandles = nil
	} else {
		panic(fmt.Errorf("%#v.sortWrites() called in illegal state!", self))
	}
}
func (self *Transaction) release() {
	if self.sortedWrites != nil && self.writeHandles == nil {
		stat := self.getStatus()
		if stat == successful {
			self.commitNumber = atomic.AddUint64(&nextCommit, 1)
		}
		sw := self.sortedWrites
		for _, w := range sw {
			current := w.handle.getVersion()
			for current.lockedBy == self {
				wanted := w.snapshot.old
				if stat == successful {
					wanted = w.snapshot.neu
					wanted.commitNumber = self.commitNumber
				}
				if w.handle.replace(current, wanted) {
					break
				}
				current = w.handle.getVersion()
			}
		}
	} else if self.sortedWrites == nil && self.writeHandles == nil {
		// all is well, ignore me
	} else {
		panic(fmt.Errorf("%#v.release() called in illegal state!", self))
	}
}
func (self *Transaction) acquire() bool {
	for _, w := range self.sortedWrites {
		for {
			lockedVersion := w.snapshot.old.clone()
			lockedVersion.lockedBy = self
			if w.handle.replace(w.snapshot.old, lockedVersion) {
				break
			}
			current := w.handle.getVersion()
			if current.lockedBy == nil {
				return false
			}
			if current.lockedBy == self {
				break
			}
			current.lockedBy.commit()
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
 commit the transaction without mutating anything inside it except with atomic
 methods. Useful for other helpful Transactions.
*/
func (self *Transaction) commit() bool {
	if !self.acquire() {
		self.Abort()
		return false
	}
	defer self.release()
	atomic.CompareAndSwapInt32(&self.status, undecided, read_check)
	if !self.readCheck() {
		self.Abort()
		return false
	}
	atomic.StoreInt32(&self.status, successful)
	return self.getStatus() == successful
}
/*
 Commit the transaction. Will return whether the commit was successful or not.
*/
func (self *Transaction) Commit() bool {
	self.sortWrites()
	return self.commit()
}

/*
 Abort the transaction.

 Unless the transaction is half-committed Abort isn't really necessary.
*/
func (self *Transaction) Abort() {
	atomic.StoreInt32(&self.status, failed)
	self.release()
}

/*
 Read will return a version of the data in h that is guaranteed to not have been changed since this Transaction started.

 Any changes made to the return value will *not* be saved when the Transaction commits.

 If another Transaction changes the data in h before this Transaction commits the commit will fail.
*/
func (self *Transaction) Read(h *Handle) (rval Clonable, err error) {
	if self.getStatus() != undecided {
		return nil, fmt.Errorf("%v is not undecided", self)
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
	if self.getStatus() != undecided {
		return nil, fmt.Errorf("%v is not undecided", self)
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
