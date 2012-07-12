# gotomic

Non blocking data structures for Go.

## Algorithms

The `List` type is implemented using [A Pragmatic Implementation of Non-Blocking Linked-Lists by Timothy L. Harris](http://www.timharris.co.uk/papers/2001-disc.pdf).

The `Hash` type is implemented using [Split-Ordered Lists: Lock-Free Extensible Hash Tables by Ori Shalev and Nir Shavit](http://www.cs.ucf.edu/~dcm/Teaching/COT4810-Spring2011/Literature/SplitOrderedLists.pdf) with the List type used as backend.

The `Transaction` type is implemented using OSTM from [Concurrent Programming Without Locks by Keir Fraser and Tim Harris](http://www.cl.cam.ac.uk/research/srg/netos/papers/2007-cpwl.pdf) with a few tweaks described in https://github.com/zond/gotomic/blob/master/stm.go.

## Performance

On my laptop I created three different benchmarks for a) regular Go `map` types, b) [Go `map` types protected by `sync.RWMutex`](https://github.com/zond/tools/blob/master/tools.go#L142) and c) the `gotomic.Hash` type.

The benchmarks for a) and b) can be found at https://github.com/zond/tools/blob/master/tools_test.go#L83 and the benchmark for c) at https://github.com/zond/gotomic/blob/master/hash_test.go#L116.

The TL;DR of it all is that the benchmark sets `runtime.GOMAXPROCS` to be `runtime.NumCPU()`, and starts that number of `goroutine`s that just mutates and reads the tested mapping.

Last time I ran these tests I got the following results:

a)

    BenchmarkNativeMap	 5000000	       567 ns/op

b)

    BenchmarkMyMapConc	  200000	     10694 ns/op
    BenchmarkMyMap	 1000000	      1427 ns/op

c)

    BenchmarkHash      500000	      5146 ns/op
    BenchmarkHashConc	  500000	     10599 ns/op

Conclusion: As expected a) is by far the fastest mapping, and it seems that the naive RWMutex wrapped native map b) is much faster at single thread operation, and about as efficient in multi thread operation, compared to c).

I find it likely that c) is more efficient than b) at higher levels of concurrency, but for regular uses an RWMutex wrapped `map` is probably a safer bet.

## Usage

See https://github.com/zond/gotomic/blob/master/examples/example.go or https://github.com/zond/gotomic/blob/master/examples/profile.go

Also, see the tests.

## Documentation

http://go.pkgdoc.org/github.com/zond/gotomic

## Bugs

No known bugs in `Hash` and `List`. The `Transaction` and `Handle` types are alpha and still fail in the tests.

I have not tried it on more than my personal laptop however, so if you want to try and force it to misbehave on a heftier machine than a 4 cpu MacBook Air please do!

## Improvements

It would be nice to have a Hash#DeleteIfPresent that atomically deletes matching key/value pairs, but since the implementation is slightly harder than trivial and I see no immediate use case I have been too lazy. Tell me if you need it and I might feel motivated :)
