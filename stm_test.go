
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

func tRead(t *testing.T, tr *Transaction, h *Handle) Thing {
	x, err := tr.Read(h)
	if err != nil {
		t.Errorf("%v should be able to read %v, but got %v", tr, h, err)
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

func TestDiffTrans1(t *testing.T) {
	tr1 := NewTransaction()
	tr2 := NewTransaction()
	h1 := NewHandle(&testNode{"a", nil, nil})
	h2 := NewHandle(&testNode{"b", nil, nil})
	h3 := NewHandle(&testNode{"c", nil, nil})
	n11 := tRead(t, tr1, h1).(*testNode)
	n12 := tRead(t, tr1, h2).(*testNode)
	n22 := tRead(t, tr2, h2).(*testNode)
	n23 := tRead(t, tr2, h3).(*testNode)
	if n11.value != "a" {
		t.Error("bad value")
	}
	if n12.value != "b" {
		t.Error("bad value")
	}
	if n22.value != "b" {
		t.Error("bad value")
	}
	if n23.value != "c" {
		t.Error("bad value")
	}
	if !tr1.Commit() {
		t.Error("should commit")
	}
	if !tr2.Commit() {
		t.Error("should commit")
	}
}

func TestDiffTrans2(t *testing.T) {
	tr1 := NewTransaction()
	tr2 := NewTransaction()
	h1 := NewHandle(&testNode{"a", nil, nil})
	h2 := NewHandle(&testNode{"b", nil, nil})
	h3 := NewHandle(&testNode{"c", nil, nil})
	n11 := tWrite(t, tr1, h1).(*testNode)
	n12 := tRead(t, tr1, h2).(*testNode)
	n22 := tRead(t, tr2, h2).(*testNode)
	n23 := tWrite(t, tr2, h3).(*testNode)
	if n11.value != "a" {
		t.Error("bad value")
	}
	if n12.value != "b" {
		t.Error("bad value")
	}
	if n22.value != "b" {
		t.Error("bad value")
	}
	if n23.value != "c" {
		t.Error("bad value")
	}
	n11.value = "a2"
	n23.value = "c2"
	if !tr1.Commit() {
		t.Error("should commit")
	}
	if !tr2.Commit() {
		t.Error("should commit")
	}
	tr3 := NewTransaction()
	if tRead(t, tr3, h1).(*testNode).value != "a2" {
		t.Error("bad value")
	}
	if tRead(t, tr3, h3).(*testNode).value != "c2" {
		t.Error("bad value")
	}
}

func TestDiffTrans3(t *testing.T) {
	tr1 := NewTransaction()
	tr2 := NewTransaction()
	h1 := NewHandle(&testNode{"a", nil, nil})
	h2 := NewHandle(&testNode{"b", nil, nil})
	h3 := NewHandle(&testNode{"c", nil, nil})
	n11 := tWrite(t, tr1, h1).(*testNode)
	n12 := tWrite(t, tr1, h2).(*testNode)
	n22 := tWrite(t, tr2, h2).(*testNode)
	n23 := tWrite(t, tr2, h3).(*testNode)
	if n11.value != "a" {
		t.Error("bad value")
	}
	if n12.value != "b" {
		t.Error("bad value")
	}
	if n22.value != "b" {
		t.Error("bad value")
	}
	if n23.value != "c" {
		t.Error("bad value")
	}
	n12.value = "b2"
	n22.value = "b3"
	if !tr1.Commit() {
		t.Error("should commit")
	}
	if tr2.Commit() {
		t.Error("should not commit")
	}
	tr3 := NewTransaction()
	if tRead(t, tr3, h2).(*testNode).value != "b2" {
		t.Error("bad value")
	}
}

func fiddleTrans(t *testing.T, x string, h1, h2 *Handle, do, done chan bool) {
	<- do
	for i := 0; i < 10000; i++ {
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
