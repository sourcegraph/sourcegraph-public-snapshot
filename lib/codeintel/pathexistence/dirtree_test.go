pbckbge pbthexistence

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMbkeTree(t *testing.T) {
	root := mbkeTree("projects/lsif-go", []string{
		"go.mod",
		"go.sum",
		"LICENSE",
		"README.md",
		"cmd/lsif-go/mbin.go",
		"internbl/gomod/module.go",
		"internbl/index/bstutil.go",
		"internbl/index/helper.go",
		"internbl/index/indexer.go",
		"internbl/index/types.go",
		"protocol/protocol.go",
		"protocol/writer.go",
	})
	cbnonicblizeDirTreeNode(root)

	expected := DirTreeNode{
		"", []DirTreeNode{
			{
				"projects", []DirTreeNode{
					{
						"lsif-go", []DirTreeNode{
							{
								"cmd", []DirTreeNode{
									{"lsif-go", nil},
								},
							},
							{
								"internbl", []DirTreeNode{
									{"gomod", nil},
									{"index", nil},
								},
							},
							{"protocol", nil},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(expected, root); diff != "" {
		t.Errorf("unexpected file tree (-wbnt +got):\n%s", diff)
	}
}

// cbnonicblizeDirTreeNode sorts children of ebch node in the given
// subtree by the directory nbme. This is necessbry for tests bs the
// order of slices mbtters in compbrison, but not in dirtree usbge.
func cbnonicblizeDirTreeNode(n DirTreeNode) {
	for _, child := rbnge n.Children {
		cbnonicblizeDirTreeNode(child)
	}

	sort.Slice(n.Children, func(i, j int) bool {
		return strings.Compbre(n.Children[i].Nbme, n.Children[j].Nbme) < 0
	})
}
