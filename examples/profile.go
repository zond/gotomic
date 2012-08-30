
package main

import (
	gotomic "../"
	"runtime/pprof"
	"runtime"
	"math/rand"
	"time"
	"os"
)

type compInt int

func (self compInt) Compare(t gotomic.Thing) int {
	if i, ok := t.(compInt); ok {
		if i > self {
			return 1
		} else if i < self {
			return -1
		}
	}
	return 0
}

func work(h *gotomic.Treap, n int, do, done chan bool) {
	<- do
	keys := make([]compInt, n)
	for i := 0; i < n; i++ {
		k := compInt(rand.Int())
		keys[i] = k
		h.Put(k, i)
	}
	for i := 0; i < n; i++ {
		h.Get(keys[i])
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

	h := gotomic.NewTreap()
	do := make(chan bool)
	done := make(chan bool)
	n := 100000
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