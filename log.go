package gotomic

/*
 * Ripped from http://graphics.stanford.edu/~seander/bithacks.html#IntegerLogLookup
 */
var LOG_TABLE_256 []uint32

func init() {
	LOG_TABLE_256 = make([]uint32, 256)
	for i := 2; i < 256; i++ {
		LOG_TABLE_256[i] = 1 + LOG_TABLE_256[i/2]
	}
}
func log2(v uint32) uint32 {
	var r, tt uint32
	if tt = v >> 24; tt != 0 {
		r = 24 + LOG_TABLE_256[tt]
	} else if tt = v >> 16; tt != 0 {
		r = 16 + LOG_TABLE_256[tt]
	} else if tt = v >> 8; tt != 0 {
		r = 8 + LOG_TABLE_256[tt]
	} else {
		r = LOG_TABLE_256[v]
	}
	return r
}
