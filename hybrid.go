// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package trope

// Splicer is the interface that the raw elements should implement for
// use with Hybrid
type Splicer interface {
	Slicer
	Splice(offset, count int, replacement interface{}) interface{}
}

// Hybrid switches between a raw array and the Node structure based on
// size thresholds. The HighMark specifies the size at which a raw
// string is converted to a Node.  The LowMark controls when a Node is
// converted back to normal usage. The actual thresholds depend on the
// underlying structures and are best found via benchmarks.
//
// The Count field is only valid if the storage is in Raw rather than
// in Node.
//
// Initialize a new hybrid structure like so:
//
//   h := Hybrid{
//       HighMark: 15000,
//       LowMark: 10000,
//       Raw: somethingThatImplementsSlicerAndSplicer,
//       Count: sizeOfTheAbove,
//   }
//
type Hybrid struct {
	HighMark, LowMark int
	Raw               Splicer
	Count             int
	Node
}

// Size returns the size irrespective of whether it is stored raw or
// in a tree
func (h Hybrid) Size() int {
	if h.Node.Count == 0 {
		return h.Count
	}
	return h.Node.Count
}

// ForEach works like Node.ForEach, iteratiing through all the leaf
// nodes.
func (h Hybrid) ForEach(fn func(v interface{}, count int)) {
	if h.Node.Count == 0 {
		fn(h.Raw, h.Count)
	} else {
		h.Node.ForEach(fn)
	}
}

// Slice returns a sub hybrid with the specified offset and count
func (h Hybrid) Slice(offset, count int) Hybrid {
	if h.Node.Count > 0 {
		h.Node = h.Node.Slice(offset, count)
		return h
	}
	h.Raw = h.Raw.Slice(offset, count).(Splicer)
	h.Count = count
	return h
}

// Splice removes the specified offset/count and then replaces it with
// the provided replacement.  This will convert from Raw to Node and
// back as specified by the HighMark and LowMark respectively.
func (h Hybrid) Splice(offset, count int, replacement Hybrid) Hybrid {
	if h.Node.Count == 0 && h.Size()+replacement.Size()-count > h.HighMark {
		h = Hybrid{h.HighMark, h.LowMark, h.Raw.Slice(0, 0).(Splicer), 0, New(h.Raw, h.Count)}
	}

	if h.Node.Count > 0 {
		n := replacement.Node
		if n.Count == 0 {
			n = New(replacement.Raw, replacement.Count)
		}
		h.Node = h.Node.Splice(offset, count, n)
		if h.Node.Count < h.LowMark {
			return h.simplify()
		}
		return h
	}

	r := replacement.simplify()
	h.Raw = h.Raw.Splice(offset, count, r.Raw).(Splicer)
	h.Count += r.Count - count
	if h.Count > h.HighMark {
		return Hybrid{h.HighMark, h.LowMark, h.Raw.Slice(0, 0).(Splicer), 0, New(h.Raw, h.Count)}
	}

	return h
}

func (h Hybrid) simplify() Hybrid {
	if h.Node.Count == 0 {
		return h
	}

	raw := h.Raw.Slice(0, 0).(Splicer)
	total := 0
	h.Node.ForEach(func(v interface{}, count int) {
		raw = raw.Splice(total, 0, v).(Splicer)
		total += count
	})
	return Hybrid{h.HighMark, h.LowMark, raw, total, Node{}}
}
