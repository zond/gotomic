package gotomic

import (
	"fmt"
	"testing"
)

type s string
func (self s) Compare(o interface{}) int {
	return 0
}

func TestTest(t *testing.T) {
	nr := new(nodeRef)
	nr.push(s("hej"))
	nr.push(s("haj"))
	nr.push(s("hoj"))
	fmt.Println(nr)
}