package gotomic

import (
	"fmt"
	"testing"
	"runtime"
)

func fiddle(n string, nr *nodeRef, do, done chan bool) {
	<- do
	for i := 0; i < 10000; i++ {
		nr.push(fmt.Sprint(n, i))
		nr.pop()
	}
	done <- true
}

func TestTest(t *testing.T) {
	runtime.GOMAXPROCS(4)
	nr := new(nodeRef)
	fmt.Println(nr)
	nr.push("hej")
	nr.push("haj")
	nr.push("hoj")
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
	do := make(chan bool)
	done := make(chan bool)
	go fiddle("a", nr, do, done)
	go fiddle("b", nr, do, done)
	go fiddle("b", nr, do, done)
	go fiddle("b", nr, do, done)
	close(do)
	<-done
	<-done
	<-done
	<-done
	fmt.Println(nr)
}

