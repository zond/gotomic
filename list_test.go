package gotomic

import (
	"fmt"
	"testing"
	"reflect"
	"runtime"
	"math/rand"
)

func fiddle(n string, nr *nodeRef, do, done chan bool) {
	<- do
	num := 10000
	for i := 0; i < num; i++ {
		nr.push(fmt.Sprint(n, rand.Int()))
	}
	for i := 0; i < num; i++ {
		nr.pop()
	}
	done <- true
}

func fiddleAndAssertSort(t *testing.T, nr *nodeRef, do, done chan bool) {
	<- do
	num := 1000
	for i := 0; i < num; i++ {
		nr.inject(c(rand.Int()))
		stuff := nr.toSlice()
		var last Comparable
		for _, item := range stuff {
			if s, ok := item.(Comparable); ok {
				if s.Compare(last) < 0 {
					t.Error(nr, " should be sorted, but ", s, " should really be before ", last)
				}
				last = s
			} else {
				t.Error(nr, " should contain only Comparable items, but found ", item)
			}
		}
	}
	for i := 0; i < num; i++ {
		nr.pop()
	}
	done <- true
}

func assertSlicey(t *testing.T, nr *nodeRef, cmp []thing) {
	sl := nr.toSlice()
	if len(sl) != len(cmp) {
		t.Error(nr, ".toSlice() should be ", cmp, " but is ", sl)
	}
	for index, th := range cmp {
		if !reflect.DeepEqual(sl[index], th) {
			t.Error(nr, ".toSlice()[", index, "] should be ", th, " but is ", sl[index])
		}
	}
}

func assertPop(t *testing.T, nr *nodeRef, th thing) {
	p := nr.pop()
	if !reflect.DeepEqual(p, th) {
		t.Error(nr, " should pop ", th, " but popped ", p)
	}
}

func TestPushPop(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(nodeRef)
	assertSlicey(t, nr, []thing{})
	nr.push("hej")
	assertSlicey(t, nr, []thing{"hej"})
	nr.push("haj")
	assertSlicey(t, nr, []thing{"haj","hej"})
	nr.push("hoj")
	assertSlicey(t, nr, []thing{"hoj","haj","hej"})
	assertPop(t, nr, "hoj")
	assertSlicey(t, nr, []thing{"haj","hej"})
	assertPop(t, nr, "haj")
	assertSlicey(t, nr, []thing{"hej"})
	assertPop(t, nr, "hej")
	assertSlicey(t, nr, []thing{})
	assertPop(t, nr, nil)
	assertSlicey(t, nr, []thing{})
	nr.push("1")
	nr.push("2")
	nr.push("3")
	nr.push("4")
	do := make(chan bool)
	done := make(chan bool)
	go fiddle("a", nr, do, done)
	go fiddle("b", nr, do, done)
	go fiddle("b", nr, do, done)
	go fiddle("b", nr, do, done)
	close(do)
	<-done
	<-done
	<-done
	<-done
	assertSlicey(t, nr, []thing{"4","3","2","1"})
}

type c int
func (self c) Compare(t thing) int {
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
	return 0
}

const ANY = "ANY VALUE"

func searchTest(t *testing.T, nr *nodeRef, s c, wb,wm,wa thing) {
	b, m, a := nr.search(s)
	if (wb != ANY && !reflect.DeepEqual(b.val(), wb)) || 
		(wm != ANY && !reflect.DeepEqual(m.val(), wm)) || 
		(wa != ANY && !reflect.DeepEqual(a.val(), wa)) {
		t.Error(nr, ".search(", s, ") should produce ", wb, wm, wa, " but produced ", b.val(), m.val(), a.val())
	}
}

func TestPushBefore(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(nodeRef)
	nr.push("h")
	nr.push("g")
	nr.push("f")
	nr.push("d")
	nr.push("c")
	nr.push("b")
	if nr.pushBefore("a", nr.node().next.node()) {
		t.Error("should not be possible")
	}
	if !nr.pushBefore("a", nr.node()) {
		t.Error("should be possible")
	}
}

func TestInject(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(nodeRef)
	nr.inject(c(5))
	nr.inject(c(1))
	nr.inject(c(3))
	nr.inject(c(8))
	nr.inject(c(4))
	assertSlicey(t, nr, []thing{c(1),c(3),c(4),c(5),c(8)})
	do := make(chan bool)
	done := make(chan bool)
	go fiddleAndAssertSort(t, nr, do, done)
	go fiddleAndAssertSort(t, nr, do, done)
	go fiddleAndAssertSort(t, nr, do, done)
	go fiddleAndAssertSort(t, nr, do, done)
	close(do)
	<-done
	<-done
	<-done
	<-done
}

func TestSearch(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(nodeRef)
	nr.push(8)
	nr.push(7)
	nr.push(6)
	nr.push(4)
	nr.push(3)
	nr.push(2)
	searchTest(t, nr, c(1), nil, nil, 2)
	searchTest(t, nr, c(2), nil, 2, 3)
	searchTest(t, nr, c(3), 2, 3, 4)
	searchTest(t, nr, c(4), 3, 4, 6)
	searchTest(t, nr, c(5), 4, nil, 6)
	searchTest(t, nr, c(6), 4, 6, 7)
	searchTest(t, nr, c(7), 6, 7, 8)
	searchTest(t, nr, c(8), 7, 8, nil)
	searchTest(t, nr, c(9), 8, nil, nil)
	do := make(chan bool)
	done := make(chan bool)
	go fiddle("a1", nr, do, done)
	go fiddle("a2", nr, do, done)
	go fiddle("a3", nr, do, done)
	go fiddle("a4", nr, do, done)
	close(do)
	searchTest(t, nr, c(1), ANY, nil, 2)
	searchTest(t, nr, c(2), ANY, 2, 3)
	searchTest(t, nr, c(3), 2, 3, 4)
	searchTest(t, nr, c(4), 3, 4, 6)
	searchTest(t, nr, c(5), 4, nil, 6)
	searchTest(t, nr, c(6), 4, 6, 7)
	searchTest(t, nr, c(7), 6, 7, 8)
	searchTest(t, nr, c(8), 7, 8, nil)
	searchTest(t, nr, c(9), 8, nil, nil)
	<-done
	<-done
	<-done
	<-done
}
