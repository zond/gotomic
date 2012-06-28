
package main

import (
	gotomic "../"
	"runtime/pprof"
	"fmt"
	"runtime"
	"math/rand"
	"time"
	"os"
)

type hashInt int
func (self hashInt) HashCode() uint32 {
	return uint32(self)
}
func (self hashInt) Equals(t gotomic.Thing) bool {
	if i, ok := t.(hashInt); ok {
		return i == self
	} 
	return false
}

func work(h *gotomic.Hash, n int, do, done chan bool) {
	<- do
	for i := 0; i < n; i++ {
		k := hashInt(i)
		h.Put(k, i)
	}
	for i := 0; i < n; i++ {
		if hv, _ := h.Get(hashInt(i)); hv != i {
			fmt.Println("bad value in hash, expected ", i, " but got ", hv)
		}
	}
	done <- true
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())
	f, err := os.Create("cpuprofile")
	if err != nil {
		panic(err.Error())
	}		
	f2, err := os.Create("memprofile")
	if err != nil {
		panic(err.Error())
	}		
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()	
	defer pprof.WriteHeapProfile(f2)

	h := gotomic.NewHash()
	do := make(chan bool)
	done := make(chan bool)
	n := 1000000
	go work(h, n, do, done)
	go work(h, n, do, done)
	go work(h, n, do, done)
	go work(h, n, do, done)
	close(do)
	<- done
	<- done
	<- done
	<- done
}