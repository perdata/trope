// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package trope_test

import (
	"github.com/perdata/trope"
	"testing"
)

func hybridNode(str string) trope.Hybrid {
	return trope.Hybrid{100, 10, Slicer(""), 0, trope.New(Slicer(str), len(str))}
}

func hybridRaw(str string) trope.Hybrid {
	return trope.Hybrid{100, 10, Slicer(str), len(str), trope.New(nil, 0)}
}

func TestHybridSwitcheroo(t *testing.T) {
	zero := trope.Hybrid{10, 5, Slicer(""), 0, trope.New(nil, 0)}
	hw := zero.Splice(0, 0, hybridRaw("hello world"))
	if hw.Count != 0 && hw.Node.Count != 11 {
		t.Fatal("Failed to switch to larger size yo")
	}

	hw = zero.Splice(0, 0, hybridNode("hello world"))
	if hw.Count != 0 && hw.Node.Count != 11 {
		t.Fatal("Failed to switch to larger size yo")
	}

	spliced := hw.Splice(0, 7, hybridRaw(""))
	if spliced.Count != 4 {
		t.Fatal("Failed to switch back to smaller size")
	}

	spliced = hw.Splice(0, 0, hybridRaw("ok "))
	if spliced.Count != 0 || toStringH(spliced) != "ok hello world" {
		t.Fatal("Failed to splice and keep nodes", toStringH(spliced))
	}

	sliced := spliced.Slice(3, 11)
	if x := toStringH(sliced); x != "hello world" {
		t.Fatal("Slice on Node failed", x)
	}
}

func TestHybridSmall(t *testing.T) {
	zero := hybridRaw("")
	if zero.Size() != 0 {
		t.Fatal("Zero initialization", zero)
	}

	if zz := zero.Slice(0, 0); toStringH(zz) != toStringH(zero) || zz.Count != 0 {
		t.Fatal("Zero slice", zz)
	}

	hello := zero.Splice(0, 0, hybridRaw("hello"))
	if x := toStringH(hello); x != "hello" {
		t.Fatal("Initial Splicing fails", x)
	}

	if x := toStringH(hello.Slice(3, 0)); x != "" {
		t.Fatal("zero slice", x)
	}

	if x := toStringH(hello.Slice(0, 4)); x != "hell" {
		t.Fatal("Initial slice  fails", x)
	}

	jello := hello.Splice(0, 1, hybridRaw("j"))
	if x := toStringH(jello); x != "jello" {
		t.Fatal("jello", x)
	}

	jimbo := jello.Splice(1, 3, hybridRaw("imb"))
	if x := toStringH(jimbo); x != "jimbo" {
		t.Fatal("jimbo", x)
	}

	if x := toStringH(jimbo.Slice(2, 2)); x != "mb" {
		t.Fatal("mb", x)
	}

	jino := jimbo.Splice(2, 2, hybridRaw("n"))
	if x := toStringH(jino); x != "jino" {
		t.Fatal("jino", x)
	}

	jinova := jino.Splice(4, 0, hybridRaw("va"))
	if x := toStringH(jinova); x != "jinova" {
		t.Fatal("jinova", x)
	}
}

func toStringH(n trope.Hybrid) string {
	result := ""
	n.ForEach(func(leaf interface{}, count int) {
		result += string(leaf.(Slicer))
	})
	return result
}
