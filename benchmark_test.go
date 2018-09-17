// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package trope_test

// Sadly, github.com/vinzmay/go-rope crashes for initial string sizes
// of less than 5000 with the benchmarks.
import (
	"fmt"
	"github.com/chewxy/skiprope"
	"github.com/perdata/lazy"
	"github.com/perdata/trope"
	"github.com/vinzmay/go-rope"
	"math/rand"
	"strings"
	"testing"
)

type InitSplicer interface {
	Init(str string)
	Splice(offset, count int, r string)
}

func Benchmark(b *testing.B) {
	tests := []struct {
		Name string
		InitSplicer
		MinSize int
	}{
		{"Trope", &tropeInitSplicer{}, 0},
		{"Hybrid", &hybridInitSplicer{}, 0},
		{"Lazy", &lazyInitSplicer{}, 0},
		{"String", &stringInitSplicer{}, 0},
		{"Skiprope", &skipRopeInitSplicer{}, 0},
		{"Rope", &ropeInitSplicer{}, 10000},
	}

	inputSizes := []int{1000, 5000, 10000, 15000, 200000, 1000000}
	spliceSizes := []int{2, 10, 100}

	for _, inputSize := range inputSizes {
		for _, spliceSize := range spliceSizes {
			for _, test := range tests {
				if test.MinSize > inputSize {
					continue
				}

				name := fmt.Sprintf("%v:%v:%v", test.Name, inputSize, spliceSize)
				b.Run(name, func(b *testing.B) {
					benchmarkTest(b.N, test.InitSplicer, inputSize, spliceSize)
				})

			}
		}
	}
}

func benchmarkTest(n int, initSplicer InitSplicer, inputSize int, spliceSize int) {
	iter := 100
	for kk := 0; kk < n; kk++ {
		rand.Seed(42)
		str := randomString(inputSize)
		initSplicer.Init(str)
		size := len(str)
		for jj := 0; jj < iter; jj++ {
			splice := randomString(spliceSize)
			offset, count := 0, 0

			if size > 0 {
				offset = rand.Intn(size)
			}
			diff := size - offset
			if diff > 100 {
				diff = 100
			}

			if diff > 0 {
				count = rand.Intn(diff)
			}
			initSplicer.Splice(offset, count, splice)
			size += spliceSize - count
		}
	}
}

type tropeInitSplicer struct {
	trope.Node
}

func (is *tropeInitSplicer) Init(str string) {
	is.Node = trope.New(Slicer(str), len(str))
}

func (is *tropeInitSplicer) Splice(offset, count int, r string) {
	is.Node = is.Node.Splice(offset, count, trope.New(Slicer(r), len(r)))
}

type hybridInitSplicer struct {
	trope.Hybrid
}

func (is *hybridInitSplicer) Init(str string) {
	is.Hybrid = trope.Hybrid{15000, 10000, Slicer(""), 0, trope.New(Slicer(str), len(str))}
}

func (is *hybridInitSplicer) Splice(offset, count int, r string) {
	replace := trope.Hybrid{15000, 10000, Slicer(r), len(r), trope.Node{}}
	is.Hybrid = is.Hybrid.Splice(offset, count, replace)
}

type stringInitSplicer struct {
	s string
}

func (is *stringInitSplicer) Init(str string) {
	is.s = str
}

func (is *stringInitSplicer) Splice(offset, count int, r string) {
	is.s = is.s[:offset] + r + is.s[offset+count:]
}

type skipRopeInitSplicer struct {
	rope *skiprope.Rope
}

func (is *skipRopeInitSplicer) Init(str string) {
	is.rope = skiprope.New()
	if err := is.rope.Insert(0, str); err != nil {
		panic(err)
	}
}

func (is *skipRopeInitSplicer) Splice(offset, count int, r string) {
	if err := is.rope.EraseAt(0, offset); err != nil {
		panic(err)
	}
	if err := is.rope.Insert(offset, r); err != nil {
		panic(err)
	}
}

type ropeInitSplicer struct {
	rope *rope.Rope
}

func (is *ropeInitSplicer) Init(str string) {
	is.rope = rope.New(str)
}

func (is *ropeInitSplicer) Splice(offset, count int, r string) {
	is.rope = is.rope.Delete(offset, count)
	is.rope = is.rope.Insert(offset, r)
}

type lazyInitSplicer struct {
	lazy.Array
}

func (is *lazyInitSplicer) Init(str string) {
	is.Array = lazyArray(str)
}

func (is *lazyInitSplicer) Splice(offset, count int, r string) {
	is.Array = is.Array.Splice(offset, count, lazyArray(r))
	if is.Limit <= 0 {
		var b strings.Builder
		b.Grow(is.Array.Count)
		is.Array.ForEach(func(v interface{}, _ int) {
			b.Write([]byte(string(v.(Slicer))))
		})
		is.Array = lazyArray(b.String())
	}
}

func randomString(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var letters = "abcdefghijklmnopqrstuvwxyz"

type lazySlicer lazy.Array

func (l lazySlicer) Slice(offset, count int) interface{} {
	return flatten(lazy.Array(l).Slice(offset, count))
}

func (l lazySlicer) Splice(offset, count int, replacement interface{}) interface{} {
	rep := lazy.Array(replacement.(lazySlicer))
	return flatten(lazy.Array(l).Splice(offset, count, rep))
}

func flatten(l lazy.Array) lazySlicer {
	if l.Limit > 0 {
		return lazySlicer(l)
	}

	var b strings.Builder
	b.Grow(l.Count)
	l.ForEach(func(v interface{}, _ int) {
		b.Write([]byte(string(v.(Slicer))))
	})
	return lazySlicer(lazyArray(b.String()))
}

func lazyArray(s string) lazy.Array {
	return lazy.Array{Limit: 15, Count: len(s), Value: Slicer(s)}
}
