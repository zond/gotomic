package gotomic

import (
	"testing"
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
	h.put(key("a"), "b")
	h.put(key("a"), "b")
	h.put(key("c"), "d")
}

