package gotomic

import (
	"bytes"
	"fmt"
	"math/rand"
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
	if i == j {
		return 0
	}
	panic(fmt.Errorf("wtf, %v and %v are not the same!", i, j))
}

type testNode struct {
	value  string
	parent *testNodeHandle
	left   *testNodeHandle
	right  *testNodeHandle
}

func (self *testNode) Clone() Clonable {
	return &testNode{fmt.Sprint(self.value), self.parent, self.left, self.right}
}

type testNodeHandle Handle

func newTestNodeHandle(v string, parent *testNodeHandle) *testNodeHandle {
	return (*testNodeHandle)(NewHandle(&testNode{v, parent, nil, nil}))
}
func (handle *testNodeHandle) has(t *Transaction, v string) (bool, error) {
	self, err := handle.readNode(t)
	if err != nil {
		return false, err
	}
	cmp := compStrings(self.value, v)
	if cmp > 0 {
		if self.left != nil {
			return self.left.has(t, v)
		}
		return false, nil
	} else if cmp < 0 {
		if self.right != nil {
			return self.right.has(t, v)
		}
		return false, nil
	}
	if self.value == v {
		return true, nil
	}
	return false, nil
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
func (handle *testNodeHandle) remove(t *Transaction, v string) (ok bool, err error) {
	self, err := handle.readNode(t)
	if err != nil {
		return
	}
	cmp := compStrings(self.value, v)
	if cmp > 0 {
		if self.left != nil {
			return self.left.remove(t, v)
		}
	} else if cmp < 0 {
		if self.right != nil {
			return self.right.remove(t, v)
		}
	} else {
		if self.right != nil && self.left != nil {
			var succN *testNode
			if rand.Float32() < 0.5 {
				successor, err := self.left.farRight(t)
				if err != nil {
					return false, err
				}
				succN, err = successor.readNode(t)
				if err != nil {
					return false, err
				}
				if err = successor.replaceInParent(t, succN.left); err != nil {
					return false, err
				}
			} else {
				successor, err := self.right.farLeft(t)
				if err != nil {
					return false, err
				}
				succN, err = successor.readNode(t)
				if err != nil {
					return false, err
				}
				if err = successor.replaceInParent(t, succN.right); err != nil {
					return false, err
				}
			}
			self, err = handle.writeNode(t)
			if err != nil {
				return false, err
			}
			self.value = succN.value
			ok = true
		} else if self.right != nil {
			if err = handle.replaceInParent(t, self.right); err != nil {
				return false, err
			}
			ok = true
		} else if self.left != nil {
			if err = handle.replaceInParent(t, self.left); err != nil {
				return false, err
			}
			ok = true
		} else {
			if err = handle.replaceInParent(t, nil); err != nil {
				return false, err
			}
			ok = true
		}
	}
	return
}
func (handle *testNodeHandle) farRight(t *Transaction) (rval *testNodeHandle, err error) {
	self, err := handle.readNode(t)
	if err != nil {
		return nil, err
	}
	if self.right == nil {
		return handle, nil
	}
	return self.right.farRight(t)
}
func (handle *testNodeHandle) farLeft(t *Transaction) (rval *testNodeHandle, err error) {
	self, err := handle.readNode(t)
	if err != nil {
		return nil, err
	}
	if self.left == nil {
		return handle, nil
	}
	return self.left.farLeft(t)
}
func (handle *testNodeHandle) replaceInParent(t *Transaction, neu *testNodeHandle) error {
	self, err := handle.readNode(t)
	if err != nil {
		return err
	}
	if self.parent == nil {
		return fmt.Errorf("%#v.replaceInParent(...): I have no parent!")
	}
	parent, err := self.parent.writeNode(t)
	if err != nil {
		return err
	}
	if parent.left == handle {
		parent.left = neu
	} else if parent.right == handle {
		parent.right = neu
	} else {
		panic(fmt.Errorf("%#v.replaceInParent(...): I don't seem to exist in my parent: %#v", self, parent))
	}
	if neu != nil {
		n, err := neu.writeNode(t)
		if err != nil {
			return err
		}
		n.parent = self.parent
	}
	return nil
}

func (handle *testNodeHandle) insert(t *Transaction, v string) error {
	self, err := handle.readNode(t)
	if err != nil {
		return err
	}
	cmp := compStrings(self.value, v)
	if cmp > 0 {
		if self.left == nil {
			self, err = handle.writeNode(t)
			if err != nil {
				return err
			}
			self.left = newTestNodeHandle(v, handle)
		} else {
			if err := self.left.insert(t, v); err != nil {
				return err
			}
		}
	} else if cmp < 0 {
		if self.right == nil {
			self, err = handle.writeNode(t)
			if err != nil {
				return err
			}
			self.right = newTestNodeHandle(v, handle)
		} else {
			if err := self.right.insert(t, v); err != nil {
				return err
			}
		}
	}
	return nil
}
func (handle *testNodeHandle) indentString(i int) string {
	self, _ := (*Handle)(handle).Current().(*testNode)
	buf := new(bytes.Buffer)
	for j := 0; j < i; j++ {
		fmt.Fprint(buf, " ")
	}
	fmt.Fprintf(buf, "%p:%#v", handle, self)
	if self.left != nil {
		fmt.Fprintf(buf, "\nl:%v", self.left.indentString(i+1))
	}
	if self.right != nil {
		fmt.Fprintf(buf, "\nr:%v", self.right.indentString(i+1))
	}
	return string(buf.Bytes())
}
func (self *testNodeHandle) String() string {
	return self.indentString(0)
}

type cmpNode struct {
	value string
	left  *cmpNode
	right *cmpNode
}

func assertTreeStructure(t *testing.T, h *testNodeHandle, c *cmpNode) {
	if c == nil {
		if h != nil {
			t.Error("should be nil, was %v", h)
		}
	} else {
		tr := NewTransaction()
		n, err := h.readNode(tr)
		if err != nil {
			t.Errorf("%v should be readable, got %v", h, err)
		}
		if c.value != n.value {
			t.Errorf("%v should have value %v, had %v", h, c.value, n.value)
		}
		assertTreeStructure(t, n.left, c.left)
		assertTreeStructure(t, n.right, c.right)
	}
}

func fiddleTestTree(t *testing.T, x string, h *testNodeHandle, do, done chan bool) {
	<-do
	n := int(10000 + rand.Int31()%1000)
	vals := make([]string, n)
	for i := 0; i < n; i++ {
		v := fmt.Sprint(rand.Int63(), ".", i, ".", x)
		vals[i] = v
		inserted := false
		for !inserted {
			tr := NewTransaction()
			if h.insert(tr, v) == nil {
				if tr.Commit() {
					inserted = true
					looked := false
					for !looked {
						tr = NewTransaction()
						ok, err := h.has(tr, v)
						if err == nil {
							looked = true
							if !ok {
								fmt.Printf("%v should contain %v\n", h, v)
								t.Fatalf("%v should contain %v\n", h, v)
							}
						}
					}
				}
			}
		}
	}
	for i := 0; i < n; i++ {
		v := vals[i]
		removed := false
		for !removed {
			tr := NewTransaction()
			ok, err := h.remove(tr, v)
			if err == nil {
				if ok {
					if tr.Commit() {
						removed = true
						looked := false
						for !looked {
							tr = NewTransaction()
							ok, err := h.has(tr, v)
							if err == nil {
								looked = true
								if ok {
									fmt.Printf("%v should not contain %v\n", h, v)
									t.Fatalf("%v should not contain %v\n", h, v)
								}
							}
						}
					}
				} else {
					fmt.Println("wtf, ", v, "is not in", h)
					t.Fatal("wtf, ", v, "is not in", h)
				}
			}
		}
	}
	done <- true
}

func TestSTMConcurrentTestTree(t *testing.T) {
	hc := newTestNodeHandle("c", nil)
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
	if !tr.Commit() {
		t.Errorf("%v should commit", tr)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"a", nil, &cmpNode{"b", nil, nil}}, &cmpNode{"d", nil, nil}})
	do := make(chan bool)
	done := make(chan bool)
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go fiddleTestTree(t, fmt.Sprint(i), hc, do, done)
	}
	close(do)
	for i := 0; i < runtime.NumCPU(); i++ {
		<-done
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"a", nil, &cmpNode{"b", nil, nil}}, &cmpNode{"d", nil, nil}})
}

func TestSTMBigTestTree(t *testing.T) {
	hc := newTestNodeHandle("c", nil)
	assertTreeStructure(t, hc, &cmpNode{"c", nil, nil})
	n := 10000
	vals := make([]string, n)
	for i := 0; i < n; i++ {
		tr := NewTransaction()
		v := fmt.Sprint(rand.Int63(), ".", i)
		vals[i] = v
		err := hc.insert(tr, v)
		if err == nil {
			if tr.Commit() {
				tr = NewTransaction()
				ok, err := hc.has(tr, v)
				if err == nil {
					if !ok {
						t.Errorf("%v should contain %v", hc, v)
					}
				} else {
					t.Errorf("%v should be able to look for %v", hc, v)
				}
			} else {
				t.Errorf("%v should commit", tr)
			}
		} else {
			t.Errorf("%v should insert %v but got %v", hc, v, err)
		}
	}
	for i := 0; i < n; i++ {
		tr := NewTransaction()
		v := vals[i]
		ok, err := hc.remove(tr, v)
		if err == nil {
			if ok {
				if tr.Commit() {
					tr = NewTransaction()
					ok, err := hc.has(tr, v)
					if err == nil {
						if ok {
							t.Errorf("%v should not contain %v", hc, v)
						}
					} else {
						t.Errorf("%v should be able to look for %v", hc, v)
					}
				} else {
					t.Errorf("%v should commit", tr)
				}
			} else {
				t.Errorf("%v should remove %#v", hc, v)
			}
		} else {
			t.Errorf("%v should remove %v but got %v", hc, v, err)
		}
	}
	assertTreeStructure(t, hc, &cmpNode{"c", nil, nil})
}

func TestSTMBasicTestTree(t *testing.T) {
	hc := newTestNodeHandle("c", nil)
	assertTreeStructure(t, hc, &cmpNode{"c", nil, nil})
	tr := NewTransaction()
	if err := hc.insert(tr, "a"); err != nil {
		t.Errorf("%v should insert 'a' but got %v", hc, err)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", nil, nil})
	if err := hc.insert(tr, "d"); err != nil {
		t.Errorf("%v should insert 'd' but got %v", hc, err)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", nil, nil})
	if err := hc.insert(tr, "b"); err != nil {
		t.Errorf("%v should insert 'b' but got %v", hc, err)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", nil, nil})
	if !tr.Commit() {
		t.Errorf("%v should commit", tr)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"a", nil, &cmpNode{"b", nil, nil}}, &cmpNode{"d", nil, nil}})
	tr = NewTransaction()
	ok, err := hc.remove(tr, "d")
	if !ok || err != nil {
		t.Errorf("%v should remove 'd' but got %v, %v", hc, ok, err)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"a", nil, &cmpNode{"b", nil, nil}}, &cmpNode{"d", nil, nil}})
	if !tr.Commit() {
		t.Errorf("%v should commit", tr)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"a", nil, &cmpNode{"b", nil, nil}}, nil})
	tr = NewTransaction()
	err = hc.insert(tr, "e")
	if err != nil {
		t.Errorf("%v should insert 'e' but got %v", hc, err)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"a", nil, &cmpNode{"b", nil, nil}}, nil})
	if !tr.Commit() {
		t.Errorf("%v should commit", tr)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"a", nil, &cmpNode{"b", nil, nil}}, &cmpNode{"e", nil, nil}})
	tr = NewTransaction()
	ok, err = hc.remove(tr, "a")
	if !ok || err != nil {
		t.Errorf("%v should remove 'a' but got %v, %v", hc, ok, err)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"a", nil, &cmpNode{"b", nil, nil}}, &cmpNode{"e", nil, nil}})
	if !tr.Commit() {
		t.Errorf("%v should commit", tr)
	}
	assertTreeStructure(t, hc, &cmpNode{"c", &cmpNode{"b", nil, nil}, &cmpNode{"e", nil, nil}})
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
	h := NewHandle(&testNode{"a", nil, nil, nil})
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
	h := NewHandle(&testNode{"a", nil, nil, nil})
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
	h1 := NewHandle(&testNode{"a", nil, nil, nil})
	h2 := NewHandle(&testNode{"b", nil, nil, nil})
	h3 := NewHandle(&testNode{"c", nil, nil, nil})
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
	h1 := NewHandle(&testNode{"a", nil, nil, nil})
	h2 := NewHandle(&testNode{"b", nil, nil, nil})
	h3 := NewHandle(&testNode{"c", nil, nil, nil})
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
	h1 := NewHandle(&testNode{"a", nil, nil, nil})
	h2 := NewHandle(&testNode{"b", nil, nil, nil})
	h3 := NewHandle(&testNode{"c", nil, nil, nil})
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
				t.Errorf("Write: thread %v, loop %v: %v should == %v in %v", x, i, n1, n2, tr.Describe())
			}
			newVal := fmt.Sprint(x, i)
			n1.(*testNode).value = newVal
			n2.(*testNode).value = newVal
			tr.Commit()
			tr = NewTransaction()
			n1, err1 = tr.Read(h1)
			n2, err2 = tr.Read(h2)
			if err1 == nil && err2 == nil && n1.(*testNode).value != n2.(*testNode).value {
				t.Errorf("Read: thread %v, loop %v: %v should == %v in %v", x, i, n1, n2, tr.Describe())
			}
		}
	}
	done <- true
}

func TestSTMTransConcurrency(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	do := make(chan bool)
	done := make(chan bool)
	h1 := NewHandle(&testNode{"a", nil, nil, nil})
	h2 := NewHandle(&testNode{"a", nil, nil, nil})
	for i := 0; i < runtime.NumCPU(); i++ {
		go fiddleTrans(t, fmt.Sprint(i), h1, h2, do, done)
	}
	close(do)
	for i := 0; i < runtime.NumCPU(); i++ {
		<-done
	}
}

func TestSTMCommit(t *testing.T) {
	h := NewHandle(&testNode{"a", nil, nil, nil})
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
