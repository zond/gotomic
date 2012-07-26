
package gotomic

import (
	"testing"
)

func TestFoo(t *testing.T) {
	treap := NewTreap()
	treap.Put(c(0), 100)
	treap.Put(c(1), 100)
	treap.Put(c(2), 100)
	t.Error(treap.Describe())
}