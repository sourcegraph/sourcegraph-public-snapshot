package trie

import "unicode/utf8"

// runeTrieNode encodes a prefix trie of values where each transition value is a single character.
type runeTrieNode struct {
	terminatesValue bool
	children        map[rune]runeTrieNode
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
		n.terminatesValue = true
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
