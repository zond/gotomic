package gotomic

import (
	"testing"
	"reflect"
	"fmt"
)

type key string
func (self key) HashCode() uint32 {
	var sum uint32
	for _, c := range string(self) {
		sum += uint32(c)
	}
	return sum
}
func (self key) Equals(t thing) bool {
	if s, ok := t.(key); ok {
		return s == self
	}
	return false
}

func assertMappy(t *testing.T, h *Hash, cmp map[Hashable]thing) {
	if e := h.verify(); e != nil {
		t.Errorf("%v should be valid, got %v", e)
	}
	if h.Size() != len(cmp) {
		t.Errorf("%v should have size %v, but had size %v", h, len(cmp), h.Size())
	}
	if tm := h.toMap(); !reflect.DeepEqual(tm, cmp) {
		t.Errorf("%v should be %#v but is %#v", h, cmp, tm)
	}
	for k, v := range cmp {
		if mv := h.get(k); !reflect.DeepEqual(mv, v) {
			t.Errorf("%v.get(%v) should produce %v but produced %v", h, k, v, mv)
		}
	}
}

func TestPutGet(t *testing.T) {
	h := NewHash()
	assertMappy(t, h, map[Hashable]thing{})
	h.put(key("a"), "b")
	assertMappy(t, h, map[Hashable]thing{key("a"): "b"})
	h.put(key("a"), "b")
	assertMappy(t, h, map[Hashable]thing{key("a"): "b"})
	h.put(key("c"), "d")
	assertMappy(t, h, map[Hashable]thing{key("a"): "b", key("c"): "d"})
	fmt.Println(h.describe())
}

