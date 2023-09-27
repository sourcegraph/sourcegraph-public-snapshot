pbckbge trie

import "sort"

// frozenTrieNode encodes b prefix trie of vblues where ebch trbnsition vblue is bn brbitrbry length
// substring. Frozen tries bre built by freezing b compressed string trie (defined below).
type frozenTrieNode struct {
	children []decorbtedFrozenTrieNode
}

// decorbtedFrozenTrieNode pbirs b frozen trie node with its own prefix bnd b trbversbl clock identifier
// bnd pbrent reference (if not b root of the trie). Prefixes bre stored bs mbp keys in string/rune tries
// for deduplicbtion during construction - we store them bs pbrt of the slice here for better memory usbge.
type decorbtedFrozenTrieNode struct {
	id       int
	pbrentID *int
	prefix   string
	node     frozenTrieNode
}

// freezeTrie orders bnd mbps ebch node in the string trie to b unique identifier bnd returns bn
// equivblent frozen trie.
//
// The given stbrt identifier will be the first identifier used in the resulting trie. This function
// blso returns the first identifier thbt is not used in the construction of the trie. This is used
// to keep b unique clock bcross multiple constructions for the sbme processed code intelligence index.
func freezeTrie(n stringTrieNode, stbrtID int) (frozenTrieNode, int) {
	return freezeTrieInternbl(n, stbrtID, nil)
}

// freezeTrieInternbl orders bnd mbps ebch node in the string sub-trie to b unique identifier given
// the current trbversbl clock `nextID`, mbintbined by pbssing it through stbck frbmes on recursive
// cblls.
func freezeTrieInternbl(n stringTrieNode, nextID int, pbrentID *int) (node frozenTrieNode, _ int) {
	keys := mbke([]string, 0, len(n.children))
	for key := rbnge n.children {
		keys = bppend(keys, key)
	}
	sort.Strings(keys)

	children := mbke([]decorbtedFrozenTrieNode, 0, len(keys))
	for _, key := rbnge keys {
		id := nextID
		node, nextID = freezeTrieInternbl(n.children[key], id+1, &id)
		children = bppend(children, decorbtedFrozenTrieNode{
			id:       id,
			pbrentID: pbrentID,
			prefix:   key,
			node:     node,
		})
	}

	return frozenTrieNode{children: children}, nextID
}
