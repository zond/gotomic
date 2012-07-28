
package gotomic

import (
	"testing"
)

func TestTreapPutGetDelete(t *testing.T) {
	treap := NewTreap()
	_, ok := treap.Get(c(3))
	if ok {
		t.Error("should not contain 3")
	}
	treap.Put(c(3), 44)
	v, ok := treap.Get(c(3))
	if !ok {
		t.Error("should contain 3")
	}
	if v != 44 {
		t.Error("should be 44")
	}
	v, ok = treap.Delete(c(3))
	if !ok {
		t.Error("should contain 3")
	}
	if v != 44 {
		t.Error("should be 44")
	}
	v, ok = treap.Get(c(3))
	if ok {
		t.Error("should not contain 3")
	}
	v, ok = treap.Delete(c(3))
	if v == 44 {
		t.Error("should not be 44")
	}
	if ok {
		t.Error("should not contain 3")
	}
}

func TestTreapMin(t *testing.T) {
	treap := NewTreap()
	k, v, ok := treap.Min()
	if ok {
		t.Error("should not have min value")
	}
	treap.Put(c(3), "3")
	k, v, ok = treap.Min()
	if !ok {
		t.Error("should have min value")
	}
	if k != c(3) {
		t.Error("min should be 3")
	}
	if v != "3" {
		t.Error("min should be 3")
	}
	treap.Put(c(2), "2")
	k, v, ok = treap.Min()
	if !ok {
		t.Error("should have min value")
	}
	if k != c(2) {
		t.Error("min should be 2")
	}
	if v != "2" {
		t.Errorf("min should be 2, not %#v", v)
	}
	treap.Put(c(4), "4")
	k, v, ok = treap.Min()
	if !ok {
		t.Error("should have min value")
	}
	if k != c(2) {
		t.Error("min should be 2")
	}
	if v != "2" {
		t.Error("min should be 2")
	}
}