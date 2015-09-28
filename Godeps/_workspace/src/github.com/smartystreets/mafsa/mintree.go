package mafsa

import (
	"fmt"
	"sort"
	"strings"
)

type (
	// MinTree implements a simple prefix tree with read-only functionality.
	// It supports Minimal Perfect Hashing (MPH) so you can lookup entries
	// by their index in a slice.
	MinTree struct {
		Root *MinTreeNode
	}

	// MinTreeNode is a node in a MinTree.
	MinTreeNode struct {
		Edges  map[rune]*MinTreeNode
		Final  bool
		Number int
	}

	// Entry represents a value in the tree along with its
	// index number (minimal perfect hashing).
	Entry struct {
		// The actual entry that was added to the tree
		Value string

		// Represents how many entries were added to the tree before this one
		Index int
	}
)

// Contains returns true if word is found in the tree,
// false otherwise.
func (t *MinTree) Contains(word string) bool {
	result := t.Traverse([]rune(word))
	return result != nil && result.Final
}

// Traverse visits nodes according to word and returns the
// last node, if there was one. Note that returning a non-nil
// node is not indicative of membership; a node may be
// returned even if the word is not in the set. Traverse
// implements a fast traversal which does not count the
// index number of the last node.
func (t *MinTree) Traverse(word []rune) *MinTreeNode {
	node := t.Root
	for i := 0; i < len(word); i++ {
		if child, ok := node.Edges[word[i]]; ok {
			node = child
		} else {
			return nil
		}
	}
	return node
}

// IndexedTraverse visits nodes according to word and
// returns the last node, if there was one, along with
// its index number. The index number represents how
// many entries, including the returned node, are in the
// tree before entries beginning with the specified prefix.
// Such an indexed traversal is slightly slower than a
// regular traversal since this kind of traversal has to
// iterate more nodes and count the numbers on each node.
// Note that returning a non-nil node is not indicative of
// membership; a node may be be returned even if the word is
// not in the set. If last node is nil, the returned index
// is -1. The algorithm is adapted directly from the paper
// by Lucceshi and Kowaltowski, 1992 (section 5.2), with
// one slight modification for correctness because of a slight
// difference in the way our tree is implemented.
func (t *MinTree) IndexedTraverse(word []rune) (*MinTreeNode, int) {
	index := 0
	node := t.Root

	for i := 0; i < len(word); i++ {
		if child, ok := node.Edges[word[i]]; ok {
			for char, child := range node.Edges {
				if char < word[i] {
					index += child.Number

					// If a previous sibling is also a final
					// node, add 1, since that word appears before
					// any word through this node. This line is
					// a modification of the algorithm described
					// in the paper.
					if child.Final {
						index++
					}
				}
			}
			node = child
			if node.Final {
				index++
			}
		} else {
			return nil, -1
		}
	}

	return node, index
}

// String serializes the tree roughly as a human-readable string
// for debugging purposes. Don't try this on large trees.
func (t *MinTree) String() string {
	return t.recursiveString(t.Root, "", 0)
}

// recursiveString is used by String to visit each node
// and build a string representation of the tree.
func (t *MinTree) recursiveString(node *MinTreeNode, str string, level int) string {
	keys := node.OrderedEdges()
	for _, char := range keys {
		child := node.Edges[char]
		str += fmt.Sprintf("%s%s %p %d\n", strings.Repeat(" ", level), string(char), child, child.Number)
		str = t.recursiveString(child, str, level+1)
	}
	return str
}

// UnmarshalBinary decodes binary data into this MinTree. It implements the functionality
// described by the encoding.BinaryMarhsaler interface. The associated
// encoding.BinaryMarshaler type is BuildTree.
func (t *MinTree) UnmarshalBinary(data []byte) error {
	return new(Decoder).decodeMinTree(t, data)
}

// OrderedEdges returns the keys of the outgoing edges in
// lexigocraphical sorted order. It enables deterministic
// traversal of edges to child nodes.
func (n *MinTreeNode) OrderedEdges() []rune {
	keys := make(runeSlice, 0, len(n.Edges))
	for char := range n.Edges {
		keys = append(keys, char)
	}
	sort.Sort(keys)
	return []rune(keys)
}

// newMinTree constructs a new, empty MA-FSA that must be assembled
// manually (e.g. when decoding from binary), but is very memory
// efficient.
func newMinTree() *MinTree {
	t := new(MinTree)
	t.Root = new(MinTreeNode)
	t.Root.Edges = make(map[rune]*MinTreeNode)
	return t
}
