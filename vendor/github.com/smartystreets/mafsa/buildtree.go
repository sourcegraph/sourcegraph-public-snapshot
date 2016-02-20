package mafsa

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type (
	// BuildTree implements an MA-FSA that contains all the "scaffolding" (state)
	// necessary to perform optimizations after inserting each item, which prevents
	// the tree from growing too big.
	BuildTree struct {
		Root         *BuildTreeNode
		idCounter    int
		nodeCount    uint
		register     map[string]*BuildTreeNode
		previousWord []rune
	}

	// BuildTreeNode is a node in a BuildTree that contains all the
	// info necessary to be a part of optimizations during item insertions.
	BuildTreeNode struct {
		Edges        map[rune]*BuildTreeNode
		id           int
		char         rune
		lastChildKey rune
		final        bool
		bytePos      int
	}
)

// Insert adds val to the tree and performs optimizations to minimize
// the number of nodes in the tree. The inserted val must be
// lexicographically equal to or higher than the last inserted val.
func (t *BuildTree) Insert(val string) error {
	if val < string(t.previousWord) {
		return errors.New("Insertions must be performed in lexicographical order")
	}

	word := []rune(val)

	// Establish prefix shared between this and the last word
	commonPrefixLen := 0
	lim := len(word)
	if len(t.previousWord) < lim {
		lim = len(t.previousWord)
	}
	for i := 0; i < lim; i++ {
		if word[i] != t.previousWord[i] {
			break
		}
		commonPrefixLen++
	}
	t.previousWord = word
	commonPrefix := word[:commonPrefixLen]

	// Traverse the tree up to the differing part (suffix)
	lastState := t.Traverse(commonPrefix)

	// Perform optimization steps
	if lastState.hasChildren() {
		t.replaceOrRegister(lastState)
	}

	// Add the differing part (suffix) to the tree
	currentSuffix := word[commonPrefixLen:]
	t.addSuffix(lastState, currentSuffix)

	return nil
}

// Finish completes the optimizations on a tree. You must call Finish
// at least once, like immediately after all entries have been inserted.
func (t *BuildTree) Finish() {
	t.replaceOrRegister(t.Root)
}

// addSuffix adds a sequence of characters to the tree starting at lastState.
func (t *BuildTree) addSuffix(lastState *BuildTreeNode, suffix []rune) {
	node := lastState
	for _, char := range suffix {
		newNode := &BuildTreeNode{
			Edges: make(map[rune]*BuildTreeNode),
			char:  char,
			id:    t.idCounter,
		}
		node.Edges[char] = newNode
		node.lastChildKey = char
		node = newNode
		t.idCounter++
		t.nodeCount++
	}
	node.final = true
}

// replaceOrRegister minimizes the number of nodes in the tree
// starting with leaf nodes below state.
func (t *BuildTree) replaceOrRegister(state *BuildTreeNode) {
	child := state.Edges[state.lastChildKey]

	if child.hasChildren() {
		t.replaceOrRegister(child)
	}
	// If there exists a state q in the tree such that
	// it is in the register and equivalent
	// to (duplicate of) the child:
	// 	1) Set the state's lastChildKey to q
	//	2) delete child
	// Otherwise, add child to the register.
	// (Deleting the child is implicitly garbage-collected.)
	childHash := child.hash()
	if equiv, ok := t.register[childHash]; ok {
		state.Edges[equiv.char] = equiv
		t.nodeCount--
	} else {
		t.register[childHash] = child
	}
}

// hasChildren returns true if the node has children, false otherwise.
func (tn *BuildTreeNode) hasChildren() bool {
	return len(tn.Edges) > 0
}

// hash returns a string representation of this node's equivalence class,
// not the unique node itself. A node is equivalent to another node if:
//  - Their incoming edge is the same (their char value is equal)
//  - They are both final or both nonfinal
//  - They have the same number of outgoing edges
//  - Their outgoing edges go to the same nodes, respectively
func (tn *BuildTreeNode) hash() string {
	hash := string(tn.char)
	if tn.final {
		hash += "1"
	} else {
		hash += "0"
	}

	// Iterating a map is not deterministic in its ordering,
	// so we have to copy the values into a slice and sort it.
	tmp := make([]string, 0, len(tn.Edges))
	for char, child := range tn.Edges {
		tmp = append(tmp, string(char)+strconv.Itoa(child.id))
	}
	sort.Strings(tmp)
	hash += strings.Join(tmp, "_")

	return hash
}

// Save encodes the MA-FSA into a binary format and writes it to a file.
// The tree can later be restored by calling the package's Load function.
func (t *BuildTree) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := t.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// String returns a roughly human-readable string representing
// the basic structure of the tree. For debugging only. Do not
// use with very large trees.
func (t *BuildTree) String() string {
	return t.recursiveString(t.Root, "", 0)
}

// recursiveString travels every node starting at node and builds the
// string representation of the tree, returning it. The level is how many
// nodes deep into the tree this iteration is starting at.
func (t *BuildTree) recursiveString(node *BuildTreeNode, str string, level int) string {
	keys := node.OrderedEdges()
	for _, char := range keys {
		child := node.Edges[char]
		str += fmt.Sprintf("%s%s %p\n", strings.Repeat(" ", level), string(child.char), child)
		str = t.recursiveString(child, str, level+1)
	}
	return str
}

// Contains returns true if word is found in the tree, false otherwise.
func (t *BuildTree) Contains(word string) bool {
	result := t.Traverse([]rune(word))
	return result != nil && result.final
}

// Traverse follows nodes down the tree according to word and returns the
// ending node if there was one or nil if there wasn't one. Note that
// this method alone does not indicate membership; some node may still be
// reached even if the word is not in the structure.
func (t *BuildTree) Traverse(word []rune) *BuildTreeNode {
	node := t.Root
	for i := 0; i < len(word); i++ {
		if child, ok := node.Edges[word[i]]; ok {
			node = child
		} else {
			return nil
		}
	}
	return node
}

// OrderedEdges returns the keys of the outgoing edges in
// lexigocraphical sorted order. It enables deterministic
// traversal of edges to child nodes.
func (tn *BuildTreeNode) OrderedEdges() []rune {
	keys := make(runeSlice, 0, len(tn.Edges))
	for char := range tn.Edges {
		keys = append(keys, char)
	}
	sort.Sort(keys)
	return []rune(keys)
}

// MarshalBinary encodes t into a binary format. It implements the functionality
// described by the encoding.BinaryMarhsaler interface. The associated
// encoding.BinaryUnmarshaler type is MinTree.
func (t *BuildTree) MarshalBinary() ([]byte, error) {
	return new(Encoder).Encode(t)
}
