
package gotomic

import (
	"testing"
	"fmt"
)

func TestFoo(t *testing.T) {
	treap := NewTreap()
	for i := 0; i < 100; i++ {
		treap.Put(c(i), 100)
	}
	fmt.Println(treap.Describe())
}

func TestBar(t *testing.T) {
	treap := NewTreap()
	for i := 0; i < 10; i++ {
		treap.Put(c(i), 100)
	}
	fmt.Println(treap.Describe())
	treap.Delete(c(4))
	fmt.Println(treap.Describe())
}