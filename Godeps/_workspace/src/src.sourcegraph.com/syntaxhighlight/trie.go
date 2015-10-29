package syntaxhighlight

// trie branch
type branch struct {
	// matching character
	v byte
	// reference to trie object
	t *trie
}

// Trie structure
type trie struct {
	// children
	children []branch
	// indicates that current object is a terminal node
	end bool
}

// Callback that defines if we should continue our search or stop at the current position
// For example, "abcd" contains prefix "abc" but we may want to accept only "abc<WORDBREAK> cases
type lookupFunc func(len int) bool

// Constructs new trie object
func newTrie() *trie {
	return &trie{children: make([]branch, 0), end: true}
}

// Inserts new prefix into trie.
// WARNING: Implementation does not allow inserting prefix where
// shorter or longer form already present there. For example
// you cannot insert "ab" after "a" or "a" after "ab".
// We expecting each tree path to contain only one terminal node
// otherwise there might be a collision when there are both
// "instanceof" and "in" prefixes
func (root *trie) insert(prefix string) {
	node := root
	for _, r := range prefix {
		b := byte(r)
		ti, ok := findBranch(node.children, b)
		if !ok {
			ti = &trie{children: make([]branch, 0)}
			node.children = append(node.children, branch{v: b, t: ti})
		} else {
			if ti.end {
				panic("Shorter prefix exists for " + prefix)
			}
		}
		node = ti
	}
	node.end = true
	if len(node.children) > 0 {
		panic("Longer prefix exists for " + prefix)
	}
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
		ti, ok := findBranch(node.children, r)
		if !ok {
			return -1
		}
		node = ti
		if node.end && fn(pos+1) {
			return pos + 1
		}
	}
	return -1
}

// Returns associated trie if there is a branch that matches given byte
func findBranch(branches []branch, value byte) (*trie, bool) {
	for _, branch := range branches {
		if branch.v == value {
			return branch.t, true
		}
	}
	return nil, false
}
