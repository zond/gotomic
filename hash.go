
package gotomic

import (
	"sync/atomic"
	"bytes"
	"unsafe"
	"fmt"
)

const MAX_EXPONENT = 32
const DEFAULT_LOAD_FACTOR = 0.5

type hashHit hit
func (self *hashHit) search(cmp *entry) (node *node) {
	node = self.node
	for {
		if node == nil {
			break
		}
		e := node.value.(*entry)
		if !e.real() {
			node = nil
			break
		}
		if e.hashCode != cmp.hashCode {
			node = nil
			break
		}
		if cmp.key.Equals(e.key) {
			break
		}
		node = node.next.node()
	}
	return
}

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
	return fmt.Sprintf("&entry{%0.32b/%0.32b, %v=>%v}", self.hashCode, self.hashKey, self.key, self.val())
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
func (self *hash) Size() int {
	return int(atomic.LoadUint64(&self.size))
}
func (self *hash) verify() error {
	bucket := self.getBucketByHashCode(0)
	if e := bucket.verify(); e != nil {
		return e
	}
	node := bucket.node()
	for node != nil {
		if en := node.value.(*entry); !en.real() {
			superIndex, subIndex := self.getBucketIndices(en.hashCode)
			subBuckets := *(*[]unsafe.Pointer)(self.buckets[superIndex])
			bucket := (*nodeRef)(subBuckets[subIndex])
			bucketEntry := bucket.node().value.(*entry)
			if bucketEntry != en {
				return fmt.Errorf("%v has a mock entry %v (%#v) that doesn't match the entry in bucket %v,%v: %v (%#v)", self, en, en, superIndex, subIndex, bucketEntry, bucketEntry)
			}
		}
		node = node.next.node()
	}
	return nil
}
func (self *hash) toMap() map[Hashable]thing {
	rval := make(map[Hashable]thing)
	bucket := self.getBucketByHashCode(0)
	node := bucket.node()
	for node != nil {
		if e := node.value.(*entry); e.real() {
			rval[e.key] = e.val()
		}
		node = node.next.node()
	}
	return rval
}
func (self *hash) describe() string {
	buffer := bytes.NewBufferString(fmt.Sprintf("&hash{%p size:%v exp:%v load:%v}\n", self, self.size, self.exponent, self.loadFactor))
	bucket := self.getBucketByHashCode(0)
	node := bucket.node()
	for node != nil {
		e := node.value.(*entry)
		if e.real() {
			fmt.Fprintln(buffer, "\t", e)
		} else {
			fmt.Fprintln(buffer, e)
		}
		bucket = node.next
		node = bucket.node()
	}
	return string(buffer.Bytes())
}
func (self *hash) String() string {
	return fmt.Sprint(self.toMap())
}
func (self *hash) get(k Hashable) (rval thing) {
	testEntry := newRealEntry(k, nil)
	bucket := self.getBucketByHashCode(testEntry.hashCode)
	hit := (*hashHit)(bucket.search(testEntry))
	if node := hit.search(testEntry); node != nil {
		return hit.node.value.(*entry).val()
	}
	return nil
}
func (self *hash) put(k Hashable, v thing) (rval thing) {
	newEntry := newRealEntry(k, v)
	for {
		bucket := self.getBucketByHashCode(newEntry.hashCode)
		hit := (*hashHit)(bucket.search(newEntry))
		if node := hit.search(newEntry); node == nil {
			if hit.leftNode.next.pushBefore(newEntry, hit.rightNode) {
				self.addSize(1)
				rval = nil
				break
			}
		} else {
			oldEntry := hit.node.value.(*entry)
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
			prev := self.getPreviousBucketIndex(mockEntry.hashKey)
			previousBucket := self.getBucketByIndex(prev)
			if hit := previousBucket.search(mockEntry); hit.node == nil {
				hit.leftNode.next.pushBefore(mockEntry, hit.rightNode)
			} else {
				atomic.CompareAndSwapPointer(&subBuckets[subIndex], nil, unsafe.Pointer(hit.ref))
			}
		}
	}
	return bucket
}