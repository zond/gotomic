
package main

import (
	gotomic "../"
	"runtime/pprof"
	"fmt"
	"runtime"
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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	f, err := os.Create("cpuprofile")
	if err != nil {
		panic(err.Error())
	}		
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()	

	h := gotomic.NewHash()
	cmp := make(map[gotomic.Hashable]interface{})
	for i := 0; i < 1000000; i++ {
		k := hashInt(i)
		h.Put(k, i)
		cmp[k] = i
	}
	for k, v := range cmp {
		if hv := h.Get(k); hv != v {
			fmt.Println("bad value in hash, expected ", v, " but got ", hv)
		}
	}
}