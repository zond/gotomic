package gotomic

import (
	"testing"
	"fmt"
)

type key string
func (self key) HashCode() uint32 {
	var sum uint32
	for _, c := range string(self) {
		sum += uint32(c)
	}
	return sum
}
func (self key) Equals(t thing) bool {
	if s, ok := t.(key); ok {
		return s == self
	}
	return false
}

func TestPutGet(t *testing.T) {
	h := newHash()
	fmt.Println(h.describe())
	h.put(key("a"), "b")
	fmt.Println(h.describe())
	h.put(key("a"), "b")
	fmt.Println(h.describe())
	h.put(key("c"), "d")
	fmt.Println(h.describe())
}

