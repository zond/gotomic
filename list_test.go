package gotomic

import (
	"testing"
	"reflect"
	"runtime"
	"math/rand"
	"time"
	"math"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func fiddle(nr *node, do, done chan bool) {
	<- do
	num := 10000
	for i := 0; i < num; i++ {
		nr.add(rand.Int())
	}
	for i := 0; i < num; i++ {
		nr.remove()
	}
	done <- true
}

func fiddleAndAssertSort(t *testing.T, nr *node, do, done chan bool) {
	<- do
	num := 1000
	for i := 0; i < num; i++ {
		nr.inject(c(-int(math.Abs(float64(rand.Int())))))
		if err := nr.verify(); err != nil {
			t.Error(nr, "should be correct, but got", err)
		}
	}
	for i := 0; i < num; i++ {
		nr.remove()
	}
	done <- true
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
		popped := l.Pop() 
		tmp[len(cmp) - ind - 1] = popped
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
	assertListy(t, l, []Thing{"blar", "hehu","knap","plur"})
}

func assertSlicey(t *testing.T, nr *node, cmp []Thing) {
	if sl := nr.ToSlice(); !reflect.DeepEqual(sl, cmp) {
		t.Errorf("%v should be %#v but is %#v", nr, cmp, sl)
	}
}

func assertPop(t *testing.T, nr *node, th Thing) {
	p := nr.remove()
	if !reflect.DeepEqual(p, th) {
		t.Error(nr, " should pop ", th, " but popped ", p)
	}
}

func TestPushPop(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(node)
	assertSlicey(t, nr, []Thing{nil})
	nr.add("hej")
	assertSlicey(t, nr, []Thing{nil,"hej"})
	nr.add("haj")
	assertSlicey(t, nr, []Thing{nil,"haj","hej"})
	nr.add("hoj")
	assertSlicey(t, nr, []Thing{nil,"hoj","haj","hej"})
	assertPop(t, nr, "hoj")
	assertSlicey(t, nr, []Thing{nil,"haj","hej"})
	assertPop(t, nr, "haj")
	assertSlicey(t, nr, []Thing{nil,"hej"})
	assertPop(t, nr, "hej")
	assertSlicey(t, nr, []Thing{nil})
	assertPop(t, nr, nil)
	assertSlicey(t, nr, []Thing{nil})
	nr.add("1")
	nr.add("2")
	nr.add("3")
	nr.add("4")
	do := make(chan bool)
	done := make(chan bool)
	go fiddle(nr, do, done)
	go fiddle(nr, do, done)
	go fiddle(nr, do, done)
	go fiddle(nr, do, done)
	close(do)
	<-done
	<-done
	<-done
	<-done
	assertSlicey(t, nr, []Thing{nil,"4","3","2","1"})
}

type c int
func (self c) Compare(t Thing) int {
	if s, ok := t.(int); ok {
		if int(self) > s {
			return 1
		} else if int(self) < s {
			return -1
		}
	}
	if s, ok := t.(c); ok {
		if self > s {
			return 1
		} else if self < s {
			return -1
		}
	}
	if t == nil {
		return 1
	}
	return 0
}

const ANY = "ANY VALUE"

func searchTest(t *testing.T, nr *node, s c, l, n, r Thing) {
	h := nr.search(s)
	if (l != ANY && !reflect.DeepEqual(h.left.val(), l)) || 
		(n != ANY && !reflect.DeepEqual(h.node.val(), n)) || 
		(r != ANY && !reflect.DeepEqual(h.right.val(), r)) {
		t.Error(nr, ".search(", s, ") should produce ", l, n, r, " but produced ", h.left.val(), h.node.val(), h.right.val())
	}
}

func TestPushBefore(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(node)
	nr.add("h")
	nr.add("g")
	nr.add("f")
	nr.add("d")
	nr.add("c")
	nr.add("b")
	node := &node{}
	if nr.addBefore("a", node, nr) {
		t.Error("should not be possible")
	}
	if !nr.addBefore("a", node, nr.next()) {
		t.Error("should be possible")
	}
}

func TestSearch(t *testing.T) {
	nr := new(node)
	nr.add(c(9))
	nr.add(c(8))
	nr.add(c(7))
	nr.add(c(5))
	nr.add(c(4))
	nr.add(c(3))
	assertSlicey(t, nr, []Thing{nil, c(3),c(4),c(5),c(7),c(8),c(9)})
	searchTest(t, nr, c(1), nil, nil, c(3))
	searchTest(t, nr, c(2), nil, nil, c(3))
	searchTest(t, nr, c(3), nil, c(3), c(4))
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
	nr := new(node)
	nr.inject(c(3))
	nr.inject(c(5))
	nr.inject(c(9))
	nr.inject(c(7))
	nr.inject(c(4))
	nr.inject(c(8))
	assertSlicey(t, nr, []Thing{nil, c(3),c(4),c(5),c(7),c(8),c(9)})
	if err := nr.verify(); err != nil {
		t.Error(nr, "should verify as ok, got", err)
	}
	nr = new(node)
	nr.add(c(3))
	nr.add(c(5))
	nr.add(c(9))
	nr.add(c(7))
	nr.add(c(4))
	nr.add(c(8))
	assertSlicey(t, nr, []Thing{nil, c(8),c(4),c(7),c(9),c(5),c(3)})
	s := "[<nil> 8 4 7 9 5 3] is badly ordered. The following nodes are in the wrong order: 8,4; 9,5; 5,3"
	if err := nr.verify(); err.Error() != s {
		t.Error(nr, "should have errors", s, "but had", err)
	}
}

func TestInjectAndSearch(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(node)
	nr.inject(c(3))
	nr.inject(c(5))
	nr.inject(c(9))
	nr.inject(c(7))
	nr.inject(c(4))
	nr.inject(c(8))
	assertSlicey(t, nr, []Thing{nil, c(3),c(4),c(5),c(7),c(8),c(9)})
	searchTest(t, nr, c(1), nil, nil, c(3))
	searchTest(t, nr, c(2), nil, nil, c(3))
	searchTest(t, nr, c(3), nil, c(3), c(4))
	searchTest(t, nr, c(4), c(3), c(4), c(5))
	searchTest(t, nr, c(5), c(4), c(5), c(7))
	searchTest(t, nr, c(6), c(5), nil, c(7))
	searchTest(t, nr, c(7), c(5), c(7), c(8))
	searchTest(t, nr, c(8), c(7), c(8), c(9))
	searchTest(t, nr, c(9), c(8), c(9), nil)
	searchTest(t, nr, c(10), c(9), nil, nil)
	searchTest(t, nr, c(11), c(9), nil, nil)
	do := make(chan bool)
	done := make(chan bool)
	go fiddleAndAssertSort(t, nr, do, done)
	go fiddleAndAssertSort(t, nr, do, done)
	go fiddleAndAssertSort(t, nr, do, done)
	go fiddleAndAssertSort(t, nr, do, done)
	close(do)
	for i := 0; i < 4; i++ {
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
		<-done
	}
	assertSlicey(t, nr, []Thing{nil, c(3),c(4),c(5),c(7),c(8),c(9)})
}
