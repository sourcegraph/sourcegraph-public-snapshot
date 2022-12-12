package codeintel

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// constructTrie constructs a prefix trie from the given set of values. The resulting trie has
// an incrementing clock identifier for each node, and stores the identifier of its parent. These
// values can be extracted by calling `searchTrie` (single query) or `TraverseTrie` (bulk query)
// with the resulting root node.
//
// The given start identifier will be the first identifier used in the resulting trie. This function
// also returns the first identifier that is not used in the construction of the trie. This is used
// to keep a unique clock across multiple constructions for the same processed code intelligence index.
func constructTrie(values []string, startID int) (frozenTrieNode, int) {
	return freezeTrie(compressTrie(constructRuneTrie(values)), startID)
}

// traverseTrie invokes the given callback for each node of the given sub-trie in a pre-order walk.
func traverseTrie(n frozenTrieNode, f func(id int, parentID *int, prefix string) error) error {
	for _, child := range n.children {
		if err := f(child.id, child.parentID, child.prefix); err != nil {
			return err
		}

		if err := traverseTrie(child.node, f); err != nil {
			return err
		}
	}

	return nil
}

// searchTrie returns the clock identifier attached to the node that terminates with the given value.
// If no such value exists in the trie, this function returns a false-valued flag.
func searchTrie(n frozenTrieNode, value string) (int, bool) {
	for _, child := range n.children {
		if !strings.HasPrefix(value, child.prefix) {
			continue
		}

		if len(value) == len(child.prefix) {
			return child.id, true
		}

		if id, ok := searchTrie(child.node, value[len(child.prefix):]); ok {
			return id, ok
		}
	}

	return 0, false
}

//
// Frozen Trie

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

//
// String Trie

// stringTrieNode encodes a prefix trie of values where each transition value is an arbitrary length
// substring. String tries are built by compressing a rune trie (defined below).
type stringTrieNode struct {
	children map[string]stringTrieNode
}

// minimumSegmentLength is the minimum length of a non-leaf internal transition value to construct when
// compressing a rune trie into a string trie. Prefixes smaller than this threshold will be repeated on
// each child node. This increases total text size/repetition but decreases the number of total nodes.
const minimumSegmentLength = 16

// compressTrie compresses the given rune trie and returns an equivalent string trie.
func compressTrie(n runeTrieNode) stringTrieNode {
	return stringTrieNode{children: compressTrieInternal(n, "")}
}

// compressTrieInternal compresses the given rune sub-trie and returns an equivalent string sub-trie.
func compressTrieInternal(n runeTrieNode, prefix string) map[string]stringTrieNode {
	if len(n.children) == 0 {
		if prefix != "" {
			// We pushed a prefix all the way to the leaf; create a node
			return map[string]stringTrieNode{prefix: {}}
		}

		return nil
	}

	if len(n.children) == 1 {
		for childPrefix, child := range n.children {
			// Collapse linear runs of the tree into a single node
			return compressTrieInternal(child, prefix+string(childPrefix))
		}
	}

	if len(prefix) < minimumSegmentLength {
		// The prefix is smaller than the threshold, so we append it to each child
		return mapNontrivialRuneNodeToStringNode(n, prefix, "")
	}

	// The prefix exceeds the threshold, so we create a new shared prefix node
	return mapNontrivialRuneNodeToStringNode(n, "", prefix)
}

// mapNontrivialRuneNodeToStringNode compresses the given rune sub-trie and returns an equivalent string sub-trie.
// The given inlined prefix is appended to each of the constructed children forming the roots of this sub-trie. The
// given shared prefix, if non-empty, will be the value of the new parent node created above each of the constructed
// children.
func mapNontrivialRuneNodeToStringNode(n runeTrieNode, inlinedPrefix, sharedPrefix string) map[string]stringTrieNode {
	children := map[string]stringTrieNode{}
	for childPrefix, child := range n.children {
		for prefix, node := range compressTrieInternal(child, inlinedPrefix+string(childPrefix)) {
			children[prefix] = node
		}
	}

	if sharedPrefix == "" {
		// Return children as-is
		return children
	}

	// Create a new parent with the shared prefix
	return map[string]stringTrieNode{sharedPrefix: {children: children}}
}

//
// Rune Trie

// runeTrieNode encodes a prefix trie of values where each transition value is a single character.
type runeTrieNode struct {
	children map[rune]runeTrieNode
}

// constructRuneTrie constructs a rune trie from the given values.
func constructRuneTrie(values []string) runeTrieNode {
	root := runeTrieNode{children: map[rune]runeTrieNode{}}
	for _, value := range values {
		root = runeTrieInsert(root, value)
	}

	return root
}

// runeTrieInsert recursively inserts the runes composing the given value into the trie rooted at
// the given node. Each layer peels off the leading character and inserts it into the trie if it
// does not already exist. The remainder of the value is then inserted into the matching, possibly
// new, child node.
func runeTrieInsert(n runeTrieNode, value string) runeTrieNode {
	if len(value) == 0 {
		return n
	}

	head, size := utf8.DecodeRuneInString(value)
	child, ok := n.children[head]
	if !ok {
		child = runeTrieNode{children: map[rune]runeTrieNode{}}
	}

	n.children[head] = runeTrieInsert(child, value[size:])
	return n
}
