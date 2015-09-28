package mafsa

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestMinTreeContains(t *testing.T) {
	mtree := newMinTree()
	minTreeLazyInsert(mtree, "cat")
	minTreeLazyInsert(mtree, "cats")
	minTreeLazyInsert(mtree, "cathode")
	minTreeLazyInsert(mtree, "erode")

	if !mtree.Contains("cat") {
		t.Errorf("Contains() should return true for 'cat', but it didn't")
	}
	if !mtree.Contains("cats") {
		t.Errorf("Contains() should return true for 'cats', but it didn't")
	}
	if !mtree.Contains("cathode") {
		t.Errorf("Contains() should return true for 'cathode', but it didn't")
	}
	if !mtree.Contains("erode") {
		t.Errorf("Contains() should return true for 'erode', but it didn't")
	}

	if mtree.Contains("cast") {
		t.Errorf("Contains() should return false for 'cast', but it didn't")
	}
	if mtree.Contains("erodes") {
		t.Errorf("Contains() should return false for 'erodes', but it didn't")
	}
}

func TestMinTreeIndexedTraverse(t *testing.T) {
	tree := minTreeFromWordList(t, []string{
		"overplay",
		"overplayed",
		"overplaying",
		"overplays",
		"overwork",
		"overworked",
		"overworking",
		"overworks",
		"replay",
		"replayed",
		"replaying",
		"replays",
		"rework",
		"reworked",
		"reworking",
		"reworks",
	})

	tests := []struct {
		input               string
		expectedNode        *MinTreeNode
		expectedIndexNumber int // represents how many words appear before words starting with this prefix
	}{
		{
			input:               "overp",
			expectedNode:        tree.Traverse([]rune("overp")),
			expectedIndexNumber: 0,
		},
		{
			input:               "repla",
			expectedNode:        tree.Traverse([]rune("repla")),
			expectedIndexNumber: 8,
		},
		{
			input:               "replay",
			expectedNode:        tree.Traverse([]rune("replay")),
			expectedIndexNumber: 9,
		},
		{
			input:               "rep",
			expectedNode:        tree.Traverse([]rune("rep")),
			expectedIndexNumber: 8,
		},
		{
			input:               "reworks",
			expectedNode:        tree.Traverse([]rune("reworks")),
			expectedIndexNumber: 16, // there are 16 words in the set, none appear after the given prefix
		},
		{
			input:               "asdf",
			expectedNode:        nil,
			expectedIndexNumber: -1,
		},
	}

	for i, test := range tests {
		actualNode, actualIndexNumber := tree.IndexedTraverse([]rune(test.input))
		if actualNode != test.expectedNode {
			t.Errorf("For '%s' (test %d), expected node %p but got %p",
				test.input, i, test.expectedNode, actualNode)
		}
		if actualIndexNumber != test.expectedIndexNumber {
			t.Errorf("For '%s' (test %d), expected index number %d but got %d",
				test.input, i, test.expectedIndexNumber, actualIndexNumber)
		}
	}

	// This next tree covers an important, extra condition
	// in the traversal algorithm that varies slightly from
	// the algorithm described in the paper, which adds 1 if
	// encountering a child node that is marked final.
	tree2 := minTreeFromWordList(t, []string{
		"ab",
		"ac",
	})

	tests2 := []struct {
		input               string
		expectedNode        *MinTreeNode
		expectedIndexNumber int // represents how many words appear before words starting with this prefix
	}{
		{
			input:               "a",
			expectedNode:        tree2.Traverse([]rune("a")),
			expectedIndexNumber: 0,
		},
		{
			input:               "ab",
			expectedNode:        tree2.Traverse([]rune("ab")),
			expectedIndexNumber: 1, // ab takes us up to the word ab so any other words starting with ab must be after ab, hence 1
		},
		{
			input:               "ac",
			expectedNode:        tree2.Traverse([]rune("ac")),
			expectedIndexNumber: 2, // ac takes us up to the word ac so any other words starting with ac must be after ac, hence 2
		},
	}

	for i, test := range tests2 {
		actualNode, actualIndexNumber := tree2.IndexedTraverse([]rune(test.input))
		if actualNode != test.expectedNode {
			t.Errorf("For '%s' (test %d), expected node %p but got %p",
				test.input, i, test.expectedNode, actualNode)
		}
		if actualIndexNumber != test.expectedIndexNumber {
			t.Errorf("For '%s' (test %d), expected index number %d but got %d",
				test.input, i, test.expectedIndexNumber, actualIndexNumber)
		}
	}
}

func TestMinTreeString(t *testing.T) {
	// a-b-c tree
	tree := minTreeFromWordList(t, []string{
		"a",
		"ab",
		"abc",
	})
	str := tree.String()
	expected := []struct {
		char   string
		node   string
		number int
	}{
		{char: "a", number: 2, node: fmt.Sprintf("%p", tree.Root.Edges['a'])},
		{char: "b", number: 1, node: fmt.Sprintf("%p", tree.Root.Edges['a'].Edges['b'])},
		{char: "c", number: 0, node: fmt.Sprintf("%p", tree.Root.Edges['a'].Edges['b'].Edges['c'])},
	}

	for i, line := range strings.Split(str, "\n") {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, " ")
		// We go from the end of parts because indentation causes empty strings on split
		if parts[len(parts)-3] != expected[i].char {
			t.Errorf("For a-b-c tree, expected char on line %d to be '%s', got '%s' instead",
				i+1, expected[i].char, parts[len(parts)-3])
		}
		if parts[len(parts)-2] != expected[i].node {
			t.Errorf("For a-b-c tree, expected pointer on line %d to be %s, got %s instead",
				i+1, expected[i].node, parts[len(parts)-2])
		}
		if parts[len(parts)-1] != strconv.Itoa(expected[i].number) {
			t.Errorf("For a-b-c tree, expected index number on line %d to be %d, got %s instead",
				i+1, expected[i].number, parts[len(parts)-1])
		}
	}

	// x-y-z tree
	tree = minTreeFromWordList(t, []string{
		"xyz",
		"yz",
		"z",
	})
	str = tree.String()
	expected = []struct {
		char   string
		node   string
		number int
	}{
		{char: "x", number: 1, node: fmt.Sprintf("%p", tree.Root.Edges['x'])},
		{char: "y", number: 1, node: fmt.Sprintf("%p", tree.Root.Edges['y'])},
		{char: "z", number: 0, node: fmt.Sprintf("%p", tree.Root.Edges['y'].Edges['z'])},
		{char: "y", number: 1, node: fmt.Sprintf("%p", tree.Root.Edges['x'].Edges['y'])},
		{char: "z", number: 0, node: fmt.Sprintf("%p", tree.Root.Edges['x'].Edges['y'].Edges['z'])},
		{char: "z", number: 0, node: fmt.Sprintf("%p", tree.Root.Edges['z'])},
	}

	for i, line := range strings.Split(str, "\n") {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, " ")
		// We go from the end of parts because indentation causes empty strings on split
		if parts[len(parts)-3] != expected[i].char {
			t.Errorf("For x-y-z tree, expected char on line %d to be '%s', got '%s' instead",
				i+1, expected[i].char, parts[len(parts)-3])
		}
		if parts[len(parts)-2] != expected[i].node {
			t.Errorf("For x-y-z tree, expected pointer on line %d to be %s, got %s instead",
				i+1, expected[i].node, parts[len(parts)-2])
		}
		if parts[len(parts)-1] != strconv.Itoa(expected[i].number) {
			t.Errorf("For x-y-z tree, expected index number on line %d to be %d, got %s instead",
				i+1, expected[i].number, parts[len(parts)-1])
		}
	}
}

func TestMinTreeUnmarshal(t *testing.T) {
	mtree := newMinTree()
	mtree.UnmarshalBinary(encodedTestTree)
	checkDecodedMinTree(t, mtree)
}

func TestMinTreeOrderedEdges(t *testing.T) {
	tree := minTreeFromWordList(t, []string{
		"fair",
		"festival",
		"fit",
		"form",
		"fun",
	})

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

// minTreeFromWordList makes a MinTree from a list of words
// by sorting the words, putting them into a BuildTree,
// encoding it, decoding it, and then returning that resulting
// MinTree. For testing only.
func minTreeFromWordList(t *testing.T, words []string) *MinTree {
	sort.Strings(words)
	intree := New()
	for _, word := range words {
		intree.Insert(word)
	}
	intree.Finish()

	enc, err := new(Encoder).Encode(intree)
	if err != nil {
		t.Fatalf("Encode produced an error before we could even test the numbers: %v", err)
	}

	outtree, err := new(Decoder).Decode(enc)
	if err != nil {
		t.Fatalf("Decode produced an error before we could even test the numbers: %v", err)
	}

	return outtree
}

// Manually inserts items into the min tree,
// with no considereation for optimization or
// putting numbers on the nodes. For testing only.
func minTreeLazyInsert(t *MinTree, word string) {
	wordRunes := []rune(word)
	node := t.Root
	for _, char := range wordRunes {
		if child, ok := node.Edges[char]; ok {
			node = child
		} else {
			node.Edges[char] = &MinTreeNode{
				Edges: make(map[rune]*MinTreeNode),
			}
			node = node.Edges[char]
		}
	}
	node.Final = true
}
