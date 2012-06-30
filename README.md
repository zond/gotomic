# gotomic

Non blocking data structures for Go.

## Algorithms

The `List` type is implemented using [A Pragmatic Implementation of Non-Blocking Linked-Lists by Timothy L. Harris](http://www.timharris.co.uk/papers/2001-disc.pdf).

The `Hash` type is implemented using [Split-Ordered Lists: Lock-Free Extensible Hash Tables by Ori Shalev and Nir Shavit](http://www.cs.ucf.edu/~dcm/Teaching/COT4810-Spring2011/Literature/SplitOrderedLists.pdf) with the List type used as backend.

## Performance

On my laptop I created three different benchmarks for a) regular Go `map` types, b) [Go `map` types wrapped by a `channel` and `goroutine`](https://github.com/zond/tools/blob/master/tools.go#L142) and c) the `gotomic.Hash` type.

The benchmarks for a) and b) can be found at https://github.com/zond/tools/blob/master/tools_test.go#L82 and the benchmark for c) at https://github.com/zond/gotomic/blob/master/hash_test.go#L107.

The TL;DR of it all is that the benchmark sets `runtime.GOMAXPROCS` to be `runtime.NumCPU()`, and starts that number of `goroutine`s that just mutates and reads the tested mapping.

Last time I ran these tests I got the following results:

a)

    BenchmarkNativeMap	 5000000	       567 ns/op

b)

    BenchmarkMyMapConc	   50000	     54408 ns/op
    BenchmarkMyMap	 1000000	      2885 ns/op

c)

    BenchmarkHash	 1000000	      4182 ns/op
    BenchmarkHashConc	  500000	      7289 ns/op

Notice that as expected a) is by far the fastest mapping, but if you want a thread safe mapping (and yeah, _Don't communicate by sharing memory; share memory by communicating_, but sometimes it really is easier to have a global mapping) the Non-Blocking mapping is more than seven times faster than the `channel`-wrapped native `map`.

## Usage

See https://github.com/zond/gotomic/blob/master/examples/example.go or https://github.com/zond/gotomic/blob/master/examples/profile.go

## Bugs

No known bugs.

I have not tried it on more than my personal laptop however, so if you want to try and force it to misbehave on a heftier machine than a 4 cpu MacBook Air please do!
