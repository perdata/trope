# trope

[![Status](https://travis-ci.com/perdata/trope.svg?branch=master)](https://travis-ci.com/perdata/trope?branch=master)
[![GoDoc](https://godoc.org/github.com/perdata/trope?status.svg)](https://godoc.org/github.com/perdata/trope)
[![codecov](https://codecov.io/gh/perdata/trope/branch/master/graph/badge.svg)](https://codecov.io/gh/perdata/trope)
[![GoReportCard](https://goreportcard.com/badge/github.com/perdata/trope)](https://goreportcard.com/report/github.com/perdata/trope)

Trope is a golang package for dealing with very large immutable arrays
at decent performance with respect to random access edits.

It uses a rope-like tree structure (which is neither balanced by
default, nor is it binary as with regular ropes).

## Benchmark stats

Trope is mainly focused on good performance under random edits (not
just appends).

The benchmark is run via something like: 

```sh
$> go test --bench=. -strlen=1000000
```

The following table shows the time (in microseconds) taken for each of
the implementations for 100 random splice operations. Trope performs
at a relatively fixed cost without much dependency on the size of the
array but this fixed cost only outweight the raw slice performance at
around 10k size.


| String Size | 5k | 10k | 100k | 1M |
| ----------- | --- | --- | --- | --- |
| Trope | 219 | 188 | 181 | 188 |
| String | 70 | 134 | 1040 | 10195 | 
| Skiprope | 123 | 193 | 1240 | 12159 |
| Rope | 233 | 256 | 452 | 2501 |


