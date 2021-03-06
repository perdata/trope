// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package trope_test

import (
	"github.com/perdata/trope"
	"testing"
)

func TestNode(t *testing.T) {
	zero := trope.New(nil, 0)
	if zero.Count != 0 {
		t.Fatal("Zero initialization", zero)
	}

	if zz := zero.Slice(0, 0); toString(zz) != toString(zero) || zz.Count != 0 {
		t.Fatal("Zero slice", zz)
	}

	hello := zero.Splice(0, 0, trope.New(Slicer("hello"), 5))
	if x := toString(hello); x != "hello" {
		t.Fatal("Initial Splicing fails", x)
	}

	if x := toString(hello.Slice(3, 0)); x != "" {
		t.Fatal("zero slice", x)
	}

	if x := toString(hello.Slice(0, 4)); x != "hell" {
		t.Fatal("Initial slice  fails", x)
	}

	jello := hello.Splice(0, 1, trope.New(Slicer("j"), 1))
	if x := toString(jello); x != "jello" {
		t.Fatal("jello", x)
	}

	jimbo := jello.Splice(1, 3, trope.New(Slicer("imb"), 3))
	if x := toString(jimbo); x != "jimbo" {
		t.Fatal("jimbo", x)
	}

	if x := toString(jimbo.Slice(2, 2)); x != "mb" {
		t.Fatal("mb", x)
	}

	jino := jimbo.Splice(2, 2, trope.New(Slicer("n"), 1))
	if x := toString(jino); x != "jino" {
		t.Fatal("jino", x)
	}

	jinova := jino.Splice(4, 0, trope.New(Slicer("va"), 2))
	if x := toString(jinova); x != "jinova" {
		t.Fatal("jinova", x)
	}

	djino := trope.New(Slicer("d"), 1).Splice(1, 0, jino)
	if x := toString(djino); x != "djino" {
		t.Fatal("djino", x)
	}

	djinojinova := djino.Splice(5, 0, jinova)
	if x := toString(djinojinova); x != "djinojinova" {
		t.Fatal("djinojinova", x)
	}
}

func TestInvalidOffsets(t *testing.T) {
	mustPanic := func(fn func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("Failed to panic")
			}
		}()
		fn()
	}

	replace := trope.New(Slicer("replace"), 7)
	initial := trope.New(Slicer("hello"), 5)
	mustPanic(func() {
		initial.Slice(-1, 4)
	})
	mustPanic(func() {
		initial.Slice(1, -2)
	})
	mustPanic(func() {
		initial.Slice(3, 20)
	})
	mustPanic(func() {
		initial.Splice(-1, 4, replace)
	})
	mustPanic(func() {
		initial.Splice(1, -2, replace)
	})
	mustPanic(func() {
		initial.Splice(3, 20, replace)
	})
}

func TestRandomSplices(t *testing.T) {
	benchmarkTest(1, &validatedInitSplicer{}, 500000, 5)
}

type validatedInitSplicer struct {
	tropeInitSplicer
	stringInitSplicer
}

func (is *validatedInitSplicer) Init(str string) {
	is.tropeInitSplicer.Init(str)
	is.stringInitSplicer.Init(str)
}

func (is *validatedInitSplicer) Splice(offset, count int, r string) {
	is.tropeInitSplicer.Splice(offset, count, r)
	is.stringInitSplicer.Splice(offset, count, r)
	if toString(is.tropeInitSplicer.Node) != is.stringInitSplicer.s {
		panic("strings diverged")
	}
}

func TestFlatten(t *testing.T) {
	zero := trope.New(Slicer(""), 0)
	if x := toString(zero.Flatten(100)); x != "" {
		t.Fatal("could not flatten zero", x)
	}

	hello := trope.New(Slicer("hello"), 5)
	flat := hello.Flatten(100)
	if x := toString(flat); x != "hello" {
		t.Fatal("Single chunk flattening failed", x)
	}

	hello4 := hello.Splice(0, 0, hello).Splice(0, 0, hello).Splice(0, 0, hello)
	flat = hello4.Flatten(3)
	if x := toString(flat); x != "hellohellohellohello" {
		t.Fatal("Single chunk flattening failed", x)
	}

	flat = hello4.Flatten(2)
	if x := toString(flat); x != "hellohellohellohello" {
		t.Fatal("Single chunk flattening failed", x)
	}
}

func toString(n trope.Node) string {
	result := ""
	n.ForEach(func(leaf interface{}, count int) {
		result += string(leaf.(Slicer))
	})
	return result
}

type Slicer string

func (s Slicer) Slice(offset, count int) interface{} {
	return s[offset : offset+count]
}

func (s Slicer) Splice(offset, count int, replacement interface{}) interface{} {
	r := replacement.(Slicer)
	return s[:offset] + r + s[offset+count:]
}
