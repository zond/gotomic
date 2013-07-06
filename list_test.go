package gotomic

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"runtime"
	"testing"
	"time"
)

type c int

func (self c) Compare(t Thing) int {
	if s, ok := t.(c); ok {
		if self > s {
			return 1
		} else if self < s {
			return -1
		} else {
			return 0
		}
	}
	panic(fmt.Errorf("%#v can only compare to other c's, not %#v of type %T", self, t, t))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func fiddle(t *testing.T, nr *element, do, done chan bool) {
	<-do
	num := 10000
	for i := 0; i < num; i++ {
		x := rand.Int()
		nr.add(x)
	}
	for i := 0; i < num; i++ {
		if x, ok := nr.remove(); !ok {
			t.Errorf("%v should pop something, but got %v", nr, x)
		}
	}
	done <- true
}

func fiddleAndAssertSort(t *testing.T, nr *element, do chan bool, ichan, rchan chan []c) {
	<-do
	num := 1000
	var injected []c
	var removed []c
	for i := 0; i < num; i++ {
		v := c(-int(math.Abs(float64(rand.Int()))))
		nr.inject(v)
		injected = append(injected, v)
		if err := nr.verify(); err != nil {
			t.Error(nr, "should be correct, but got", err)
		}
	}
	for i := 0; i < num; i++ {
		if r, ok := nr.remove(); ok {
			removed = append(removed, r.(c))
		} else {
			t.Error(nr, "should remove something, but got", r)
		}
	}
	ichan <- injected
	rchan <- removed
}

func assertListy(t *testing.T, l *List, cmp []Thing) {
	if l.Size() != len(cmp) {
		t.Errorf("%v should have size %v but had %v", l, len(cmp), l.Size())
	}
	if sl := l.ToSlice(); !reflect.DeepEqual(sl, cmp) {
		t.Errorf("%v should be %#v but is %#v", l, cmp, sl)
	}
	tmp := make([]Thing, len(cmp))
	for ind, v := range cmp {
		popped, _ := l.Pop()
		tmp[len(cmp)-ind-1] = popped
		if !reflect.DeepEqual(v, popped) {
			t.Errorf("element %v of %v should be %v but was %v", ind, l, v, popped)
		}
	}
	for _, v := range tmp {
		l.Push(v)
	}
}

func TestList(t *testing.T) {
	l := NewList()
	assertListy(t, l, []Thing{})
	l.Push("plur")
	assertListy(t, l, []Thing{"plur"})
	l.Push("knap")
	assertListy(t, l, []Thing{"knap", "plur"})
	l.Push("hehu")
	assertListy(t, l, []Thing{"hehu", "knap", "plur"})
	l.Push("blar")
	assertListy(t, l, []Thing{"blar", "hehu", "knap", "plur"})
}

func assertSlicey(t *testing.T, nr *element, cmp []Thing) {
	if sl := nr.ToSlice(); !reflect.DeepEqual(sl, cmp) {
		t.Errorf("%v should be %#v but is %#v", nr.Describe(), cmp, sl)
	}
}

func assertPop(t *testing.T, nr *element, th Thing) {
	p, _ := nr.remove()
	if !reflect.DeepEqual(p, th) {
		t.Error(nr, " should pop ", th, " but popped ", p)
	}
}

func TestPushPop(t *testing.T) {
	nr := new(element)
	assertSlicey(t, nr, []Thing{nil})
	nr.add("hej")
	assertSlicey(t, nr, []Thing{nil, "hej"})
	nr.add("haj")
	assertSlicey(t, nr, []Thing{nil, "haj", "hej"})
	nr.add("hoj")
	assertSlicey(t, nr, []Thing{nil, "hoj", "haj", "hej"})
	assertPop(t, nr, "hoj")
	assertSlicey(t, nr, []Thing{nil, "haj", "hej"})
	assertPop(t, nr, "haj")
	assertSlicey(t, nr, []Thing{nil, "hej"})
	assertPop(t, nr, "hej")
	assertSlicey(t, nr, []Thing{nil})
	assertPop(t, nr, nil)
	assertSlicey(t, nr, []Thing{nil})
}

func TestConcPushPop(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(element)
	assertSlicey(t, nr, []Thing{nil})
	nr.add("1")
	nr.add("2")
	nr.add("3")
	nr.add("4")
	do := make(chan bool)
	done := make(chan bool)
	go fiddle(t, nr, do, done)
	go fiddle(t, nr, do, done)
	go fiddle(t, nr, do, done)
	go fiddle(t, nr, do, done)
	close(do)
	<-done
	<-done
	<-done
	<-done
	assertSlicey(t, nr, []Thing{nil, "4", "3", "2", "1"})
}

const ANY = "ANY VALUE"

func searchTest(t *testing.T, nr *element, s c, l, n, r Thing) {
	h := nr.search(s)
	if (l != ANY && !reflect.DeepEqual(h.left.val(), l)) ||
		(n != ANY && !reflect.DeepEqual(h.element.val(), n)) ||
		(r != ANY && !reflect.DeepEqual(h.right.val(), r)) {
		t.Error(nr, ".search(", s, ") should produce ", l, n, r, " but produced ", h.left.val(), h.element.val(), h.right.val())
	}
}

func TestListEach(t *testing.T) {
	nr := new(element)
	nr.add("h")
	nr.add("g")
	nr.add("f")
	nr.add("d")
	nr.add("c")
	nr.add("b")

	var a []Thing

	nr.each(func(t Thing) bool {
		a = append(a, t)
		return false
	})

	exp := []Thing{nil, "b", "c", "d", "f", "g", "h"}
	if !reflect.DeepEqual(a, exp) {
		t.Error(a, "should be", exp)
	}
}

func TestListEachInterrupt(t *testing.T) {
	nr := new(element)
	nr.add("h")
	nr.add("g")
	nr.add("f")
	nr.add("d")
	nr.add("c")
	nr.add("b")

	var a []Thing

	interrupted := nr.each(func(t Thing) bool {
		a = append(a, t)
		return len(a) == 2
	})

	if !interrupted {
		t.Error("Iteration should have been interrupted.")
	}

	if len(a) != 2 {
		t.Error("List should have 2 elements. Have", len(a))
	}
}

func TestPushBefore(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(element)
	nr.add("h")
	nr.add("g")
	nr.add("f")
	nr.add("d")
	nr.add("c")
	nr.add("b")
	element := &element{}
	if nr.addBefore("a", element, nr) {
		t.Error("should not be possible")
	}
	if !nr.addBefore("a", element, nr.next()) {
		t.Error("should be possible")
	}
}

func TestSearch(t *testing.T) {
	nr := &element{nil, &list_head}
	nr.add(c(9))
	nr.add(c(8))
	nr.add(c(7))
	nr.add(c(5))
	nr.add(c(4))
	nr.add(c(3))
	assertSlicey(t, nr, []Thing{&list_head, c(3), c(4), c(5), c(7), c(8), c(9)})
	searchTest(t, nr, c(1), &list_head, nil, c(3))
	searchTest(t, nr, c(2), &list_head, nil, c(3))
	searchTest(t, nr, c(3), &list_head, c(3), c(4))
	searchTest(t, nr, c(4), c(3), c(4), c(5))
	searchTest(t, nr, c(5), c(4), c(5), c(7))
	searchTest(t, nr, c(6), c(5), nil, c(7))
	searchTest(t, nr, c(7), c(5), c(7), c(8))
	searchTest(t, nr, c(8), c(7), c(8), c(9))
	searchTest(t, nr, c(9), c(8), c(9), nil)
	searchTest(t, nr, c(10), c(9), nil, nil)
	searchTest(t, nr, c(11), c(9), nil, nil)
}

func TestVerify(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := &element{nil, &list_head}
	nr.inject(c(3))
	nr.inject(c(5))
	nr.inject(c(9))
	nr.inject(c(7))
	nr.inject(c(4))
	nr.inject(c(8))
	assertSlicey(t, nr, []Thing{&list_head, c(3), c(4), c(5), c(7), c(8), c(9)})
	if err := nr.verify(); err != nil {
		t.Error(nr, "should verify as ok, got", err)
	}
	nr = &element{nil, &list_head}
	nr.add(c(3))
	nr.add(c(5))
	nr.add(c(9))
	nr.add(c(7))
	nr.add(c(4))
	nr.add(c(8))
	assertSlicey(t, nr, []Thing{&list_head, c(8), c(4), c(7), c(9), c(5), c(3)})
	s := fmt.Sprintf("[%v 8 4 7 9 5 3] is badly ordered. The following elements are in the wrong order: 8,4; 9,5; 5,3", &list_head)
	if err := nr.verify(); err.Error() != s {
		t.Error(nr, "should have errors", s, "but had", err)
	}
}

func TestInjectAndSearch(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := &element{nil, &list_head}
	nr.inject(c(3))
	nr.inject(c(5))
	nr.inject(c(9))
	nr.inject(c(7))
	nr.inject(c(4))
	nr.inject(c(8))
	assertSlicey(t, nr, []Thing{&list_head, c(3), c(4), c(5), c(7), c(8), c(9)})
	searchTest(t, nr, c(1), &list_head, nil, c(3))
	searchTest(t, nr, c(2), &list_head, nil, c(3))
	searchTest(t, nr, c(3), &list_head, c(3), c(4))
	searchTest(t, nr, c(4), c(3), c(4), c(5))
	searchTest(t, nr, c(5), c(4), c(5), c(7))
	searchTest(t, nr, c(6), c(5), nil, c(7))
	searchTest(t, nr, c(7), c(5), c(7), c(8))
	searchTest(t, nr, c(8), c(7), c(8), c(9))
	searchTest(t, nr, c(9), c(8), c(9), nil)
	searchTest(t, nr, c(10), c(9), nil, nil)
	searchTest(t, nr, c(11), c(9), nil, nil)
}

func TestConcInjectAndSearch(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := &element{nil, &list_head}
	nr.inject(c(3))
	nr.inject(c(5))
	nr.inject(c(9))
	nr.inject(c(7))
	nr.inject(c(4))
	nr.inject(c(8))
	assertSlicey(t, nr, []Thing{&list_head, c(3), c(4), c(5), c(7), c(8), c(9)})
	do := make(chan bool)
	ichan := make(chan []c)
	rchan := make(chan []c)
	var injected [][]c
	var removed [][]c
	for i := 0; i < runtime.NumCPU(); i++ {
		go fiddleAndAssertSort(t, nr, do, ichan, rchan)
	}
	close(do)
	for i := 0; i < runtime.NumCPU(); i++ {
		searchTest(t, nr, c(1), ANY, ANY, c(3))
		searchTest(t, nr, c(2), ANY, ANY, c(3))
		searchTest(t, nr, c(3), ANY, c(3), c(4))
		searchTest(t, nr, c(4), c(3), c(4), c(5))
		searchTest(t, nr, c(5), c(4), c(5), c(7))
		searchTest(t, nr, c(6), c(5), nil, c(7))
		searchTest(t, nr, c(7), c(5), c(7), c(8))
		searchTest(t, nr, c(8), c(7), c(8), c(9))
		searchTest(t, nr, c(9), c(8), c(9), nil)
		searchTest(t, nr, c(10), c(9), nil, nil)
		searchTest(t, nr, c(11), c(9), nil, nil)
		injected = append(injected, <-ichan)
		removed = append(removed, <-rchan)
	}
	assertSlicey(t, nr, []Thing{&list_head, c(3), c(4), c(5), c(7), c(8), c(9)})
	imap := make(map[c]int)
	for _, vals := range injected {
		for _, val := range vals {
			imap[val] = imap[val] + 1
		}
	}
	rmap := make(map[c]int)
	for _, vals := range removed {
		for _, val := range vals {
			rmap[val] = rmap[val] + 1
		}
	}
	for val, num := range imap {
		if num2, ok := rmap[val]; ok {
			if num2 != num {
				t.Errorf("fiddlers injected %v of %v but removed %v", num, val, num2)
			}
		} else {
			t.Errorf("fiddlers injected %v of %v but removed none", num, val)
		}
	}
	for val, num := range rmap {
		if num2, ok := imap[val]; ok {
			if num2 != num {
				t.Errorf("fiddlers removed %v of %v but injected %v", num, val, num2)
			}
		} else {
			t.Errorf("fiddlers removed %v of %v but injected none", num, val)
		}
	}
}
