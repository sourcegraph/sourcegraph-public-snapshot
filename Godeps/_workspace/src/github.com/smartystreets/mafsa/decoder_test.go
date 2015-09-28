package mafsa

import "testing"

func TestDecoder(t *testing.T) {
	mtree, err := new(Decoder).Decode(encodedTestTree)
	if err != nil {
		t.Errorf("Decode produced an error: %v", err)
	}
	checkDecodedMinTree(t, mtree)
}

func TestDecoderAgain(t *testing.T) {
	// This tree contains "ab" and "ac"
	input := []byte{
		0x01, 0x06, 0x01, 0x04, 0x00, 0x00,
		0x61, 0x02, 0x00, 0x00, 0x00, 0x02,
		0x62, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x63, 0x03, 0x00, 0x00, 0x00, 0x00,
	}

	mtree, err := new(Decoder).Decode(input)
	if err != nil {
		t.Errorf("Decode produced an error: %v", err)
	}

	a := mtree.Root.Edges['a']
	if a.Edges['b'] != a.Edges['c'] {
		t.Errorf("In tree with 'ab' and 'ac', 'b' and 'c' edges should go to the same node: %p, %p",
			a.Edges['b'], a.Edges['c'])
	}
}

func TestDecoderWhenInputTooShort(t *testing.T) {
	_, err := new(Decoder).Decode([]byte{0x01, 0x06, 0x01})
	if err == nil {
		t.Errorf("Decoder with fewer than 4 bytes input should return an error and not panic")
	}
}

func TestDecoderWithBadCharLength(t *testing.T) {
	charLength := byte(0x03)
	_, err := new(Decoder).Decode([]byte{0x01, 0x05, charLength, 0x02, 0x00, 0x61, 0x03, 0x00, 0x00, 0x00})
	if err == nil {
		t.Errorf("Decoder given input specifying the character length as 3, a bad length, should return an error")
	}
}

func TestDecoderWithBadPtrSize(t *testing.T) {
	ptrSize := byte(0x05)
	_, err := new(Decoder).Decode([]byte{0x01, 0x06, 0x02, ptrSize, 0x00, 0x00, 0x61, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00})
	if err == nil {
		t.Errorf("Decoder given input specifying pointer size of 5, a bad size, should return an error")
	}
}

func TestDecoderNumbering(t *testing.T) {
	// The numbers in these tests are ordered by the
	// nodes as printed by the MinTree's String() function.
	tests := []struct {
		words   []string
		numbers []int
	}{
		{
			words:   []string{"ab", "ac"},
			numbers: []int{2, 0, 0},
		},
		{
			words:   []string{"cite", "cited", "cities", "city"},
			numbers: []int{4, 4, 4, 1, 0, 1, 1, 0, 0},
		},
		{
			words:   []string{"cities", "city", "pities", "pity"},
			numbers: []int{2, 2, 2, 1, 1, 0, 0, 2, 2, 2, 1, 1, 0, 0},
		},
		{
			words:   []string{"overplay", "overplayed", "overplaying", "overplays", "overwork", "overworked", "overworking", "overworks", "replay", "replayed", "replaying", "replays", "rework", "reworked", "reworking", "reworks"},
			numbers: []int{8, 8, 8, 8, 4, 4, 4, 3, 1, 0, 1, 1, 0, 0, 4, 4, 4, 3, 1, 0, 1, 1, 0, 0, 8, 8, 4, 4, 4, 3, 1, 0, 1, 1, 0, 0, 4, 4, 4, 3, 1, 0, 1, 1, 0, 0},
		},
		{
			words:   []string{"hello", "jello", "zounds"},
			numbers: []int{1, 1, 1, 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 0},
		},
	}

	for testIdx, test := range tests {
		outtree := minTreeFromWordList(t, test.words)
		numbers := minTreeNumbers(t, outtree.Root, []int{})

		if len(numbers) != len(test.numbers) {
			t.Errorf("Number slice was different length (different number of nodes than expected?). Expected %d numbers, but had %d.",
				len(test.numbers), len(numbers))
		}

		for i, num := range test.numbers {
			if num != numbers[i] {
				t.Errorf("On test %v (test %d):\nExpected: %v\nActual:   %v",
					test.words, testIdx, test.numbers, numbers)
				break
			}
		}
	}
}

// Gets the numbers out of the min tree
func minTreeNumbers(t *testing.T, node *MinTreeNode, numbers []int) []int {
	keys := node.OrderedEdges()
	for _, char := range keys {
		child := node.Edges[char]
		numbers = minTreeNumbers(t, child, append(numbers, child.Number))
	}
	return numbers
}

// Checks the decoded min tree found in encoder_test.go
// for the proper entries.
func checkDecodedMinTree(t *testing.T, mtree *MinTree) {
	if !mtree.Contains("jello") {
		t.Errorf("Resulting tree should contain 'jello', but it didn't")
	}
	if !mtree.Contains("hello") {
		t.Errorf("Resulting tree should contain 'hello', but it didn't")
	}
	if !mtree.Contains("dog") {
		t.Errorf("Resulting tree should contain 'dog', but it didn't")
	}
	if !mtree.Contains("dogs") {
		t.Errorf("Resulting tree should contain 'dogs', but it didn't")
	}
	if mtree.Contains("do") {
		t.Errorf("Resulting tree should NOT contain 'do', but it did")
	}
	if mtree.Contains("") {
		t.Errorf("Resulting tree should NOT contain '', but it did")
	}
}
