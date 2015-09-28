package stringmetrics

type bkTreeNode struct {
	word     string
	distance int // from parent node's word
	children []*bkTreeNode
}

// BkTreeMetric should define the metric space over which the BK-Tree is built.
// Normally this is something like Levenshtein distance.
//
// The metric doesn't directly control the performance of the datastructure (though
// evaluation of the metric while searching the tree is a large constant factor, so
// it should be as efficient as possible).  Besides this, search performance is
// determined by how much of the tree is explored, which is directly related to the
// distance passed to FindSimilarWords and how many words the metric will classify
// as being that distance away.
type BkTreeMetric func(a, b string) (distance int)

// Implements a Burkhard-Keller tree over strings, with a customizable distance metric.
// See http://en.wikipedia.org/wiki/BK-tree and the linked papers for additional detail.
type BkTree struct {
	root   *bkTreeNode
	metric BkTreeMetric
}

func (p *BkTree) insertWord(n *bkTreeNode, w string) {
	d := p.metric(n.word, w)
	if d == 0 {
		return
	}

	// Search for an existing child of this distance, recurse if so
	for _, childNode := range n.children {
		if childNode.distance == d {
			p.insertWord(childNode, w)
			return
		}
	}

	n.children = append(n.children, &bkTreeNode{word: w, distance: d})
}

func NewBkTree(words []string, metric BkTreeMetric) *BkTree {
	if len(words) == 0 {
		return nil
	}

	// Construct the tree, picking the root node arbitrarily.
	// TODO(pete) some nodes do give better query times, but a more rigorous sampling is required
	// We could improve query time by roughly balancing the tree afterwards as well.
	tree := &BkTree{&bkTreeNode{word: words[len(words)/2]}, metric}
	for _, w := range words {
		tree.insertWord(tree.root, w)
	}

	return tree
}

// Incrementally adds new words to the tree.
func (p *BkTree) AddWord(word string) {
	p.insertWord(p.root, word)
}

// Returns a list of all words in the tree with edit-distances not greater than 'distance' away from 'w'.
func (p *BkTree) FindSimilarWords(w string, distance int) []string {
	results := []string{}

	// TODO(pete) use a ring buffer
	searchNodes := make([]*bkTreeNode, 0, 16)
	searchNodes = append(searchNodes, p.root)

	for i := 0; i < len(searchNodes); i++ {
		node := searchNodes[i]
		d := p.metric(node.word, w)
		if d <= distance {
			results = append(results, node.word)
		}

		// visit all children that are not too far away
		// TODO(pete) faster search!  I did try sorting the children by distance so we can early out instead
		// of looping through the whole array, but this slowed things down measurably (?!).  Warrants additional
		// investigation
		for _, child := range node.children {
			if child.distance >= d-distance && child.distance <= d+distance {
				searchNodes = append(searchNodes, child)
			}
		}
	}

	return results
}
