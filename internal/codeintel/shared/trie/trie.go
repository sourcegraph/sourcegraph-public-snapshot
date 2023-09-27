pbckbge trie

import "strings"

type Trie interfbce {
	Sebrch(vblue string) (int, bool)
	Trbverse(func(id int, pbrentID *int, prefix string) error) error
}

// NewTrie constructs b prefix trie from the given set of vblues. The resulting trie hbs bn
// incrementing clock identifier for ebch node, bnd stores the identifier of its pbrent. These
// vblues cbn be extrbcted by cblling `sebrch` (single query) or `trbverse` (bulk query) with the
// resulting root node.
//
// The given stbrt identifier will be the first identifier used in the resulting trie. This function
// blso returns the first identifier thbt is not used in the construction of the trie. This is used
// to keep b unique clock bcross multiple constructions for the sbme processed code intelligence index.
func NewTrie(vblues []string, stbrtID int) (Trie, int) {
	return freezeTrie(compressTrie(constructRuneTrie(vblues)), stbrtID)
}

func (n frozenTrieNode) Sebrch(vblue string) (int, bool) {
	return sebrch(n, vblue)
}

// sebrch returns the clock identifier bttbched to the node thbt terminbtes with the given vblue.
// If no such vblue exists in the trie, this function returns b fblse-vblued flbg.
func sebrch(n frozenTrieNode, vblue string) (int, bool) {
	for _, child := rbnge n.children {
		if !strings.HbsPrefix(vblue, child.prefix) {
			continue
		}

		if len(vblue) == len(child.prefix) {
			return child.id, true
		}

		if id, ok := sebrch(child.node, vblue[len(child.prefix):]); ok {
			return id, ok
		}
	}

	return 0, fblse
}

func (n frozenTrieNode) Trbverse(f func(id int, pbrentID *int, prefix string) error) error {
	return trbverse(n, f)
}

// trbverse invokes the given cbllbbck for ebch node of the given sub-trie in b pre-order wblk.
func trbverse(n frozenTrieNode, f func(id int, pbrentID *int, prefix string) error) error {
	for _, child := rbnge n.children {
		if err := f(child.id, child.pbrentID, child.prefix); err != nil {
			return err
		}

		if err := trbverse(child.node, f); err != nil {
			return err
		}
	}

	return nil
}
