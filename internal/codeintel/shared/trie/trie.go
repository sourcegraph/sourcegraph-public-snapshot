package trie

import "strings"

type Trie interface {
	Search(value string) (int, bool)
	Traverse(func(id int, parentID *int, prefix string) error) error
}

// NewTrie constructs a prefix trie from the given set of values. The resulting trie has an
// incrementing clock identifier for each node, and stores the identifier of its parent. These
// values can be extracted by calling `search` (single query) or `traverse` (bulk query) with the
// resulting root node.
//
// The given start identifier will be the first identifier used in the resulting trie. This function
// also returns the first identifier that is not used in the construction of the trie. This is used
// to keep a unique clock across multiple constructions for the same processed code intelligence index.
func NewTrie(values []string, startID int) (Trie, int) {
	return freezeTrie(compressTrie(constructRuneTrie(values)), startID)
}

func (n frozenTrieNode) Search(value string) (int, bool) {
	return search(n, value)
}

// search returns the clock identifier attached to the node that terminates with the given value.
// If no such value exists in the trie, this function returns a false-valued flag.
func search(n frozenTrieNode, value string) (int, bool) {
	for _, child := range n.children {
		if !strings.HasPrefix(value, child.prefix) {
			continue
		}

		if len(value) == len(child.prefix) {
			return child.id, true
		}

		if id, ok := search(child.node, value[len(child.prefix):]); ok {
			return id, ok
		}
	}

	return 0, false
}

func (n frozenTrieNode) Traverse(f func(id int, parentID *int, prefix string) error) error {
	return traverse(n, f)
}

// traverse invokes the given callback for each node of the given sub-trie in a pre-order walk.
func traverse(n frozenTrieNode, f func(id int, parentID *int, prefix string) error) error {
	for _, child := range n.children {
		if err := f(child.id, child.parentID, child.prefix); err != nil {
			return err
		}

		if err := traverse(child.node, f); err != nil {
			return err
		}
	}

	return nil
}
