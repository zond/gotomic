package gotomic

import (
	"testing"
	"fmt"
)

func TestLog2(t *testing.T) {
	if log2(0) != 0 {
		t.Error("log2(0) should be 0 but was ", log2(0))
	}
	h := newHash()
	for i := 0; i < 100; i++ {
		sup, sub := h.getBucketIndices(uint32(i))
		fmt.Println(i, log2(uint32(i)), " => ", sup, sub)
	}
}
