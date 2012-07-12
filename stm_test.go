package gotomic

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
)

func compStrings(i, j string) int {
	l := len(i)
	if len(j) < l {
		l = len(j)
	}
	for ind := 0; ind < l; ind++ {
		if i[ind] < j[ind] {
			return -1
		} else if i[ind] > j[ind] {
			return 1
		}
	}
	if len(i) < len(j) {
		return -1
	} else if len(i) > len(j) {
		return 1
	}
	return 0
}

type testNode struct {
	value string
	left  *testNodeHandle
	right *testNodeHandle
}

func (self *testNode) Clone() Clonable {
	rval := *self
	return &rval
}

type testNodeHandle Handle

func newTestNodeHandle(v string) *testNodeHandle {
	return (*testNodeHandle)(NewHandle(&testNode{v, nil, nil}))
}
func (handle *testNodeHandle) getNode(t *Transaction) *testNode {
	node, err := handle.readNode(t)
	if err != nil {
		panic(err)
	}
	return node
}
func (handle *testNodeHandle) readNode(t *Transaction) (*testNode, error) {
	node, err := t.Read((*Handle)(handle))
	if err != nil {
		return nil, err
	}
	return node.(*testNode), nil
}
func (handle *testNodeHandle) writeNode(t *Transaction) (*testNode, error) {
	node, err := t.Write((*Handle)(handle))
	if err != nil {
		return nil, err
	}
	return node.(*testNode), nil
}
func (handle *testNodeHandle) insert(t *Transaction, v string) error {
	self, err := handle.writeNode(t)
	if err != nil {
		return err
	}
	cmp := compStrings(self.value, v)
	if cmp > 0 {
		if self.left == nil {
			self.left = newTestNodeHandle(v)
		} else {
			if err := self.left.insert(t, v); err != nil {
				return err
			}
		}
	} else if cmp < 0 {
		if self.right == nil {
			self.right = newTestNodeHandle(v)
		} else {
			if err := self.right.insert(t, v); err != nil {
				return err
			}
		}
	}
	return nil
}
func (handle *testNodeHandle) indentString(t *Transaction, i int) string {
	self, err := handle.readNode(t)
	if err != nil {
		return err.Error()
	}
	buf := new(bytes.Buffer)
	for j := 0; j < i; j++ {
		fmt.Fprint(buf, " ")
	}
	fmt.Fprintf(buf, "%#v", self)
	if self.left != nil {
		fmt.Fprintf(buf, "\nl:%v", self.left.indentString(t, i+1))
	}
	if self.right != nil {
		fmt.Fprintf(buf, "\nr:%v", self.right.indentString(t, i+1))
	}
	return string(buf.Bytes())
}
func (self *testNodeHandle) String() string {
	return self.indentString(NewTransaction(), 0)
}

func TestSTMBasicTestTree(t *testing.T) {
	hc := newTestNodeHandle("c")
	tr := NewTransaction()
	if err := hc.insert(tr, "a"); err != nil {
		t.Errorf("%v should insert 'a' but got %v", hc, err)
	}
	if err := hc.insert(tr, "d"); err != nil {
		t.Errorf("%v should insert 'd' but got %v", hc, err)
	}
	if err := hc.insert(tr, "b"); err != nil {
		t.Errorf("%v should insert 'b' but got %v", hc, err)
	}
	tr.Commit()
	tr = NewTransaction()
	nc := hc.getNode(tr)
	if nc.value != "c" {
		t.Error("bad value")
	}
	nd := nc.right.getNode(tr)
	if nd.value != "d" {
		t.Error("bad value")
	}
	na := nc.left.getNode(tr)
	if na.value != "a" {
		t.Error("bad value")
	}
	nb := na.right.getNode(tr)
	if nb.value != "b" {
		t.Error("bad value")
	}
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

func TestSTMIsolation(t *testing.T) {
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

func TestSTMReadBreakage(t *testing.T) {
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

func TestSTMDiffTrans1(t *testing.T) {
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

func TestSTMDiffTrans2(t *testing.T) {
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

func TestSTMDiffTrans3(t *testing.T) {
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
	<-do
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

func TestSTMTransConcurrency(t *testing.T) {
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
		<-done
	}
}

func TestSTMCommit(t *testing.T) {
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
