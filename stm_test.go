
package gotomic

import (
	"testing"
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
