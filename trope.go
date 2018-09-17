// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// Package trope implements a simple rope-like datastructure for large
// immutable collections.
//
// Unlike a rope which uses binary trees, trope uses a dynamic branch
// factor.
//
// The data structure is mainly optimized for performance when a large
// number of edit operations are made.  The underlying tree is not
// automatically rebalanced as most situations can do without this
// step.  A Flatten() method is provided to construct a somewhat
// balanced tree in an efficient manner but this is not a true
// balancing. In particular, the number of nodes depends on the number
// of edits and the root node has unbounded branching factor. But in
// most practical sitations, this will work fine.
//
// The rope datastructure is often too expensive for fairly small
// arrays. The Hybrid type is defined to get the best of both worlds
// by using the regular array implementation for small counts and
// switching to the more complex structure at a configured high water
// mark.
//
// Benchmarks
//
// These benchmarks include 100 iterations of random slicing (on top
// of the previous result).  The comparison is between trope.Node,
// trope.Hybrid and a simple string-slice based splice operation.
//
// String size of 5k:
//
//    BenchmarkTrope-4      	   10000	    220341 ns/op	  211384 B/op	    1871 allocs/op
//    BenchmarkHybrid-4     	   20000	     88510 ns/op	  292664 B/op	     413 allocs/op
//    BenchmarkString-4     	   20000	     71310 ns/op	  283440 B/op	     200 allocs/op
//
//
// String size of 200k:
//
//    BenchmarkTrope-4      	   10000	    186928 ns/op	  188264 B/op	    1583 allocs/op
//    BenchmarkHybrid-4     	   10000	    193612 ns/op	  188264 B/op	    1583 allocs/op
//    BenchmarkString-4     	    1000	   2243631 ns/op	20226499 B/op	     200 allocs/op
//
//
// The benchmarks are obviously specific to the hardware but it gives
// an idea about the relative performance characterestics.
//
package trope

// Slicer is an optional interface to be implemented by the leaf-node
// values.  If the leaf-node value are all single-item arrays, this is
// not needed at all.
type Slicer interface {
	Slice(offset, count int) interface{}
}

// Node is the immutable node in the tree representing the
// collection. Use New() to create a new node.  The ID is guaranteed
// to be unique for a  specific tree -- edits will cause those nodes
// that changed (typically along the path of the edit) to get a new
// ID.
//
// If Children is nil, the node simply holds the underlying leaf
// element(s). Count is still valid and specifies the number of
// elements.
type Node struct {
	getID    func() int
	ID       int
	Children []Node
	Leaf     interface{}
	Count    int
}

// New creates a new node populated  with the initial elements of
// specified count. The provided initial elements are stored as the
// Leaf value.
func New(initial interface{}, count int) Node {
	id := 0
	getID := func() int {
		id++
		return id
	}
	return Node{ID: id, getID: getID, Leaf: initial, Count: count}
}

// ForEach recursively traverses the node and its children calling the
// provided function on all the Leaf values
func (n Node) ForEach(fn func(v interface{}, count int)) {
	n.forEach(func(leaf Node) {
		fn(leaf.Leaf, leaf.Count)
	})
}

func (n Node) forEach(fn func(n Node)) {
	if n.Count == 0 {
		return
	}

	if n.Children == nil {
		fn(n)
		return
	}

	for _, child := range n.Children {
		child.forEach(fn)
	}
}

// Flatten constructs a 2-level list. The leaf nodes are all
// aggregated into the first level in groups of the specified chunk
// size and these are all then aggregated into the root node.
//
// Note that the root node won't honor the chunk size.
func (n Node) Flatten(chunkSize int) Node {
	children := []Node(nil)
	leafs := []Node(nil)
	count := 0
	n.forEach(func(leaf Node) {
		leafs = append(leafs, leaf)
		count += leaf.Count
		if len(leafs) == chunkSize {
			children = append(children, Node{
				ID:       n.getID(),
				getID:    n.getID,
				Children: leafs,
				Count:    count,
			})
			count = 0
			leafs = nil
		}
	})
	if leafs != nil {
		children = append(children, Node{
			ID:       n.getID(),
			getID:    n.getID,
			Children: leafs,
			Count:    count,
		})
	}
	n.ID = n.getID()
	n.Children = children
	return n
}

// Slice returns a Node which references only the elements between
// offset and offset+count. If this involves slicing leaf nodes, it
// will look for the leaf node elements to implement the Slicer
// interface.
func (n Node) Slice(offset, count int) Node {
	if offset < 0 || count < 0 || offset+count > n.Count {
		panic("Unexpected offset, count")
	}

	if offset == 0 && count == n.Count {
		return n
	}

	if count == 0 {
		return Node{ID: n.getID(), getID: n.getID}
	}

	if n.Children == nil {
		return n.sliceLeaf(offset, count)
	}

	seen := 0
	children := []Node{}
	for kk := 0; kk < len(n.Children) && seen < offset+count; kk++ {
		child, start, end := n.Children[kk], seen, seen+n.Children[kk].Count
		if offset > start {
			start = offset
		}
		if offset+count < end {
			end = offset + count
		}
		if start < end {
			children = append(children, child.Slice(start-seen, end-start))
		}
		seen = end
	}

	n.ID = n.getID()
	n.Children = children
	n.Count = count
	return n
}

// Splice removes the elements at the provided offset and replaces
// them with the provided replacement.
func (n Node) Splice(offset, count int, replacement Node) Node {
	if offset == 0 && count == n.Count {
		replacement.ID = n.getID()
		return replacement
	}

	if offset == n.Count && count == 0 {
		return n.join(replacement)
	}

	// if it affects a sub-node only, then optimize for it
	seen := 0
	for kk := 0; kk < len(n.Children) && seen <= offset; kk++ {
		child := n.Children[kk]
		if seen+child.Count >= offset+count {
			child = child.Splice(offset-seen, count, replacement)
			n.Children = append([]Node(nil), n.Children...)
			n.Children[kk] = child
			n.ID = n.getID()
			n.Count += replacement.Count - count
			return n
		}
		seen += child.Count
	}

	// slow path
	children := n.Children
	if len(children) > 0 {
		first := children[0]
		last := children[len(children)-1]
		if offset >= first.Count || offset+count <= n.Count-last.Count {
			return n.spliceChildren(offset, count, replacement)
		}
	}

	// slower path
	right := n.Slice(offset+count, n.Count-offset-count)
	return n.Slice(0, offset).join(replacement).join(right)
}

func (n Node) spliceChildren(offset, count int, replacement Node) Node {
	left, right, mid := 0, 0, 0
	leftCount, rightCount, midCount := 0, 0, 0
	seen := 0
	for _, ch := range n.Children {
		switch {
		case seen+ch.Count <= offset:
			left++
			leftCount += ch.Count
		case seen >= offset+count:
			right++
			rightCount += ch.Count
		default:
			mid++
			midCount += ch.Count
		}
		seen += ch.Count
	}
	innerLeft := n.Children[left].Slice(0, offset-leftCount)
	r := n.Children[left+mid-1]
	offsetr := offset + count - (n.Count - rightCount - r.Count)
	countr := r.Count - offsetr
	innerRight := r.Slice(offsetr, countr)
	inner := innerLeft.join(replacement).join(innerRight)
	result := n
	result.ID = n.getID()
	result.Count = n.Count - count + replacement.Count
	result.Children = n.Children[:left:left]
	if len(inner.Children) > 0 {
		result.Children = append(result.Children, inner.Children...)
	} else {
		result.Children = append(result.Children, inner)
	}
	result.Children = append(result.Children, n.Children[left+mid:]...)
	return result
}

// Threshold at which node height is increased in favor of creating
// larger chlidren array.  This threshold is very likely dependent on
// hardware and such but the number is high enough for this to be rare
const limit = 100

func (n Node) join(o Node) Node {
	result := n
	switch {
	case n.Count == 0:
		result = o
	case o.Count == 0:
	case n.Children == nil && o.Children == nil:
		result.Children = []Node{n, o}
	case n.Children == nil, o.Children != nil && len(n.Children) > limit:
		result.Children = append([]Node{n}, o.Children...)
	case o.Children == nil:
		result.Children = append(append([]Node(nil), n.Children...), o)
	default:
		result.Children = append(append([]Node(nil), n.Children...), o.Children...)
	}
	result.Count = n.Count + o.Count
	result.ID = n.getID()
	return result
}

func (n Node) sliceLeaf(offset, count int) Node {
	n.ID = n.getID()
	n.Leaf = (n.Leaf).(Slicer).Slice(offset, count)
	n.Count = count
	return n
}
