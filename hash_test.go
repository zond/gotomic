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

func assertMappy(t *testing.T, h *hash, cmp map[Hashable]thing) {
	if tm := h.toMap(); !reflect.DeepEqual(tm, cmp) {
		t.Errorf("%v should be %#v but is %#v", h, cmp, tm)
	}
}

func TestPutGet(t *testing.T) {
	h := newHash()
	assertMappy(t, h, map[Hashable]thing{})
	h.put(key("a"), "b")
	assertMappy(t, h, map[Hashable]thing{key("a"): "b"})
	h.put(key("a"), "b")
	assertMappy(t, h, map[Hashable]thing{key("a"): "b"})
	h.put(key("c"), "d")
	assertMappy(t, h, map[Hashable]thing{key("a"): "b", key("c"): "d"})
	fmt.Println(h.describe())
}

