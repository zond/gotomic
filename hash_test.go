package gotomic

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
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
	<-do
	cmp := make(map[Hashable]Thing)
	n := 100000
	for i := 0; i < n; i++ {
		k := StringKey(fmt.Sprint(s, i))
		v := fmt.Sprint(k, "value")
		if hv, _ := h.Put(k, v); hv != nil {
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
		if hv, _ := h.Put(k, v2); !reflect.DeepEqual(hv, v) {
			t.Errorf("2 Put(%v, %v) should produce %v but produced %v", k, v2, v, hv)
		}
	}
	for k, v := range cmp {
		if hv, _ := h.Get(k); !reflect.DeepEqual(hv, v) {
			t.Errorf("2 Get(%v) should produce %v but produced %v", k, v, hv)
		}
	}
	for k, v := range cmp {
		if hv, _ := h.Delete(k); !reflect.DeepEqual(hv, v) {
			t.Errorf("1 Delete(%v) should produce %v but produced %v", k, v, hv)
		}
	}
	for k, _ := range cmp {
		if hv, _ := h.Delete(k); hv != nil {
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
	<-do
	for j := 0; j < i; j++ {
		k := hashInt(j)
		m.Put(k, rand.Int())
		m.Get(k)
	}
	done <- true
}

func BenchmarkHashConc(b *testing.B) {
	b.StopTimer()
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
		<-done
	}
	runtime.GOMAXPROCS(1)
}

func TestPutIfPresent(t *testing.T) {
	h := NewHash()
	assertMappy(t, h, map[Hashable]Thing{})
	if h.PutIfPresent(StringKey("k"), StringKey("v"), StringKey("blabla")) {
		t.Error(h, "should not contain 'k': 'v'")
	}
	assertMappy(t, h, map[Hashable]Thing{})
	if old, _ := h.Put(StringKey("k"), StringKey("v")); old != nil {
		t.Error(h, "should not contain 'k': 'v'")
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("k"): StringKey("v")})
	if h.PutIfPresent(StringKey("k"), StringKey("v3"), StringKey("v2")) {
		t.Error(h, "should not contain 'k': 'v2'")
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("k"): StringKey("v")})
	if !h.PutIfPresent(StringKey("k"), StringKey("v2"), StringKey("v")) {
		t.Error(h, "should contain 'k': 'v'")
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("k"): StringKey("v2")})
	if h.PutIfPresent(StringKey("k"), StringKey("v2"), StringKey("v")) {
		t.Error(h, "should not contain 'k': 'v'")
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("k"): StringKey("v2")})
	if !h.PutIfPresent(StringKey("k"), StringKey("v3"), StringKey("v2")) {
		t.Error(h, "should contain 'k': 'v2'")
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("k"): StringKey("v3")})
}

func TestNilValues(t *testing.T) {
	h := NewHash()
	assertMappy(t, h, map[Hashable]Thing{})
	h.Put(StringKey("k"), nil)
	assertMappy(t, h, map[Hashable]Thing{StringKey("k"): nil})
	v, ok := h.Get(StringKey("k"))
	if !ok {
		t.Error(h, "should contain 'k'")
	}
	if v != nil {
		t.Error(h, "should contain 'k' => nil")
	}
}

func TestPutIfMissing(t *testing.T) {
	h := NewHash()
	assertMappy(t, h, map[Hashable]Thing{})
	if !h.PutIfMissing(StringKey("k"), "v") {
		t.Error(h, "should not contain 'k'")
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("k"): "v"})
	if h.PutIfMissing(StringKey("k"), "v") {
		t.Error(h, "should contain 'k'")
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("k"): "v"})
}

func TestConcurrency(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	h := NewHash()
	cmp := make(map[Hashable]Thing)
	for i := 0; i < 1000; i++ {
		k := StringKey(fmt.Sprint("StringKey", i))
		v := fmt.Sprint("value", i)
		h.Put(k, v)
		cmp[k] = v
	}
	assertMappy(t, h, cmp)
	do := make(chan bool)
	done := make(chan bool)
	for i := 0; i < runtime.NumCPU(); i++ {
		go fiddleHash(t, h, fmt.Sprint("fiddler-", i, "-"), do, done)
	}
	close(do)
	for i := 0; i < runtime.NumCPU(); i++ {
		<-done
	}
	assertMappy(t, h, cmp)
}

func TestHashEach(t *testing.T) {
	h := NewHash()
	h.Put(StringKey("a"), "1")
	h.Put(StringKey("b"), "2")
	h.Put(StringKey("c"), "3")
	h.Put(StringKey("d"), "4")

	cmp := make(map[Hashable]Thing)
	cmp[StringKey("a")] = "1"
	cmp[StringKey("b")] = "2"
	cmp[StringKey("c")] = "3"
	cmp[StringKey("d")] = "4"

	m := make(map[Hashable]Thing)

	h.Each(func(k Hashable, v Thing) bool {
		m[k] = v
		return false
	})

	if !reflect.DeepEqual(cmp, m) {
		t.Error(m, "should be", cmp)
	}
}

func TestHashEachInterrupt(t *testing.T) {
	h := NewHash()
	h.Put(StringKey("a"), "1")
	h.Put(StringKey("b"), "2")
	h.Put(StringKey("c"), "3")
	h.Put(StringKey("d"), "4")

	m := make(map[Hashable]Thing)

	interrupted := h.Each(func(k Hashable, v Thing) bool {
		m[k] = v
		
		// Break the iteration when we reach 2 elements
		return len(m) == 2
	})

	if !interrupted {
		t.Error("Iteration should have been interrupted.")
	}

	if len(m) != 2 {
		t.Error(m, "should have 2 elements. Have", len(m))
	}
}

func TestPutDelete(t *testing.T) {
	h := NewHash()
	if v, _ := h.Delete(StringKey("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{})
	h.Put(StringKey("a"), "b")
	if v, _ := h.Delete(StringKey("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("a"): "b"})
	h.Put(StringKey("a"), "b")
	if v, _ := h.Delete(StringKey("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("a"): "b"})
	h.Put(StringKey("c"), "d")
	if v, _ := h.Delete(StringKey("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("a"): "b", StringKey("c"): "d"})
	if v, _ := h.Delete(StringKey("a")); v != "b" {
		t.Error(h, "should be able to delete 'a' but got ", v)
	}
	if v, _ := h.Delete(StringKey("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("c"): "d"})
	if v, _ := h.Delete(StringKey("a")); v != nil {
		t.Error(h, "should not be able to delete 'a' but got ", v)
	}
	if v, _ := h.Delete(StringKey("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{StringKey("c"): "d"})
	if v, _ := h.Delete(StringKey("c")); v != "d" {
		t.Error(h, "should be able to delete 'c' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{})
	if v, _ := h.Delete(StringKey("c")); v != nil {
		t.Error(h, "should not be able to delete 'c' but got ", v)
	}
	if v, _ := h.Delete(StringKey("e")); v != nil {
		t.Error(h, "should not be able to delete 'e' but got ", v)
	}
	assertMappy(t, h, map[Hashable]Thing{})
}
