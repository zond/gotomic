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
	fmt.Println("popped", nr.pop())
	fmt.Println(nr)
	fmt.Println("popped", nr.pop())
	fmt.Println(nr)
	fmt.Println("popped", nr.pop())
	fmt.Println(nr)
	nr.push(s("1"))
	nr.push(s("2"))
	nr.push(s("3"))
	nr.push(s("4"))
	nr.push(s("5"))
	fmt.Println(nr)
}