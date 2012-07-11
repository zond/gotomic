
package gotomic

import (
	"testing"
	"runtime"
	"fmt"
)

type testNode struct {
	value string
	left *testNode
	right *testNode
}
func (self *testNode) Clone() Clonable {
	rval := *self
	return &rval
}

func tWrite(t *testing.T, tr *Transaction, h *Handle) Thing {
	x, err := tr.Write(h)
	if err != nil {
		t.Errorf("%v should be able to write %v, but got %v", tr, h, err)
	}
	return x
}

func TestIsolation(t *testing.T) {
	h := NewHandle(&testNode{"a", nil, nil})
	tr := NewTransaction()
	n := tWrite(t, tr, h).(*testNode)
	if n.value != "a" {
		t.Errorf("%v should be 'a'", n.value)
	}
	n.value = "b"
	if n.value != "b" {
		t.Errorf("%v should be 'b'", n.value)
	}
	tr2 := NewTransaction()
	n2 := tWrite(t, tr2, h).(*testNode)
	if n2.value != "a" {
		t.Errorf("%v should be 'a'", n2.value)
	}
	n2.value = "c"
	if n2.value != "c" {
		t.Errorf("%v should be 'c'", n2.value)
	}
}

func TestReadBreakage(t *testing.T) {
	h := NewHandle(&testNode{"a", nil, nil})
	tr := NewTransaction()
	tr2 := NewTransaction()
	n2 := tWrite(t, tr2, h).(*testNode)
	if n2.value != "a" {
		t.Errorf("%v should be 'a'", n2.value)
	}
	if !tr2.Commit() {
		t.Errorf("%v should commit!")
	}
	n, err := tr.Write(h)
	if err == nil {
		t.Errorf("%v should not allow reading of %v, but got %v", tr, h, n)
	}
}

func fiddleTrans(t *testing.T, x string, h1, h2 *Handle, do, done chan bool) {
	<- do
	for i := 0; i < 100; i++ {
		tr := NewTransaction()
		n1, err1 := tr.Write(h1)
		n2, err2 := tr.Write(h2)
		if err1 == nil && err2 == nil {
			if n1.(*testNode).value != n2.(*testNode).value {
				t.Errorf("%v, %v: %v should == %v", x, i, n1, n2)
			}
			n1.(*testNode).value = x
			n2.(*testNode).value = x
			tr = NewTransaction()
			n1, err1 = tr.Read(h1)
			n2, err2 = tr.Read(h2)
			if err1 == nil && err2 == nil && n1.(*testNode).value != n2.(*testNode).value {
				t.Errorf("%v, %v: %v should == %v", x, i, n1, n2)
			}
		}
	}
	done <- true
}

func TestTransConcurrency(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	do := make(chan bool)
	done := make(chan bool)
	h1 := NewHandle(&testNode{"a", nil, nil})
	h2 := NewHandle(&testNode{"a", nil, nil})
	for i := 0; i < runtime.NumCPU(); i++ {
		go fiddleTrans(t, fmt.Sprint(i), h1, h2, do, done)
	}
	close(do)
	for i := 0; i < runtime.NumCPU(); i++ {
		<- done
	}
}

func TestCommit(t *testing.T) {
	h := NewHandle(&testNode{"a", nil, nil})
	tr := NewTransaction()
	n := tWrite(t, tr, h).(*testNode)
	if n.value != "a" {
		t.Errorf("%v should be 'a'", n.value)
	}
	n.value = "b"
	if n.value != "b" {
		t.Errorf("%v should be 'b'", n.value)
	}
	tr2 := NewTransaction()
	n2 := tWrite(t, tr2, h).(*testNode)
	if n2.value != "a" {
		t.Errorf("%v should be 'a'", n2.value)
	}
	n2.value = "c"
	if n2.value != "c" {
		t.Errorf("%v should be 'c'", n2.value)
	}
	if !tr.Commit() {
		t.Errorf("%v should commit", tr)
	}
	tr3 := NewTransaction()
	n3 := tWrite(t, tr3, h).(*testNode)
	if n3.value != "b" {
		t.Errorf("%v should be 'b'", n3.value)
	}
	if n2.value != "c" {
		t.Errorf("%v should be 'c'", n2.value)
	}
	if tr2.Commit() {
		t.Errorf("%v should not commit", tr2)
	}
	tr4 := NewTransaction()
	n4 := tWrite(t, tr4, h).(*testNode)
	if n4.value != "b" {
		t.Errorf("%v should be 'b'", n4.value)
	}
}
