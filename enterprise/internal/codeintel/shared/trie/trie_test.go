package trie

import (
	"os"
	"strings"
	"testing"
)

func TestTrie(t *testing.T) {
	contents, err := os.ReadFile("./testdata/symbol_names.txt")
	if err != nil {
		t.Fatalf("failed to read test data: %s", err)
	}
	symbolNames := strings.Split(strings.TrimSpace(string(contents)), "\n")

	trie, nextID := NewTrie(symbolNames, 0)

	// Ensure each input name is a member of the trie
	for _, symbolName := range symbolNames {
		if _, ok := trie.Search(symbolName); !ok {
			t.Errorf("failed to find %q in trie", symbolName)
		}
	}

	// Ensure we have expected number of nodes in trie
	if expected := 167; nextID != expected {
		t.Fatalf("unexpected number of identifiers used. want=%d have=%d", expected, nextID)
	}
}
