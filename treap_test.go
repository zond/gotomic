
package gotomic

import (
	"testing"
	"reflect"
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

func assertTreapSlice(t *testing.T, treap *Treap, keys []Comparable, values []Thing) {
	found_keys, found_values := treap.ToSlice()
	if !reflect.DeepEqual(keys, found_keys) {
		t.Errorf("%v.ToSlice keys should be %#v but was %#v", treap, keys, found_keys)
	}
	if !reflect.DeepEqual(values, found_values) {
		t.Errorf("%v.ToSlice values should be %#v but was %#v", treap, values, found_values)
	}
}

func TestTreapToSlice(t *testing.T) {
	treap := NewTreap()
	treap.Put(c(4), "4")
	treap.Put(c(6), "6")
	treap.Put(c(1), "1")
	treap.Put(c(8), "8")
	treap.Put(c(5), "5")
	assertTreapSlice(t, treap, []Comparable{c(1), c(4), c(5), c(6), c(8)}, []Thing{"1", "4", "5", "6", "8"})
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

func TestTreapMax(t *testing.T) {
	treap := NewTreap()
	k, v, ok := treap.Max()
	if ok {
		t.Error("should not have max value")
	}
	treap.Put(c(3), "3")
	k, v, ok = treap.Max()
	if !ok {
		t.Error("should have max value")
	}
	if k != c(3) {
		t.Error("max should be 3")
	}
	if v != "3" {
		t.Error("max should be 3")
	}
	treap.Put(c(2), "2")
	k, v, ok = treap.Max()
	if !ok {
		t.Error("should have max value")
	}
	if k != c(3) {
		t.Error("max should be 3")
	}
	if v != "3" {
		t.Errorf("max should be 3, not %#v", v)
	}
	treap.Put(c(4), "4")
	k, v, ok = treap.Max()
	if !ok {
		t.Error("should have max value")
	}
	if k != c(4) {
		t.Error("max should be 4")
	}
	if v != "4" {
		t.Error("max should be 4")
	}
}