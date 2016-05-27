package mafsa

import (
	"bytes"
	"encoding/binary"
	"io"
	"sort"
)

// Encoder is a type which can encode a BuildTree into a byte slice
// which can be written to a file.
type Encoder struct {
	queue   []*BuildTreeNode
	counter int
}

// Encode serializes a BuildTree t into a byte slice.
func (e *Encoder) Encode(t *BuildTree) ([]byte, error) {
	e.queue = []*BuildTreeNode{}
	e.counter = len(t.Root.Edges) + 1

	// First "word" (fixed-length entry) is a null entry
	// that specifies the file format:
	// First byte indicates the flag scheme (basically a file format verison number)
	// Second byte is word length in bytes (at least 4)
	// Third byte is char length in bytes
	// Fourth byte is pointer length in bytes
	//   Note: Word length (the first byte)
	//   must be exactly Second byte + 1 (flags) + Fourth byte
	// Any leftover bytes in this first word are zero
	data := []byte{0x01, 0x06, 0x01, 0x04}
	for i := len(data); i < int(data[1]); i++ {
		data = append(data, 0x00)
	}

	data = e.encodeEdges(t.Root, data)

	for len(e.queue) > 0 {
		// Pop first item off the queue
		top := e.queue[0]
		e.queue = e.queue[1:]

		// Recursively marshal child nodes
		data = e.encodeEdges(top, data)
	}

	return data, nil
}

// WriteTo encodes and saves the BuildTree to a io.Writer.
func (e *Encoder) WriteTo(wr io.Writer, t *BuildTree) error {
	bs, err := e.Encode(t)
	if err != nil {
		return err
	}

	_, err = io.Copy(wr, bytes.NewReader(bs))
	if err != nil {
		return err
	}

	return nil
}

// encodeEdges encodes the edges going out of node into bytes which are appended
// to data. The modified byte slice is returned.
func (e *Encoder) encodeEdges(node *BuildTreeNode, data []byte) []byte {
	// We want deterministic output for testing purposes,
	// so we need to order the keys of the edges map.
	edgeKeys := sortEdgeKeys(node)

	for i := 0; i < len(edgeKeys); i++ {
		child := node.Edges[edgeKeys[i]]
		word := []byte(string(edgeKeys[i]))

		var flags byte
		if child.final {
			flags |= 0x01 // end of word
		}
		if i == len(edgeKeys)-1 {
			flags |= 0x02 // end of node (last child outgoing from this node)
		}

		word = append(word, flags)

		// If bytePos is 0, we haven't encoded this edge yet
		if child.bytePos == 0 {
			if len(child.Edges) > 0 {
				child.bytePos = e.counter
				e.counter += len(child.Edges)
			}
			e.queue = append(e.queue, child)
		}

		pointer := child.bytePos
		pointerBytes := make([]byte, int(data[3]))
		switch len(pointerBytes) {
		case 2:
			binary.BigEndian.PutUint16(pointerBytes, uint16(pointer))
		case 4:
			binary.BigEndian.PutUint32(pointerBytes, uint32(pointer))
		case 8:
			binary.BigEndian.PutUint64(pointerBytes, uint64(pointer))
		}

		word = append(word, pointerBytes...)

		data = append(data, word...)
	}

	return data
}

// sortEdgeKeys returns a sorted list of the keys
// of the map containing outgoing edges.
func sortEdgeKeys(node *BuildTreeNode) []rune {
	edgeKeys := make(runeSlice, 0, len(node.Edges))
	for char := range node.Edges {
		edgeKeys = append(edgeKeys, char)
	}
	sort.Sort(edgeKeys)
	return []rune(edgeKeys)
}

type runeSlice []rune

func (s runeSlice) Len() int           { return len(s) }
func (s runeSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s runeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
