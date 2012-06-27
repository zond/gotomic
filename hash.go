
package gotomic

import (
	"sync/atomic"
	"bytes"
	"unsafe"
	"fmt"
)

const MAX_EXPONENT = 32
const DEFAULT_LOAD_FACTOR = 0.5

type Hashable interface {
	Equals(thing) bool
	HashCode() uint32
}

type entry struct {
	hashCode uint32
	hashKey uint32
	key Hashable
	value unsafe.Pointer
}
func newRealEntry(k Hashable, v thing) *entry {
	hc := k.HashCode()
	return &entry{hc, reverse(hc) | 1, k, unsafe.Pointer(&v)}
}
func newMockEntry(hashCode uint32) *entry {
	return &entry{hashCode, reverse(hashCode) &^ 1, nil, nil}
}
func (self *entry) real() bool {
	return self.hashKey & 1 == 1
}
func (self *entry) val() thing {
	if self.value == nil {
		return nil
	}
	return *(*thing)(self.value)
}
func (self *entry) String() string {
	return fmt.Sprintf("&entry{%b/%b, %v=>%v}", self.hashCode, self.hashKey, self.key, self.val())
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
	size uint64
	loadFactor float64
}
func newHash() *hash {
	rval := &hash{0, make([]unsafe.Pointer, MAX_EXPONENT), 0, DEFAULT_LOAD_FACTOR}
	b := make([]unsafe.Pointer, 1)
	rval.buckets[0] = unsafe.Pointer(&b)
	return rval
}
func (self *hash) String() string {
	buffer := bytes.NewBufferString("{")
	node := self.getBucketByHashCode(0).node()
	for node != nil {
		e := node.value.(*entry)
		if e.real() {
			fmt.Fprintf(buffer, "%#v => %#v", e.key, e.val())
		}
		node = node.next.node()
		if e.real() && node != nil {
			fmt.Fprint(buffer, ", ")
		}
	}
	fmt.Fprint(buffer, "}")
	return string(buffer.Bytes())
}
func (self *hash) put(k Hashable, v thing) (rval thing) {
	newEntry := newRealEntry(k, v)
	for {
		bucket := self.getBucketByHashCode(newEntry.hashCode)
		if before, match, after := bucket.search(newEntry); match == nil {
			if before.next.pushBefore(newEntry, after) {
				self.addSize(1)
				rval = nil
				break
			}
		} else {
			oldEntry := match.value.(*entry)
			rval = *(*thing)(atomic.LoadPointer(&oldEntry.value))
			atomic.StorePointer(&oldEntry.value, newEntry.value)
			break
		}
	}
	return
}
func (self *hash) addSize(i int) {
	atomic.AddUint64(&self.size, uint64(i))
	if atomic.LoadUint64(&self.size) > uint64(self.loadFactor * float64(uint32(1) << self.exponent)) {
		self.grow()
	}
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
func (self *hash) getBucketByIndex(index uint32) (bucket *nodeRef) {
	superIndex, subIndex := self.getBucketIndices(index)
	subBuckets := *(*[]unsafe.Pointer)(self.buckets[superIndex])
	for {
		bucket = (*nodeRef)(subBuckets[subIndex])
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
			if before, match, after := previousBucket.search(mockEntry); match == nil {
				before.next.pushBefore(mockEntry, after)
			} else {
				atomic.CompareAndSwapPointer(&subBuckets[subIndex], nil, unsafe.Pointer(match))
			}
		}
	}
	return bucket
}