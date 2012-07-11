
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

func TestBlaj(t *testing.T) {
	h := NewHandle(&testNode{"a", nil, nil})
	tr := NewTransaction()
	n := tr.Read(h).(*testNode)
	n.value = "b"
	t.Errorf("%v", n)
	tr2 := NewTransaction()
	n2 := tr2.Read(h)
	t.Errorf("%v", n2)
}