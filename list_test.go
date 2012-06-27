package gotomic

import (
	"fmt"
	"testing"
	"reflect"
	"runtime"
)

func fiddle(n string, nr *nodeRef, do, done chan bool) {
	<- do
	for i := 0; i < 10000; i++ {
		nr.push(fmt.Sprint(n, i))
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

type c string
func (self c) Compare(t thing) int {
	if s, ok := t.(string); ok {
		if self[0] > s[0] {
			return 1
		} else if self[0] < s[0] {
			return -1
		}
	}
	if s, ok := t.(c); ok {
		if self[0] > s[0] {
			return 1
		} else if self[0] < s[0] {
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
	nr.inject(c("h"))
	nr.inject(c("a"))
	nr.inject(c("b"))
	nr.inject(c("x"))
	nr.inject(c("d"))
	assertSlicey(t, nr, []thing{c("a"),c("b"),c("d"),c("h"),c("x")})
}

func TestSearch(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	nr := new(nodeRef)
	nr.push("h")
	nr.push("g")
	nr.push("f")
	nr.push("d")
	nr.push("c")
	nr.push("b")
	searchTest(t, nr, "a", nil, nil, "b")
	searchTest(t, nr, "b", nil, "b", "c")
	searchTest(t, nr, "c", "b", "c", "d")
	searchTest(t, nr, "d", "c", "d", "f")
	searchTest(t, nr, "e", "d", nil, "f")
	searchTest(t, nr, "f", "d", "f", "g")
	searchTest(t, nr, "g", "f", "g", "h")
	searchTest(t, nr, "h", "g", "h", nil)
	searchTest(t, nr, "i", "h", nil, nil)
	do := make(chan bool)
	done := make(chan bool)
	go fiddle("a1", nr, do, done)
	go fiddle("a2", nr, do, done)
	go fiddle("a3", nr, do, done)
	go fiddle("a4", nr, do, done)
	close(do)
	searchTest(t, nr, "a", ANY, nil, "b")
	searchTest(t, nr, "b", ANY, "b", "c")
	searchTest(t, nr, "c", "b", "c", "d")
	searchTest(t, nr, "d", "c", "d", "f")
	searchTest(t, nr, "e", "d", nil, "f")
	searchTest(t, nr, "f", "d", "f", "g")
	searchTest(t, nr, "g", "f", "g", "h")
	searchTest(t, nr, "h", "g", "h", nil)
	searchTest(t, nr, "i", "h", nil, nil)
	<-done
	<-done
	<-done
	<-done
}
