package mafsa

import (
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
)

// Decoder is a type which can decode a byte slice into a MinTree.
type Decoder struct {
	fileVer int
	wordLen int
	charLen int
	ptrLen  int
	nodeMap map[int]*MinTreeNode
	tree    *MinTree
}

// Decode transforms the binary serialization of a MA-FSA into a
// new MinTree (a read-only MA-FSA).
func (d *Decoder) Decode(data []byte) (*MinTree, error) {
	tree := newMinTree()
	return tree, d.decodeMinTree(tree, data)
}

// ReadFrom reads the binary serialization of a MA-FSA into a
// new MinTree (a read-only MA-FSA) from a io.Reader.
func (d *Decoder) ReadFrom(r io.Reader) (*MinTree, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	tree := newMinTree()
	return tree, d.decodeMinTree(tree, data)
}

// decodeMinTree transforms the binary serialization of a MA-FSA into a
// read-only MA-FSA pointed to by t.
func (d *Decoder) decodeMinTree(t *MinTree, data []byte) error {
	if len(data) < 4 {
		return errors.New("Not enough bytes")
	}

	// First word contains some file format information
	d.fileVer = int(data[0])
	d.wordLen = int(data[1])
	d.charLen = int(data[2])
	d.ptrLen = int(data[3])

	// The node map translates from byte slice offsets to
	// actual node pointers in the resulting tree
	d.nodeMap = make(map[int]*MinTreeNode)

	// The node map is only needed during decoding
	defer func() {
		d.nodeMap = make(map[int]*MinTreeNode)
	}()

	// We need access to the tree so we can implement
	// minimal perfect hashing in the recursive function later
	d.tree = t

	// Begin decoding at the root node, which starts
	// at wordLen in the byte slice
	err := d.decodeEdge(data, t.Root, d.wordLen, []rune{})
	if err != nil {
		return err
	}

	// Traverse the tree once it is built so that each node
	// has the right number. The number represents the number
	// of words attainable by STARTING at that node.
	d.doNumbers(t.Root)

	return nil
}

// decodeEdge decodes the edge described by the word of the byte slice
// starting at offset, and adds each subsequent edge on this same
// node to parent, which is already in the tree. After adding the
// immediate child nodes to parent, it recursively follows the
// pointer at the end of the word to subsequent child nodes.
func (d *Decoder) decodeEdge(data []byte, parent *MinTreeNode, offset int, entry []rune) error {
	for i := offset; i < len(data); i += d.wordLen {
		// Break the word apart into the pieces we need
		charBytes := data[i : i+d.charLen]
		flags := data[i+d.charLen]
		ptrBytes := data[i+d.charLen+1 : i+d.wordLen]

		final := flags&endOfWord == endOfWord
		lastChild := flags&endOfNode == endOfNode

		char, err := d.decodeCharacter(charBytes)
		if err != nil {
			return err
		}

		ptr, err := d.decodePointer(ptrBytes)
		if err != nil {
			return err
		}

		// If this word/edge points to a node we haven't
		// seen before, add it to the node map
		if _, ok := d.nodeMap[ptr]; !ok {
			d.nodeMap[ptr] = &MinTreeNode{
				Edges: make(map[rune]*MinTreeNode),
				Final: final,
			}
		}

		// Add edge to node
		parent.Edges[char] = d.nodeMap[ptr]
		entry := append(entry, char) // TODO: Ugh, redeclaring entry seems weird here, but it's necessary, no?

		// If there are edges to other nodes, decode them
		if ptr > 0 {
			d.decodeEdge(data, d.nodeMap[ptr], ptr*d.wordLen, entry)
		}

		// If this word represents the last outgoing edge
		// for this node, stop iterating the file at this level
		if lastChild {
			break
		}
	}

	return nil
}

// doNumbers sets the number on this node to the number
// of entries accessible by starting at this node.
func (d *Decoder) doNumbers(node *MinTreeNode) {
	if node.Number > 0 {
		// We've already visited this node
		return
	}
	for _, child := range node.Edges {
		// A node's number is the sum of the
		// numbers of its immediate child nodes.
		// (The paper did not explicitly state
		// this, but as it turns out, that's the
		// rule.)
		d.doNumbers(child)
		if child.Final {
			node.Number++
		}
		node.Number += child.Number
	}
}

// decodeCharacter converts a byte slice containing a character
// value to a rune which can be used as a map key or elsewhere.
func (d *Decoder) decodeCharacter(charBytes []byte) (rune, error) {
	switch d.charLen {
	case 1:
		return rune(charBytes[0]), nil
	case 2:
		return rune(binary.BigEndian.Uint16(charBytes)), nil
	case 4:
		return rune(binary.BigEndian.Uint32(charBytes)), nil
	default:
		return 0, errors.New("Character must be encoded as 1, 2, or 4 bytes")
	}
}

// decodePointer converts a byte slice containing a number to
// the offset in the byte array where the next child is to
// an int that can be used to index into the byte slice.
func (d *Decoder) decodePointer(ptrBytes []byte) (int, error) {
	switch d.ptrLen {
	case 2:
		return int(binary.BigEndian.Uint16(ptrBytes)), nil
	case 4:
		return int(binary.BigEndian.Uint32(ptrBytes)), nil
	case 8:
		return int(binary.BigEndian.Uint64(ptrBytes)), nil
	default:
		return 0, errors.New("Child offset pointer must be only be 2, 4, or 8 bytes long")
	}
}

const (
	endOfWord = 1 << iota
	endOfNode
)
