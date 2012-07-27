
package gotomic

import (
	"math/rand"
	"time"
	"bytes"
	"sync/atomic"
	"fmt"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

/*
 Transaction controlled treap
 */
type treap struct {
	root *nodeHandle
}
func (self *treap) Clone() Clonable {
	return &treap{self.root}
}

/*
 Non-transaction controlled "user space" type
 */
type Treap struct {
	handle *Handle
	size int64
}
func NewTreap() *Treap {
	return &Treap{NewHandle(&treap{}), 0}
}
/*
 Get a readable *treap from the Treap
 */
func (self *Treap) ropen(t *Transaction) (*treap, error) {
	r, err := t.Read(self.handle)
	if err != nil {
		return nil, err
	}
	return r.(*treap), nil
}
/*
 Get a writable *treap from the Treap
 */
func (self *Treap) wopen(t *Transaction) (*treap, error) {
	r, err := t.Write(self.handle)
	if err != nil {
		return nil, err
	}
	return r.(*treap), nil
}
func (treap *Treap) Describe() string {
	rval, err := treap.describe()
	for err != nil {
		rval, err = treap.describe()
	}
	return rval
}
func (treap *Treap) describe() (rval string, err error) {
	t := NewTransaction()
	self, err := treap.ropen(t)
	if err != nil {
		return
	}
	buf := bytes.NewBufferString(fmt.Sprintf("&Treap{%p size:%v}\n", treap, treap.size))
	if self.root != nil {
		err = self.root.describe(t, buf, 0)
		if err != nil {
			return
		}
	}
	return string(buf.Bytes()), nil
}
func (treap *Treap) Put(k Comparable, v Thing) Thing {
	rval, err := treap.put(k, v)
	for err != nil {
		rval, err = treap.put(k, v)
	}
	return rval
}
func (treap *Treap) put(k Comparable, v Thing) (Thing, error) {
	t := NewTransaction()
	self, err := treap.ropen(t)
	if err != nil {
		return nil, err
	}
	newNode := newNodeHandle(k, v)
	newRoot, err := self.root.insert(t, newNode)
	if err != nil {
		return nil, err
	}
	if newRoot != self.root {
		self, err = treap.wopen(t)
		if err != nil {
			return nil, err
		}
		self.root = newRoot
	}
	if !t.Commit() {
		return nil, fmt.Errorf("%v changed during put", treap)
	}
	atomic.AddInt64(&treap.size, 1)
	return nil, nil
}


type node struct {
	left *nodeHandle
	right *nodeHandle
	weight int32
	key Comparable
	value Thing
}
func (self *node) Clone() Clonable {
	rval := *self
	return &rval
}

type nodeHandle Handle
func (handle *nodeHandle) ropen(t *Transaction) (*node, error) {
	n, err := t.Read((*Handle)(handle))
	if err != nil {
		return nil, err
	}
	return n.(*node), nil
}
func (handle *nodeHandle) describe(t *Transaction, buf *bytes.Buffer, indent int) error {
	self, err := handle.ropen(t)
	if err != nil {
		return err
	}
	for i := 0; i < indent; i++ {
		fmt.Fprintf(buf, " ")
	}
	fmt.Fprintf(buf, "%v => %v (%v)\n", self.key, self.value, self.weight)
	if self.left != nil {
		fmt.Fprintf(buf, "l:")
		err = self.left.describe(t, buf, indent + 1)
		if err != nil {
			return err
		}
	}
	if self.right != nil {
		fmt.Fprintf(buf, "r:")
		err = self.right.describe(t, buf, indent + 1)
		if err != nil {
			return err
		}
	}
	return nil
}
func (handle *nodeHandle) wopen(t *Transaction) (*node, error) {
	r, err := t.Write((*Handle)(handle))
	if err != nil {
		return nil, err
	}
	return r.(*node), nil
}
func newNodeHandle(k Comparable, v Thing) *nodeHandle {
	return (*nodeHandle)(NewHandle(&node{nil, nil, rand.Int31(), k, v}))
}
func (handle *nodeHandle) rotateLeft(t *Transaction) (result *nodeHandle, err error) {
	self, err := handle.wopen(t)
	if err != nil {
		return
	}
	result = self.left
	resultNode, err := result.wopen(t)
	if err != nil {
		return
	}
	tmp := resultNode.right
	resultNode.right = handle
	self.left = tmp
	return
}
func (handle *nodeHandle) rotateRight(t *Transaction) (result *nodeHandle, err error) {
	self, err := handle.wopen(t)
	if err != nil {
		return
	}
	result = self.right
	resultNode, err := result.wopen(t)
	if err != nil {
		return
	}
	tmp := resultNode.left
	resultNode.left = handle
	self.right = tmp
	return
}
func (handle *nodeHandle) insert(t *Transaction, newHandle *nodeHandle) (result *nodeHandle, err error) {
	if handle == nil {
		return newHandle, nil
	}
	result = handle
	self, err := handle.ropen(t)
	if err != nil {
		return
	}
	newNode, err := newHandle.ropen(t)
	if err != nil {
		return 
	}
	switch cmp := newNode.key.Compare(self.key); {
	case cmp < 0:
		var newLeft *nodeHandle
		newLeft, err = self.left.insert(t, newHandle)
		if err != nil {
			return
		}
		if newLeft != self.left {
			self, err = handle.wopen(t)
			if err != nil {
				return
			}
			self.left = newLeft
			var leftNode *node
			leftNode, err = self.left.ropen(t)
			if err != nil {
				return
			}
			if leftNode.weight < self.weight {
				result, err = handle.rotateLeft(t)
				if err != nil {
					return
				}
			}
		}
	case cmp > 0:
		var newRight *nodeHandle
		newRight, err = self.right.insert(t, newHandle)
		if err != nil {
			return
		}
		if newRight != self.right {
			self, err = handle.wopen(t)
			if err != nil {
				return
			}
			self.right = newRight
			var rightNode *node
			rightNode, err = self.right.ropen(t)
			if rightNode.weight < self.weight {
				result, err = handle.rotateRight(t)
				if err != nil {
					return
				}
			}
		}
	default:
		if self, err = handle.wopen(t); err != nil {
			return nil, err
		} else {
			self.value = newNode.value
		}
	}	
	return
}

	
