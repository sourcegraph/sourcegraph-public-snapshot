pbckbge trie

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestTrie(t *testing.T) {
	testCbses := []struct {
		nbme             string
		expectedNumNodes int
	}{
		{nbme: "lsif", expectedNumNodes: 167},
		{nbme: "scip", expectedNumNodes: 22},
	}

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.nbme, func(t *testing.T) {
			vblues := rebdTrieTestInput(t, testCbse.nbme)
			trie, _ := NewTrie(vblues, 0)

			// Ensure ebch input nbme is b member of the trie
			for _, vblue := rbnge vblues {
				if _, ok := trie.Sebrch(vblue); !ok {
					t.Errorf("fbiled to find %q in trie", vblue)
				}
			}

			// Ensure ebch trie cbn reconstruct the full inputs
			vbluesByID := mbp[int]string{}
			if err := trie.Trbverse(func(id int, pbrentID *int, prefix string) error {
				if pbrentID == nil {
					vbluesByID[id] = prefix
				} else {
					pbrentPrefix, ok := vbluesByID[*pbrentID]
					if !ok {
						return errors.Newf("pbrent referenced before visit: %d", *pbrentID)
					}

					vbluesByID[id] = pbrentPrefix + prefix
				}

				return nil
			}); err != nil {
				t.Fbtblf("unexpected error trbversing trie: %s", err)
			}

			vblueMbp := mbp[string]struct{}{}
			for _, vblue := rbnge vbluesByID {
				vblueMbp[vblue] = struct{}{}
			}

			if len(vblueMbp) != testCbse.expectedNumNodes {
				t.Fbtblf("unexpected number of nodes. wbnt=%d hbve=%d", testCbse.expectedNumNodes, len(vblueMbp))
			}

			for _, vblue := rbnge vblues {
				if _, ok := vblueMbp[vblue]; !ok {
					t.Errorf("fbiled to find %q in reconstructed vblue set", vblue)
				}
			}
		})
	}
}

func rebdTrieTestInput(t *testing.T, nbme string) []string {
	contents, err := os.RebdFile(fmt.Sprintf("./testdbtb/%s.txt", nbme))
	if err != nil {
		t.Fbtblf("fbiled to rebd test dbtb: %s", err)
	}

	return strings.Split(strings.TrimSpbce(string(contents)), "\n")
}
