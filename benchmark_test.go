// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package trope_test

// Sadly, github.com/vinzmay/go-rope crashes for initial string sizes
// of less than 5000 with the benchmarks.
import (
	"flag"
	"github.com/chewxy/skiprope"
	"github.com/perdata/trope"
	"github.com/vinzmay/go-rope"
	"math/rand"
	"testing"
)

func BenchmarkTrope(b *testing.B) {
	var v trope.Node
	init := func(str string) {
		v = trope.New(Slicer(str), len(str))
	}
	splice := func(offset, count int, r string) {
		v = v.Splice(offset, count, trope.New(Slicer(r), len(r)))
	}
	benchmark(b, init, splice)
}

func BenchmarkString(b *testing.B) {
	s := ""
	init := func(str string) {
		s = str
	}
	splice := func(offset, count int, r string) {
		s = s[:offset] + r + s[offset+count:]
	}
	benchmark(b, init, splice)
}

func BenchmarkSkiprope(b *testing.B) {
	skip := skiprope.New()
	init := func(str string) {
		if err := skip.Insert(0, str); err != nil {
			b.Fatal("Failed to insert", err)
		}
	}
	splice := func(offset, count int, r string) {
		if err := skip.EraseAt(offset, count); err != nil {
			b.Fatal("EraseAt", err)
		}
		if err := skip.Insert(offset, r); err != nil {
			b.Fatal("Insert", err)
		}
	}
	benchmark(b, init, splice)
}

func BenchmarkRope(b *testing.B) {
	if strlen < 5000 {
		b.Fatal("The rope benchmark only works for >= 5000 strlen")
	}

	rr := rope.New("")
	init := func(str string) {
		rr = rope.New(str)
	}
	splice := func(offset, count int, r string) {
		rr = rr.Delete(offset, count)
		rr = rr.Insert(offset, r)
	}
	benchmark(b, init, splice)
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var largeRandomString string
var strlen int

func init() {
	strlen = 1000000
	flag.IntVar(&strlen, "strlen", 1000000, "length of large string to  use")
	flag.Parse()

	initRandomString(strlen)
}

func initRandomString(size int) {
	rand.Seed(42)
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	largeRandomString = string(b)
}

func benchmark(b *testing.B, init func(string), splice func(offset, count int, v string)) {
	for n := 0; n < b.N; n++ {
		benchmarkRun(init, splice)
	}
}

func benchmarkRun(init func(string), splice func(offset, count int, v string)) {
	rand.Seed(42)
	iter := 100
	str := largeRandomString
	init(str)
	size := len(str)
	for kk := 0; kk < iter; kk++ {
		randAlpha := 'a' + rune(rand.Intn(26))
		v := string([]rune{randAlpha})
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

		splice(offset, count, v)
		size += 1 - count
	}
}
