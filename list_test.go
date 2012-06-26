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
	nr.push("1")
	nr.push("2")
	nr.push("3")
	nr.push("4")
	nr.push("5")
	fmt.Println(nr)
}