
package gotomic

import (
	"testing"
	"fmt"
)

func TestFoo(t *testing.T) {
	treap := NewTreap()
	for i := 0; i < 10; i++ {
		treap.Put(c(i), 100)
	}
	fmt.Println(treap.Describe())

}