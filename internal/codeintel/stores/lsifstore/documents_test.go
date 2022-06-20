package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDocumentPaths(t *testing.T) {
	store := populateTestStore(t)

	if paths, count, err := store.DocumentPaths(context.Background(), testBundleID, "%%"); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if count != 7 || len(paths) != 7 {
		t.Errorf("expected %d document paths but got none: count=%d len=%d", 7, count, len(paths))
	} else {
		expected := []string{
			"cmd/lsif-go/main.go",
			"internal/gomod/module.go",
			"internal/index/helper.go",
			"internal/index/indexer.go",
			"internal/index/types.go",
			"protocol/protocol.go",
			"protocol/writer.go",
		}

		if diff := cmp.Diff(expected, paths); diff != "" {
			t.Errorf("unexpected document paths (-want +got):\n%s", diff)
		}
	}
}
