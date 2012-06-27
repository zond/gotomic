
package gotomic

import (
	"sync/atomic"
	"unsafe"
	"fmt"
)

const MAX_EXPONENT = 32

type Hashable interface {
	Equals(thing) bool
	HashCode() uint32
}

type entry struct {
	hashCode uint32
	hashKey uint32
	key Hashable
	value thing
}
func newMockEntry(hashCode uint32) *entry {
	return &entry{hashCode, reverse(hashCode) &^ 1, nil, nil}
}
func (self *entry) Compare(t thing) int {
	if e, ok := t.(*entry); ok {
		if self.hashKey > e.hashKey {
			return 1
		} else if self.hashKey < e.hashKey {
			return -1
		} else {
			return 0
		}
	}
	panic(fmt.Sprint(self, " can only compare itself against other *entry, not against ", t))
}


type hash struct {
	exponent uint32
	buckets []unsafe.Pointer
}
func newHash() *hash {
	rval := &hash{0, make([]unsafe.Pointer, MAX_EXPONENT)}
	b := make([]unsafe.Pointer, 1)
	rval.buckets[0] = unsafe.Pointer(&b)
	return rval
}
func (self *hash) grow() {
	oldExponent := atomic.LoadUint32(&self.exponent)
	newExponent := oldExponent + 1
	newBuckets := make([]unsafe.Pointer, 1 << oldExponent)
	if atomic.CompareAndSwapPointer(&self.buckets[newExponent], nil, unsafe.Pointer(&newBuckets)) {
		atomic.CompareAndSwapUint32(&self.exponent, oldExponent, newExponent)
	}
}
func (self *hash) getPreviousBucketIndex(bucketKey uint32) uint32 {
	exp := atomic.LoadUint32(&self.exponent)
	return reverse( ((bucketKey >> (MAX_EXPONENT - exp)) - 1) << (MAX_EXPONENT - exp));
}
func (self *hash) getBucketByHashCode(hashCode uint32) *nodeRef {
	return self.getBucketByIndex(hashCode & ((1 << self.exponent) - 1))
}
func (self *hash) getBucketIndices(index uint32) (superIndex, subIndex uint32) {
	if index > 0 {
		superIndex = log2(index)
		subIndex = index - (1 << superIndex)
		superIndex++
	}
	return
}
func (self *hash) getBucketByIndex(index uint32) *nodeRef {
	superIndex, subIndex := self.getBucketIndices(index)
	subBuckets := *(*[]unsafe.Pointer)(self.buckets[superIndex])
	var bucket *nodeRef
	for {
		bucket := (*nodeRef)(subBuckets[subIndex])
		if bucket != nil {
			break
		}
		mockEntry := newMockEntry(index)
		if index == 0 {
			bucket := new(nodeRef)
			bucket.push(mockEntry)
			atomic.CompareAndSwapPointer(&subBuckets[subIndex], nil, unsafe.Pointer(bucket))
		} else {
			previousBucket := self.getBucketByIndex(self.getPreviousBucketIndex(mockEntry.hashKey))
			_, match, _ := previousBucket.search(mockEntry)
			if match == nil {
			} else {
				
			}
		}
	}
	return bucket
}