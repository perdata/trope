# trope

[![Status](https://travis-ci.com/perdata/trope.svg?branch=master)](https://travis-ci.com/perdata/trope?branch=master)
[![GoDoc](https://godoc.org/github.com/perdata/trope?status.svg)](https://godoc.org/github.com/perdata/trope)
[![codecov](https://codecov.io/gh/perdata/trope/branch/master/graph/badge.svg)](https://codecov.io/gh/perdata/trope)
[![GoReportCard](https://goreportcard.com/badge/github.com/perdata/trope)](https://goreportcard.com/report/github.com/perdata/trope)

Trope is a golang package for dealing with very large immutable arrays
at decent performance with respect to random access edits.

It uses a rope-like tree structure (which is neither balanced by
default, nor is it binary as with regular ropes).

Most rope-like structures perform poorly at small sizes.  So, the
package provides a Hybrid implementation which uses regular slices for
small sizes and switches to the rope-like structure at configured
high threshold to get the  best  of both worlds.

## Benchmark stats

Trope is mainly focused on good performance under random edits (not
just appends).

The benchmark is run via something like: 

```sh
$> go test --bench=. -strlen=1000000
```

The following table shows the time (in microseconds) taken for each of
the implementations for 100 random splice operations (on my machine).
Trope performs
at a relatively fixed cost without much dependency on the size of the
array but this fixed cost only outweighs the raw slice performance at
around 15k size.


| String Size | 5k | 10k | 100k | 1M |
| ----------- | --- | --- | --- | --- |
| Trope | 219 | 188 | 181 | 188 |
| String | 70 | 134 | 1040 | 10195 | 
| Skiprope | 123 | 193 | 1240 | 12159 |
| Rope | 233 | 256 | 452 | 2501 |


The memory usage is also a bit of a concern. Here too, the trope
implementation has a much more *steady* usage that does not depend
that much on the size of the string.

```sh
$ go test --bench=. --strlen=100000 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/perdata/trope
BenchmarkTrope-4      	   10000	    180376 ns/op	  172648 B/op	    1519 allocs/op
BenchmarkString-4     	    2000	   1079424 ns/op	10084772 B/op	     200 allocs/op
BenchmarkSkiprope-4   	    1000	   1235641 ns/op	  339567 B/op	    3231 allocs/op
BenchmarkRope-4       	    3000	    451715 ns/op	  671616 B/op	    4398 allocs/op
```

For string sizes larger than 100k, trope does about the same while
other implementations get worse (either in B/op or in allocs/op):

```sh
$ go test --bench=. --strlen=1000000 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/perdata/trope
BenchmarkTrope-4      	   10000	    186289 ns/op	  188648 B/op	    1577 allocs/op
BenchmarkString-4     	     100	  10645611 ns/op	100033109 B/op	     200 allocs/op
BenchmarkSkiprope-4   	     100	  11908668 ns/op	 3331194 B/op	   31355 allocs/op
BenchmarkRope-4       	     500	   2436868 ns/op	 4276162 B/op	    4399 allocs/op
```

For relatively smaller string sizes, trope stays at roughly the same
usage and only Skiprope is better.

```sh
$ go test --bench=. --strlen=10000 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/perdata/trope
BenchmarkTrope-4      	   10000	    193261 ns/op	  182008 B/op	    1602 allocs/op
BenchmarkString-4     	   10000	    136390 ns/op	  796433 B/op	     200 allocs/op
BenchmarkSkiprope-4   	   10000	    188041 ns/op	   33977 B/op	     415 allocs/op
BenchmarkRope-4       	    5000	    267254 ns/op	  323072 B/op	    4584 allocs/op
```

All these combinations point to a custom "hybrid" method that uses
flat strings for small strings and the actual tree structure only for
large strings. Hence the Hybrid type which implements this idea

A Hybrid implementation which switches from string to trope.Node at
15k size and switches back at 10k size has following characterestics
at 5k string size:

```sh
BenchmarkTrope-4      	   10000	    220341 ns/op	  211384 B/op	    1871 allocs/op
BenchmarkHybrid-4     	   20000	     88510 ns/op	  292664 B/op	     413 allocs/op
BenchmarkString-4     	   20000	     71310 ns/op	  283440 B/op	     200 allocs/op
```

This basically tracks the string slice costs at low sizes. At 200k
string sizes, the Hybrid performs similar to Trope basically getting
the best of both worlds

```sh
BenchmarkTrope-4      	   10000	    186928 ns/op	  188264 B/op	    1583 allocs/op
BenchmarkHybrid-4     	   10000	    193612 ns/op	  188264 B/op	    1583 allocs/op
BenchmarkString-4     	    1000	   2243631 ns/op	20226499 B/op	     200 allocs/op
```

