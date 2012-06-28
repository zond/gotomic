package gotomic

import (
	"testing"
	"reflect"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type key string
func (self key) HashCode() uint32 {
	var sum uint32
	for _, c := range string(self) {
		sum += uint32(c)
	}
	return sum
}
func (self key) Equals(t Thing) bool {
	if s, ok := t.(key); ok {
		return s == self
	}
	return false
}

func assertMappy(t *testing.T, h *Hash, cmp map[Hashable]Thing) {
	if e := h.Verify(); e != nil {
		t.Errorf("%v should be valid, got %v", e)
	}
	if h.Size() != len(cmp) {
		t.Errorf("%v should have size %v, but had size %v", h, len(cmp), h.Size())
	}
	if tm := h.ToMap(); !reflect.DeepEqual(tm, cmp) {
		t.Errorf("%v should be %#v but is %#v", h, cmp, tm)
	}
	for k, v := range cmp {
		if mv := h.Get(k); !reflect.DeepEqual(mv, v) {
			t.Errorf("%v.get(%v) should produce %v but produced %v", h, k, v, mv)
		}
	}
}

func fiddleHash(t *testing.T, h *Hash, s string) {
	cmp := make(map[Hashable]Thing)
	for i := 0; i < 1000; i++ {
		k := key(fmt.Sprint(s, rand.Int()))
		v := fmt.Sprint(k, "value")
		h.Put(k, v)
		cmp[k] = v
	}
	for k, v := range cmp {
		if hv := h.Get(k); !reflect.DeepEqual(hv, v) {
			t.Errorf("[%v] should produce %v but produced %v", k, v, hv)
		}
	}
}

func TestPutDelete(t *testing.T) {
	h := NewHash()
	if v := h.Delete(key("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{})
	h.Put(key("a"), "b")
	if v := h.Delete(key("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{key("a"): "b"})
	h.Put(key("a"), "b")
	if v := h.Delete(key("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{key("a"): "b"})
	h.Put(key("c"), "d")
	if v := h.Delete(key("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{key("a"): "b", key("c"): "d"})
	if v := h.Delete(key("a")); v != "b" {
		t.Error(h, "should be able to delete 'a' but got ", v)
	}
	if v := h.Delete(key("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{key("c"): "d"})
	if v := h.Delete(key("a")); v != nil {
		t.Error(h, "should not be able to delete 'a' but got ", v)
	}
	if v := h.Delete(key("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{key("c"): "d"})
	if v := h.Delete(key("c")); v != "d" {
		t.Error(h, "should be able to delete 'c' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{})
	if v := h.Delete(key("c")); v != nil {
		t.Error(h, "should not be able to delete 'c' but got ", v)
	}
	if v := h.Delete(key("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{})
}

