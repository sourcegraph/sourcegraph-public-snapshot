package trie

import "sort"

// frozenTrieNode encodes a prefix trie of values where each transition value is an arbitrary length
// substring. Frozen tries are built by freezing a compressed string trie (defined below).
type frozenTrieNode struct {
	children []decoratedFrozenTrieNode
}

// decoratedFrozenTrieNode pairs a frozen trie node with its own prefix and a traversal clock identifier
// and parent reference (if not a root of the trie). Prefixes are stored as map keys in string/rune tries
// for deduplication during construction - we store them as part of the slice here for better memory usage.
type decoratedFrozenTrieNode struct {
	id       int
	parentID *int
	prefix   string
	node     frozenTrieNode
}

// freezeTrie orders and maps each node in the string trie to a unique identifier and returns an
// equivalent frozen trie.
//
// The given start identifier will be the first identifier used in the resulting trie. This function
// also returns the first identifier that is not used in the construction of the trie. This is used
// to keep a unique clock across multiple constructions for the same processed code intelligence index.
func freezeTrie(n stringTrieNode, startID int) (frozenTrieNode, int) {
	return freezeTrieInternal(n, startID, nil)
}

// freezeTrieInternal orders and maps each node in the string sub-trie to a unique identifier given
// the current traversal clock `nextID`, maintained by passing it through stack frames on recursive
// calls.
func freezeTrieInternal(n stringTrieNode, nextID int, parentID *int) (node frozenTrieNode, _ int) {
	keys := make([]string, 0, len(n.children))
	for key := range n.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	children := make([]decoratedFrozenTrieNode, 0, len(keys))
	for _, key := range keys {
		id := nextID
		node, nextID = freezeTrieInternal(n.children[key], id+1, &id)
		children = append(children, decoratedFrozenTrieNode{
			id:       id,
			parentID: parentID,
			prefix:   key,
			node:     node,
		})
	}

	return frozenTrieNode{children: children}, nextID
}
