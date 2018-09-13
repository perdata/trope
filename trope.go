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
// step.
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
	if n.Count == 0 {
		return
	}

	if n.Children == nil {
		fn(n.Leaf, n.Count)
		return
	}

	for _, child := range n.Children {
		child.ForEach(fn)
	}
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
	right := n.Slice(offset+count, n.Count-offset-count)
	return n.Slice(0, offset).join(replacement).join(right)
}

func (n Node) join(o Node) Node {
	result := n
	switch {
	case n.Count == 0:
		result = o
	case o.Count == 0:
	case n.Children == nil && o.Children == nil:
		result.Children = []Node{n, o}
	case n.Children == nil:
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
