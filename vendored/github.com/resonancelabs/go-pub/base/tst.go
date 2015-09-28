package base

import ()

// Implements a ternary search tree as described by Jon Bentley and Robert Sedgewick
// in "Fast Algorithms for Sorting and Searching Strings".
//
// We make a single modification, which is to construct the tree such that the first level is
// actually a multi-way/binary trie.
// Additional improvements can be made by compressing the edges leading to leaf nodes (compress
// when there is only a single child along the whole path).
// Further improvements can be made by compressing sequences of internal edges like a patricia tree.
//
// TODO(pete) we don't actually clean up the tree when we remove key/value pairs.  We should do so
// if we anticipate a lot of churn in the datastructure (we currently don't).  Also if we do cleanup
// make sure to add nodes back to the free list.
//

const (
	kNodeAllocSize = 128
)

type leaf struct {
	key    string
	values []interface{}
}

type tstNode struct {
	split byte
	leaf  *leaf // indicates an actual word ends here
	eqkid *tstNode
	lokid *tstNode
	hikid *tstNode
}

type TernarySearchTree struct {
	roots map[byte]*tstNode
}

func NewTernarySearchTree() *TernarySearchTree {
	return &TernarySearchTree{
		roots: make(map[byte]*tstNode),
	}
}

// Inserts a new key:value pair into the tree (previous values are retained).
// Empty strings are not accepted.
func (self *TernarySearchTree) Insert(s string, v interface{}) {
	if len(s) == 0 {
		panic("empty strings not accepted in search tree")
	}

	// first node is actually a digital trie
	pn := self.roots[s[0]]
	if pn == nil {
		pn = &tstNode{split: s[0]}
		self.roots[pn.split] = pn
	}

	// Iterate over bytes remaining in string
	for i := 1; i < len(s); {
		b := s[i]

		var nn **tstNode
		if b < pn.split {
			nn = &pn.lokid
		} else if b > pn.split {
			nn = &pn.hikid
		} else {
			nn = &pn.eqkid
			i++
		}

		if *nn == nil {
			// Start inserting new nodes (we could also just iterate over the remainder of the string, inserting new nodes)
			// Or if we do edge compression, we could make a final node right here and mark it as having multiple runes.
			*nn = &tstNode{split: b}
		}

		pn = *nn
	}

	if pn.leaf == nil {
		pn.leaf = &leaf{key: s, values: []interface{}{v}}
	} else {
		pn.leaf.values = append(pn.leaf.values, v)
	}
}

func (self *TernarySearchTree) findNode(s string) *tstNode {
	// first node is actually a digital trie
	pn := self.roots[s[0]]
	if pn == nil {
		pn = &tstNode{split: s[0]}
		self.roots[r] = pn
	}

	// Iterate over bytes remaining in string
	for i := 1; i < len(s) && pn != nil; {
		if b := s[i]; b < pn.split {
			pn = pn.lokid
		} else if b > pn.split {
			pn = pn.hikid
		} else {
			pn = pn.eqkid
			i++
		}
	}

	return pn
}

// Removes a key:value pair from the tree, returns if the value existed
func (self *TernarySearchTree) Remove(s string, v interface{}) bool {
	n := self.findNode(s)
	if n != nil && n.leaf != nil {
		// Attempt to find and remove the value
		values := n.leaf.values
		for i, leafval := range values {
			if leafval == v {
				if len(n.leaf.values) == 0 {
					n.leaf = nil
				} else {
					// Shuffle values to replace this one.
					values[i] = values[len(values)-1]
					n.leaf.values = values[0 : len(values)-1]
				}

				return true
			}
		}
	}

	return false
}

// Returns if the tree contains the given key (as a true key, not a partial match)
func (self *TernarySearchTree) ContainsKey(s string) bool {
	return self.ExactMatches(s) != nil
}

// Returns all the exact matches of that string
func (self *TernarySearchTree) ExactMatches(s string) []interface{} {
	n := self.findNode(s)
	if n != nil && n.leaf != nil {
		return n.leaf.values
	} else {
		return nil
	}
}

// Iterates through partial matches to the specified string.  Iteration will
// be stopped when the iterator returns false.
func (self *TernarySearchTree) PartialMatches(s string, it func(matches string, values []interface{}) bool) {
	n := self.findNode(s)
	if n == nil {
		return
	}

	queue := []*tstNode{n}
	for i := 0; i < len(queue); i++ {
		n = queue[i]
		if n == nil {
			continue
		}

		if n.leaf != nil && !it(n.leaf.key, n.leaf.values) {
			return
		}
		queue = append(queue, n.lokid, n.eqkid, n.hikid)
	}
}
