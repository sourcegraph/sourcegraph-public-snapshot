package vfsutil

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/ctxvfs"
)

func testVFS(t *testing.T, fs ctxvfs.FileSystem, want map[string]string) {
	tree, err := ctxvfs.ReadAllFiles(context.Background(), fs, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != len(want) {
		t.Errorf("got %d files, want %d files", len(tree), len(want))
	}
	for wantFile, wantContents := range want {
		contentsBytes, ok := tree[wantFile]
		if !ok {
			t.Errorf("missing file %s", wantFile)
			continue
		}
		contents := string(contentsBytes)

		if strings.HasSuffix(wantContents, "...") {
			// Allow specifying expected contents with "..." at the
			// end for ease of test creation.
			if len(contents) >= len(wantContents)+3 {
				contents = contents[:len(wantContents)-3] + "..."
			}
		}
		if contents != wantContents {
			t.Errorf("%s: got contents %q, want %q", wantFile, contents, wantContents)
		}
	}
	for file := range tree {
		if _, ok := want[file]; !ok {
			t.Errorf("extra file %s", file)
		}
	}
}
