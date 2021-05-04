package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDatabaseHover(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	// `\tcontents, err := findContents(pkgs, p, f, obj)`
	//                     ^^^^^^^^^^^^

	if actualText, actualRange, exists, err := store.Hover(context.Background(), testBundleID, "internal/index/indexer.go", 628, 20); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if !exists {
		t.Errorf("no hover found")
	} else {
		docstring := "findContents returns contents used as hover info for given object."
		signature := "func findContents(pkgs []*Package, p *Package, f *File, obj Object) ([]MarkedString, error)"
		expectedText := "```go\n" + signature + "\n```\n\n---\n\n" + docstring
		expectedRange := newRange(628, 18, 628, 30)

		if actualText != expectedText {
			t.Errorf("unexpected hover text. want=%s have=%s", expectedText, actualText)
		}

		if diff := cmp.Diff(expectedRange, actualRange); diff != "" {
			t.Errorf("unexpected hover range (-want +got):\n%s", diff)
		}
	}
}
