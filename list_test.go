package gotomic

import (
	"fmt"
	"testing"
)

type s string
func (self s) Compare(o interface{}) int {
	return 0
}

func fiddle(nr *nodeRef, do, done chan bool) {
	<- do
	for i := 0; i < 1000; i++ {
		nr.push(s(fmt.Sprint("x", i)))
		nr.pop()
	}
	done <- true
}

func TestTest(t *testing.T) {
	nr := new(nodeRef)
	fmt.Println(nr)
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
	do := make(chan bool)
	done := make(chan bool)
	go fiddle(nr, do, done)
	go fiddle(nr, do, done)
	go fiddle(nr, do, done)
	close(do)
	<-done
	<-done
	<-done
	fmt.Println(nr)
}

