
package gotomic

import (
	"testing"
)

func TestPutGetDelete(t *testing.T) {
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
