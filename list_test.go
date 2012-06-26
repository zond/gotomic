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

