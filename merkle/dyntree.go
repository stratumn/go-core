// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle

import (
	"crypto/sha256"
	"sync"
)

// DynTreeNode is a node within a DynTree.
type DynTreeNode struct {
	hash   Hash
	left   *DynTreeNode
	right  *DynTreeNode
	parent *DynTreeNode
	height int
}

// Hash returns the hash of the node.
func (n *DynTreeNode) Hash() Hash {
	return n.hash
}

// Left returns the node to the left, if any.
func (n *DynTreeNode) Left() *DynTreeNode {
	return n.left
}

// Right returns the node to the right, if any.
func (n *DynTreeNode) Right() *DynTreeNode {
	return n.right
}

// Parent returns the parent node, if any.
func (n *DynTreeNode) Parent() *DynTreeNode {
	return n.parent
}

func (n *DynTreeNode) rehash(a, b Hash) error {
	h := sha256.New()
	if _, err := h.Write(a[:]); err != nil {
		return err
	}
	if _, err := h.Write(b[:]); err != nil {
		return err
	}
	copy(n.hash[:], h.Sum(nil))

	if n.parent != nil {
		if n.left != nil {
			n.parent.rehash(n.left.hash, n.hash)
		} else {
			n.parent.rehash(n.hash, n.right.hash)
		}
	}

	return nil
}

// DynTree is designed for Merkle trees that can mutate.
type DynTree struct {
	nodes  []DynTreeNode
	root   *DynTreeNode
	leaves []*DynTreeNode
	height int
	mutex  sync.RWMutex
}

// NewDynTree creates a DynTree
func NewDynTree(initialCap int) *DynTree {
	return &DynTree{
		nodes: make([]DynTreeNode, 0, initialCap),
	}
}

// NumLeaves implements Tree.NumLeaves.
func (t *DynTree) NumLeaves() int {
	return len(t.leaves)
}

// Root implements Tree.Root.
func (t *DynTree) Root() Hash {
	return t.root.hash
}

// Leaf implements Tree.Leaf.
func (t *DynTree) Leaf(index int) Hash {
	return t.leaves[index].hash
}

// Path implements Tree.Path.
func (t *DynTree) Path(index int) Path {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if len(t.nodes) < 2 {
		return Path{}
	}

	var (
		path  = make(Path, t.height)
		node  = t.leaves[index]
		level = 0
	)

	for node.parent != nil {
		if node.left != nil {
			path[level] = HashTriplet{
				Left:   node.left.hash,
				Right:  node.hash,
				Parent: node.parent.hash,
			}
		} else {
			path[level] = HashTriplet{
				Left:   node.hash,
				Right:  node.right.hash,
				Parent: node.parent.hash,
			}
		}

		node = node.parent
		level++
	}

	return path[:level]
}

// Add adds a leaf to the tree.
func (t *DynTree) Add(leaf Hash) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.nodes = append(t.nodes, DynTreeNode{hash: leaf})
	node := &t.nodes[len(t.nodes)-1]
	t.leaves = append(t.leaves, node)

	if t.root == nil {
		t.root = node
	} else {
		left := t.leaves[len(t.leaves)-2]

		for left.parent != nil && left.parent.height == left.height+1 {
			left = left.parent
		}

		t.nodes = append(t.nodes, DynTreeNode{
			left:   left.left,
			parent: left.parent,
			height: left.height + 1,
		})

		parent := &t.nodes[len(t.nodes)-1]
		node.parent, node.left = parent, left
		left.parent, left.right = parent, node

		if left.left != nil {
			left.left.right, left.left = parent, nil
		}

		if parent.parent == nil {
			t.root = parent
		}

		if parent.height > t.height {
			t.height = parent.height
		}

		if err := parent.rehash(left.hash, leaf); err != nil {
			return err
		}
	}

	return nil
}

// Update updates a leaf of the tree.
func (t *DynTree) Update(index int, hash Hash) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	node := t.leaves[index]
	node.hash = hash

	if node.left != nil {
		return node.parent.rehash(node.left.hash, hash)
	} else if node.right != nil {
		return node.parent.rehash(hash, node.right.hash)
	}

	return nil
}