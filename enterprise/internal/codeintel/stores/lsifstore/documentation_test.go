package lsifstore

import (
	"context"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Note: You can `go test ./pkg -update` to update the expected `want` values in these tests.
// See https://github.com/hexops/autogold for more information.

func TestDocumentationPage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store := populateTestStore(t)

	got, err := store.DocumentationPage(context.Background(), testBundleID, "/")
	if err != nil {
		t.Fatal(err)
	}

	// Just a snapshot of the output so we know when anything changes.
	autogold.Equal(t, got)
}

func TestDocumentationPathInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store := populateTestStore(t)

	testCases := []struct {
		pathID string
		want   autogold.Value
	}{
		{"/github.com/sourcegraph/lsif-go", autogold.Want("/github.com/sourcegraph/lsif-go", &precise.DocumentationPathInfoData{
			PathID:  "/github.com/sourcegraph/lsif-go",
			IsIndex: true,
			Children: []string{
				"/github.com/sourcegraph/lsif-go/cmd",
				"/github.com/sourcegraph/lsif-go/internal",
			},
		})},
		{"/github.com/sourcegraph/lsif-go/cmd", autogold.Want("/github.com/sourcegraph/lsif-go/cmd", &precise.DocumentationPathInfoData{
			PathID:  "/github.com/sourcegraph/lsif-go/cmd",
			IsIndex: true,
			Children: []string{
				"/github.com/sourcegraph/lsif-go/cmd/lsif-go",
			},
		})},
		{"/github.com/sourcegraph/lsif-go/internal", autogold.Want("/github.com/sourcegraph/lsif-go/internal", &precise.DocumentationPathInfoData{
			PathID:  "/github.com/sourcegraph/lsif-go/internal",
			IsIndex: true,
			Children: []string{
				"/github.com/sourcegraph/lsif-go/internal/gomod",
				"/github.com/sourcegraph/lsif-go/internal/index",
			},
		})},
		{"/github.com/sourcegraph/lsif-go/internal/index", autogold.Want("/github.com/sourcegraph/lsif-go/internal/index", &precise.DocumentationPathInfoData{
			PathID:   "/github.com/sourcegraph/lsif-go/internal/index",
			Children: []string{},
		})},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got, err := store.DocumentationPathInfo(context.Background(), testBundleID, tc.pathID)
			if err != nil {
				t.Fatal(err)
			}
			tc.want.Equal(t, got)
		})
	}
}

func TestDocumentationDefinitions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	store := populateTestStore(t)

	testCases := []autogold.Value{
		autogold.Want("/github.com/sourcegraph/lsif-go/internal/index#NewIndexer", []Location{{
			DumpID: 1,
			Path:   "internal/index/indexer.go",
			Range: Range{
				Start: Position{
					Line:      62,
					Character: 5,
				},
				End: Position{
					Line:      62,
					Character: 15,
				},
			},
		}}),
		autogold.Want("/github.com/sourcegraph/lsif-go/internal/gomod#versionPattern", []Location{{
			DumpID: 1,
			Path:   "internal/gomod/module.go",
			Range: Range{
				Start: Position{
					Line:      84,
					Character: 4,
				},
				End: Position{
					Line:      84,
					Character: 18,
				},
			},
		}}),
		autogold.Want("/github.com/sourcegraph/lsif-go/cmd/lsif-go#main", []Location{{
			DumpID: 1,
			Path:   "cmd/lsif-go/main.go",
			Range: Range{
				Start: Position{
					Line:      26,
					Character: 5,
				},
				End: Position{
					Line:      26,
					Character: 9,
				},
			},
		}}),
	}
	for _, want := range testCases {
		t.Run(want.Name(), func(t *testing.T) {
			pathID := want.Name()
			got, _, err := store.DocumentationDefinitions(context.Background(), testBundleID, pathID, 10, 0)
			if err != nil {
				t.Fatal(err)
			}
			want.Equal(t, got)
		})
	}
}
