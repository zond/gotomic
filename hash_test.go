package gotomic

import (
	"testing"
	"reflect"
	"fmt"
	"math/rand"
	"runtime"
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
		fmt.Println(h.Describe())
		t.Errorf("%v should be valid, got %v", h, e)
	}
	if h.Size() != len(cmp) {
		t.Errorf("%v should have size %v, but had size %v", h, len(cmp), h.Size())
	}
	if tm := h.ToMap(); !reflect.DeepEqual(tm, cmp) {
		t.Errorf("%v should be %#v but is %#v", h, cmp, tm)
	}
	for k, v := range cmp {
		if mv, _ := h.Get(k); !reflect.DeepEqual(mv, v) {
			t.Errorf("%v.get(%v) should produce %v but produced %v", h, k, v, mv)
		}
	}
}

func fiddleHash(t *testing.T, h *Hash, s string, do, done chan bool) {
	<- do
	cmp := make(map[Hashable]Thing)
	n := 10000
	for i := 0; i < n; i++ {
		k := key(fmt.Sprint(s, rand.Int()))
		v := fmt.Sprint(k, "value")
		if hv := h.Put(k, v); hv != nil {
			t.Errorf("1 Put(%v, %v) should produce nil but produced %v", k, v, hv)
		}
		cmp[k] = v
	}
	for k, v := range cmp {
		if hv, _ := h.Get(k); !reflect.DeepEqual(hv, v) {
			t.Errorf("1 Get(%v) should produce %v but produced %v", k, v, hv)
		}
	}
	for k, v := range cmp {
		v2 := fmt.Sprint(v, ".2")
		cmp[k] = v2
		if hv := h.Put(k, v2); !reflect.DeepEqual(hv, v) {
			t.Errorf("2 Put(%v, %v) should produce %v but produced %v", k, v2, v, hv)
		}
	}
	for k, v := range cmp {
		if hv, _ := h.Get(k); !reflect.DeepEqual(hv, v) {
			t.Errorf("2 Get(%v) should produce %v but produced %v", k, v, hv)
		}
	}
	for k, v := range cmp {
		if hv := h.Delete(k); !reflect.DeepEqual(hv, v) {
			t.Errorf("1 Delete(%v) should produce %v but produced %v", k, v, hv)
		}
	}
	for k, _ := range cmp {
		if hv := h.Delete(k); hv != nil {
			t.Errorf("2 Delete(%v) should produce nil but produced %v", k, hv)
		}
	}
	for k, _ := range cmp {
		if hv, _ := h.Get(k); hv != nil {
			t.Errorf("3 Get(%v) should produce nil but produced %v", k, hv)
		}
	}
	done <- true
}

type hashInt int
func (self hashInt) HashCode() uint32 {
	return uint32(self)
}
func (self hashInt) Equals(t Thing) bool {
	if i, ok := t.(hashInt); ok {
		return i == self
	} 
	return false
}

func BenchmarkHash(b *testing.B) {
	m := NewHash()
	for i := 0; i < b.N; i++ {
		k := hashInt(i)
		m.Put(k, i)
		j, _ := m.Get(k)
		if j != i {
			b.Error("should be same value")
		}
	}
}

func action(b *testing.B, m *Hash, i int, do, done chan bool) {
	<- do
	for j := 0; j < i; j++ {
		k := hashInt(j)
		m.Put(k, j)
		l, _ := m.Get(k)
		if l != j {
			b.Error("should be same value")
		}
	}
	done <- true
}

func BenchmarkHashConc(b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	do := make(chan bool)
	done := make(chan bool)
	m := NewHash()
	for i := 0; i < runtime.NumCPU(); i++ {
		go action(b, m, b.N, do, done)
	}
	close(do)
	b.StartTimer()
	for i := 0; i < runtime.NumCPU(); i++ {
		<- done
	}
	b.StopTimer()
	runtime.GOMAXPROCS(1)
}

func TestConcurrency(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	h := NewHash()
	cmp := make(map[Hashable]Thing)
	for i := 0; i < 1000; i++ {
		k := key(fmt.Sprint("key", i))
		v := fmt.Sprint("value", i)
		h.Put(k, v)
		cmp[k] = v
	}
	assertMappy(t, h, cmp)
	do := make(chan bool)
	done := make(chan bool)
	for i := 0; i < runtime.NumCPU(); i++ {
		go fiddleHash(t, h, "fiddlerA", do, done)
	}
	close(do)
	for i := 0; i < runtime.NumCPU(); i++ {
		<- done
	}
	assertMappy(t, h, cmp)
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

