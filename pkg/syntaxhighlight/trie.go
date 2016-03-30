package syntaxhighlight

import (
	"errors"
)

// trie branch
type branch struct {
	// matching character
	V byte `json:"v"`
	// reference to trie object
	T *trie `json:"t"`
}

// Trie structure
type trie struct {
	// children
	Children []branch `json:"children"`
	// indicates that current object is a terminal node
	End bool `json:"end"`
}

// Callback that defines if we should continue our search or stop at the current position
// For example, "abcd" contains prefix "abc" but we may want to accept only "abc<WORDBREAK> cases
type lookupFunc func(len int) bool

// Constructs new trie object
func newTrie() *trie {
	return &trie{Children: make([]branch, 0), End: true}
}

// Inserts new prefix into trie.
// WARNING: Implementation does not allow inserting prefix where
// shorter or longer form already present there. For example
// you cannot insert "ab" after "a" or "a" after "ab".
// We expecting each tree path to contain only one terminal node
// otherwise there might be a collision when there are both
// "instanceof" and "in" prefixes
func (root *trie) insert(prefix string) error {
	node := root
	for _, r := range prefix {
		b := byte(r)
		ti, ok := findBranch(node.Children, b)
		if !ok {
			ti = &trie{Children: make([]branch, 0)}
			node.Children = append(node.Children, branch{V: b, T: ti})
		} else {
			if ti.End {
				return errors.New("Shorter prefix exists for " + prefix)
			}
		}
		node = ti
	}
	node.End = true
	if len(node.Children) > 0 {
		return errors.New("Longer prefix exists for " + prefix)
	}
	return nil
}

// Returns length of found prefix in a given slice or -1
// if there is no matching prefix. fn is used to determine
// if there was a match at the termnial position.
// For example we may expect that "interface" should be followed
// by the word break (otherwise it's a name "interfacexxx", not a keyword)
// while "<<=" may be followed by any character
func (root *trie) lookup(s []byte, fn lookupFunc) int {
	node := root
	for pos, r := range s {
		ti, ok := findBranch(node.Children, r)
		if !ok {
			return -1
		}
		node = ti
		if node.End && fn(pos+1) {
			return pos + 1
		}
	}
	return -1
}

// Returns associated trie if there is a branch that matches given byte
func findBranch(branches []branch, value byte) (*trie, bool) {
	for _, branch := range branches {
		if branch.V == value {
			return branch.T, true
		}
	}
	return nil, false
}
