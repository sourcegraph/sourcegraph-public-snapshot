package mafsa

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestBuildTreeInsert(t *testing.T) {
	tree := New()

	err := tree.Insert("test")
	if err != nil {
		t.Fatalf("Insert should have succeeded, but it returned an error: %v", err)
	}

	if tree.nodeCount != 4 {
		t.Errorf("Node count should be exactly 4, but is %d", tree.nodeCount)
	}

	node, ok := tree.Root.Edges['t']
	if !ok || len(node.Edges) != 1 {
		t.Errorf("Root node should contain a single outgoing edge to 'e'")
	}
	if node.final {
		t.Errorf("Root node should NOT be marked as final")
	}

	node, ok = node.Edges['e']
	if !ok || len(node.Edges) != 1 {
		t.Errorf("First 't' node should contain a single outgoing edge to 'e'")
	}
	if node.final {
		t.Errorf("First 't' node should NOT be marked as final")
	}

	node, ok = node.Edges['s']
	if !ok || len(node.Edges) != 1 {
		t.Errorf("'e' node should contain a single outgoing edge to 's'")
	}
	if node.final {
		t.Errorf("'e' node should NOT be marked as final")
	}

	node, ok = node.Edges['t']
	if !ok || len(node.Edges) != 0 {
		t.Errorf("Second 't' node should NOT contain any outgoing edges")
	}
	if !node.final {
		t.Errorf("Second 't' node should be marked as final")
	}
}

func TestBuildTreeInsertBadOrdering(t *testing.T) {
	tree := New()

	err := tree.Insert("b")
	if err != nil {
		t.Errorf("The first insert should have succeeded, but it returned an error: %v", err)
	}

	err = tree.Insert("a")
	if err == nil {
		t.Errorf("The out-of-order insert should have failed, but it succeeded")
	}

	if tree.nodeCount != 1 {
		t.Errorf("Node count should be exactly 1, but is %d", tree.nodeCount)
	}
}

func TestBuildTreeContains(t *testing.T) {
	tree := New()

	err := tree.Insert("test")
	if err != nil {
		t.Fatalf("Insert should have succeeded, but it returned an error: %v", err)
	}

	if !tree.Contains("test") {
		t.Errorf("Contains should return 'true' for a word in the tree")
	}

	if tree.Contains("tess") {
		t.Errorf("Contains should return 'false' for a word not in the tree")
	}

	if tree.Contains("tes") {
		t.Errorf("Contains should return 'false' for a (shorter) word not in the tree")
	}

	if tree.Contains("tests") {
		t.Errorf("Contains should return 'false' for a (longer) word not in the tree")
	}
}

func TestBuildTreeReplaceOrRegister(t *testing.T) {
	tree := New()

	err := tree.Insert("hello")
	if err != nil {
		t.Fatalf("Insert should have succeeded, but it returned an error: %v", err)
	}

	if tree.nodeCount != 5 {
		t.Errorf("Node count should be exactly 5, but is %d", tree.nodeCount)
	}

	err = tree.Insert("jello")
	if err != nil {
		t.Fatalf("Insert should have succeeded, but it returned an error: %v", err)
	}

	tree.Finish()

	if tree.nodeCount != 6 {
		t.Errorf("Node count should be exactly 6, but is %d", tree.nodeCount)
	}

	if len(tree.Root.Edges) != 2 {
		t.Errorf("Root node should have exactly 2 children")
	}

	path1node := tree.Root.Edges['j'].Edges['e']
	path2node := tree.Root.Edges['h'].Edges['e']
	if path1node != path2node {
		t.Errorf("The second letter, e, should have folded over to be the same node.\nh-e: %p\nj-e: %p", path1node, path2node)
	}

	path1node = path1node.Edges['l']
	path2node = path2node.Edges['l']
	if path1node != path2node {
		t.Errorf("The third letter, l, should have folded over to be the same node.\nh-e-l: %p\nj-e-l: %p", path1node, path2node)
	}

	path1node = path1node.Edges['l']
	path2node = path2node.Edges['l']
	if path1node != path2node {
		t.Errorf("The fourth letter, l, should have folded over to be the same node.\nh-e-l-l: %p\nj-e-l-l: %p", path1node, path2node)
	}

	path1node = path1node.Edges['o']
	path2node = path2node.Edges['o']
	if path1node != path2node {
		t.Errorf("The fifth letter, o, should have folded over to be the same node.\nh-e-l-l-o: %p\nj-e-l-l-o: %p", path1node, path2node)
	}
}

func TestBuildTreeCommonPrefix(t *testing.T) {
	tree := New()

	err := tree.Insert("precise")
	if err != nil {
		t.Fatalf("Insert should have succeeded, but it returned an error: %v", err)
	}

	err = tree.Insert("precision")
	if err != nil {
		t.Fatalf("Insert should have succeeded, but it returned an error: %v", err)
	}

	if !tree.Contains("precise") {
		t.Errorf("'precise' should be contained in the structure, but it reportedly isn't")
	}
	if !tree.Contains("precision") {
		t.Errorf("'precision' should be contained in the structure, but it reportedly isn't")
	}

	if tree.nodeCount != 10 {
		t.Errorf("Node count should be exactly 10, but is %d", tree.nodeCount)
	}
}

func TestbuildTreeNodeCount(t *testing.T) {
	tree := New()

	for _, word := range []string{
		"city",
		"cities",
		"dog",
		"pity",
		"pities",
	} {
		tree.Insert(word)
	}

	tree.Finish()

	if tree.nodeCount != 8 {
		t.Errorf("Node count should be exactly 8, but is %d", tree.nodeCount)
	}

	// Try another set with words that have similar letters,
	// but edges common in the middle of the word ("ello")
	// don't get folded together
	tree = New()

	for _, word := range []string{
		"hello",
		"jello",
		"yellow",
	} {
		tree.Insert(word)
	}

	if tree.nodeCount != 12 {
		t.Errorf("Node count should be exactly 12, but is %d", tree.nodeCount)
	}
}

func TestBuildTreeEdgeCases(t *testing.T) {
	tree := New()

	// Inserting empty string
	err := tree.Insert("")
	if err != nil {
		t.Errorf("Inserting empty string should work, but returned an error: %v", err)
	}
	if !tree.Root.final {
		t.Errorf("With empty string in tree, root node should be marked final, but wasn't")
	}
	if !tree.Contains("") {
		t.Errorf("With empty string in tree, Contains should return true for empty string")
	}
	if tree.nodeCount != 0 {
		t.Errorf("Node count should be exactly 0 since root node doesn't count, but was %d", tree.nodeCount)
	}

	// Inserting Unicode
	err = tree.Insert("Hello, 世界")
	if err != nil {
		t.Errorf("Inserting Unicode string should work, but returned an error: %v", err)
	}
	if !tree.Contains("Hello, 世界") {
		t.Errorf("With Unicode string in tree, Contains should return true for it")
	}
	if tree.nodeCount != 9 {
		t.Errorf("Node count should be exactly 9, but was %d", tree.nodeCount)
	}
}

func TestbuildTreeNodeHash(t *testing.T) {
	tree := New()

	// It's extremely important that the hash function
	// generates a string that represents the equivalence
	// class of the node, not the unique node itself.
	// Two different nodes can, and SHOULD, hash to the same
	// digest if they are to be consolidated. In other words,
	// their unique IDs should not be considered.
	// The hash function can be implemented a number of
	// different ways and this is just one way to do it.

	for _, word := range []string{
		"cities",
		"city",
		"pities",
		"pity",
	} {
		tree.Insert(word)
	}
	tree.Finish()

	node := tree.Root.Edges['c']
	if node.hash() != "c0i1" {
		t.Errorf("Expected hash to be c0i1, but was: %s", node.hash())
	}

	node = node.Edges['i']
	if node.hash() != "i0t2" {
		t.Errorf("Expected hash to be i0t2, but was: %s", node.hash())
	}

	node = node.Edges['t']
	if node.hash() != "t0i3_y6" {
		t.Errorf("Expected hash to be t0i3_y6, but was: %s", node.hash())
	}

	node = node.Edges['y']
	if node.hash() != "y1" {
		t.Errorf("Expected hash to be y1, but was: %s", node.hash())
	}
}

func TestBuildTreeHasChildren(t *testing.T) {
	tree := New()
	tree.Insert("Fabulous")
	tree.Insert("Freedom")
	tree.Finish()

	if !tree.Root.hasChildren() {
		t.Errorf("hasChildren should return true for root node")
	}
	if !tree.Root.Edges['F'].hasChildren() {
		t.Errorf("hasChildren should return true for first ('F') node")
	}
	if tree.Traverse([]rune("Fabulous")).hasChildren() {
		t.Errorf("hasChildren should return false for a leaf node")
	}
}

func TestBuildTreeMarshalBinary(t *testing.T) {
	tree := New()
	tree.Insert("dog")
	tree.Insert("dogs")
	tree.Insert("hello")
	tree.Insert("jello")
	tree.Finish()

	actual, err := tree.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() returned an error: %v", err)
	}

	if !bytes.Equal(actual, encodedTestTree) {
		t.Errorf("\nUnmarshalBinary was wrong.\nExpected:\n%v\nGot:\n%v", encodedTestTree, actual)
	}
}

func TestBuildTreeString(t *testing.T) {
	// a-b-c tree
	tree := New()
	tree.Insert("a")
	tree.Insert("ab")
	tree.Insert("abc")
	tree.Finish()
	str := tree.String()
	expected := []struct {
		char string
		node string
	}{
		{char: "a", node: fmt.Sprintf("%p", tree.Root.Edges['a'])},
		{char: "b", node: fmt.Sprintf("%p", tree.Root.Edges['a'].Edges['b'])},
		{char: "c", node: fmt.Sprintf("%p", tree.Root.Edges['a'].Edges['b'].Edges['c'])},
	}

	for i, line := range strings.Split(str, "\n") {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, " ")
		// We go from the end of parts because indentation causes empty strings on split
		if parts[len(parts)-2] != expected[i].char {
			t.Errorf("For a-b-c tree, expected char on line %d to be '%s', got '%s' instead",
				i+1, expected[i].char, parts[len(parts)-2])
		}
		if parts[len(parts)-1] != expected[i].node {
			t.Errorf("For a-b-c tree, expected pointer on line %d to be %s, got %s instead",
				i+1, expected[i].node, parts[len(parts)-1])
		}
	}

	// x-y-z tree
	tree = New()
	tree.Insert("xyz")
	tree.Insert("yz")
	tree.Insert("z")
	tree.Finish()
	str = tree.String()
	expected = []struct {
		char string
		node string
	}{
		{char: "x", node: fmt.Sprintf("%p", tree.Root.Edges['x'])},
		{char: "y", node: fmt.Sprintf("%p", tree.Root.Edges['y'])},
		{char: "z", node: fmt.Sprintf("%p", tree.Root.Edges['y'].Edges['z'])},
		{char: "y", node: fmt.Sprintf("%p", tree.Root.Edges['x'].Edges['y'])},
		{char: "z", node: fmt.Sprintf("%p", tree.Root.Edges['x'].Edges['y'].Edges['z'])},
		{char: "z", node: fmt.Sprintf("%p", tree.Root.Edges['z'])},
	}

	for i, line := range strings.Split(str, "\n") {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, " ")
		// We go from the end of parts because indentation causes empty strings on split
		if parts[len(parts)-2] != expected[i].char {
			t.Errorf("For x-y-z tree, expected char on line %d to be '%s', got '%s' instead",
				i+1, expected[i].char, parts[len(parts)-2])
		}
		if parts[len(parts)-1] != expected[i].node {
			t.Errorf("For x-y-z tree, expected pointer on line %d to be %s, got %s instead",
				i+1, expected[i].node, parts[len(parts)-1])
		}
	}
}

func TestBuildTreeOrderedEdges(t *testing.T) {
	tree := New()
	tree.Insert("fair")
	tree.Insert("festival")
	tree.Insert("fit")
	tree.Insert("form")
	tree.Insert("fun")
	tree.Finish()

	expected := []rune{'a', 'e', 'i', 'o', 'u'}

	for i := 0; i < 100; i++ {
		errored := false
		actual := tree.Root.Edges['f'].OrderedEdges()
		if len(actual) != len(expected) {
			t.Errorf("Ordered keys not correct.\nExpected: %s\nActual:   %s",
				string(expected), string(actual))
			break
		}
		for i, key := range actual {
			if key != expected[i] {
				t.Errorf("Ordered keys not ordered properly.\nExpected: %s\nActual:   %s",
					string(expected), string(actual))
				errored = true
				break
			}
		}
		if errored {
			break
		}
	}
}
