package trie

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestTrie(t *testing.T) {
	testCases := []struct {
		name             string
		expectedNumNodes int
	}{
		{name: "lsif", expectedNumNodes: 167},
		{name: "scip", expectedNumNodes: 22},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			values := readTrieTestInput(t, testCase.name)
			trie, _ := NewTrie(values, 0)

			// Ensure each input name is a member of the trie
			for _, value := range values {
				if _, ok := trie.Search(value); !ok {
					t.Errorf("failed to find %q in trie", value)
				}
			}

			// Ensure each trie can reconstruct the full inputs
			valuesByID := map[int]string{}
			if err := trie.Traverse(func(id int, parentID *int, prefix string) error {
				if parentID == nil {
					valuesByID[id] = prefix
				} else {
					parentPrefix, ok := valuesByID[*parentID]
					if !ok {
						return errors.Newf("parent referenced before visit: %d", *parentID)
					}

					valuesByID[id] = parentPrefix + prefix
				}

				return nil
			}); err != nil {
				t.Fatalf("unexpected error traversing trie: %s", err)
			}

			valueMap := map[string]struct{}{}
			for _, value := range valuesByID {
				valueMap[value] = struct{}{}
			}

			if len(valueMap) != testCase.expectedNumNodes {
				t.Fatalf("unexpected number of nodes. want=%d have=%d", testCase.expectedNumNodes, len(valueMap))
			}

			for _, value := range values {
				if _, ok := valueMap[value]; !ok {
					t.Errorf("failed to find %q in reconstructed value set", value)
				}
			}
		})
	}
}

func readTrieTestInput(t *testing.T, name string) []string {
	contents, err := os.ReadFile(fmt.Sprintf("./testdata/%s.txt", name))
	if err != nil {
		t.Fatalf("failed to read test data: %s", err)
	}

	return strings.Split(strings.TrimSpace(string(contents)), "\n")
}
