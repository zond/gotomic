package main

import (
	gotomic "../"
)

type key string
func (self key) HashCode() uint32 {
	var rval uint32
	for c := range self {
		rval = rval + uint32(c)
	}
	return rval
}
func (self key) Equals(t gotomic.Thing) bool {
	return t.(key) == self
}

func main() {
	h := gotomic.NewHash()
	h.Put(key("key"), "value")
	if val, _ := h.Get(key("key")); val != "value" {
		panic("wth?")
	}
}
