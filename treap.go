
package gotomic

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Treap struct {
	root *nodeHandle
	size int64
}

type node struct {
	parent *nodeHandle
	left *nodeHandle
	right *nodeHandle
	weight int32
	key Comparable
	value Thing
}

type nodeHandle Handle
func (handle *nodeHandle) ropen(t Transaction) (rval *node, err error) {
	node, err := t.Read((*Handle)(handle))
	if err != nil {
		return nil, err
	}
	return node.(*node), nil
}
func (handle *nodeHandle) wopen(t Transaction) (rval *node, err error) {
	node, err := t.Write((*Handle)(handle))
	if err != nil {
		return nil, err
	}
	return node.(*node), nil
}
func newNodeHandle(parent *nodeHandle, k Comparable, v Thing) *nodeHandle {
	return (*nodeHandle)(NewHandle(&node{parent, nil, nil, rand.Int31(), k, v}))
}
func (handle *nodeHandle) rotate() {
}
func (handle *nodeHandle) insert(t Transaction, k Comparable, v Thing) (err error) {
	self, err := handle.ropen(t)
	if err != nil {
		return err
	}
	switch cmp := k.Compare(self.key); {
	case cmp < 0:
		if self.left == nil {
			if self, err = handle.wopen(t); err != nil {
				return err
			} else {
				self.left = newNodeHandle(handle, k, v)
				self.left.rotate()
			}
		} else {
			if err = self.left.insert(t, k, v); err != nil {
				return err
			}
		}
	case cmp > 0:
		if self.right == nil {
			if self, err = handle.wopen(t); err != nil {
				return err
			} else {
				self.right = newNodeHandle(handle, k, v)
				self.right.rotate()
			}
		} else {
			if err = self.right.insert(t, k, v); err != nil {
				return err
			}
		}
		
	default:
		if self, err = handle.wopen(t); err != nil {
			return err
		} else {
			self.value = v
		}
	}	
	return
}

	
