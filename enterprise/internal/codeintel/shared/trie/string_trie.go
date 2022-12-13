package trie

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

	if n.terminatesValue {
		// If we're terminating an original value at this rune node, then we can't compress it into the
		// child nodes. Emit a node here, even if it has a single child, so that we generate a stable
		// identifier for it when we freeze the trie later.
		return mapNontrivialRuneNodeToStringNode(n, "", prefix)
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
