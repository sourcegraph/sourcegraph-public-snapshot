package dbstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestReferencesForUpload(t *testing.T) {
	t.Skip()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(db)

	insertUploads(t, db,
		Upload{ID: 1, Commit: makeCommit(2), Root: "sub1/"},
		Upload{ID: 2, Commit: makeCommit(3), Root: "sub2/"},
		Upload{ID: 3, Commit: makeCommit(4), Root: "sub3/"},
		Upload{ID: 4, Commit: makeCommit(3), Root: "sub4/"},
		Upload{ID: 5, Commit: makeCommit(2), Root: "sub5/"},
	)

	// insertPackageReferences(t, store, []shared.PackageReference{
	// 	{Package: shared.Package{DumpID: 1, Scheme: "gomod", Name: "leftpad", Version: "1.1.0"}},
	// 	{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "2.1.0"}},
	// 	{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "3.1.0"}},
	// 	{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "4.1.0"}},
	// 	{Package: shared.Package{DumpID: 3, Scheme: "gomod", Name: "leftpad", Version: "5.1.0"}},
	// })

	scanner, err := store.ReferencesForUpload(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error getting filters: %s", err)
	}

	filters, err := consumeScanner(scanner)
	if err != nil {
		t.Fatalf("unexpected error from scanner: %s", err)
	}

	expected := []shared.PackageReference{
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "2.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "3.1.0"}},
		{Package: shared.Package{DumpID: 2, Scheme: "gomod", Name: "leftpad", Version: "4.1.0"}},
	}
	if diff := cmp.Diff(expected, filters); diff != "" {
		t.Errorf("unexpected filters (-want +got):\n%s", diff)
	}
}

// consumeScanner reads all values from the scanner into memory.
func consumeScanner(scanner PackageReferenceScanner) (references []shared.PackageReference, _ error) {
	for {
		reference, exists, err := scanner.Next()
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}

		references = append(references, reference)
	}
	if err := scanner.Close(); err != nil {
		return nil, err
	}

	return references, nil
}
