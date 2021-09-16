package migration

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestReferenceCountMigrator(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := dbstore.NewWithDB(db, &observation.TestContext)
	migrator := NewReferenceCountMigrator(store, 75)

	n := 150
	expectedNumReferences := make([]int, 0, n)
	for i := 0; i < n; i++ {
		expectedNumReferences = append(expectedNumReferences, n-i-1)
	}

	assertProgress := func(expectedProgress float64) {
		if progress, err := migrator.Progress(context.Background()); err != nil {
			t.Fatalf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. want=%.2f have=%.2f", expectedProgress, progress)
		}
	}

	assertNumReferences := func(expectedNumReferences []int) {
		query := sqlf.Sprintf(`SELECT u.num_references FROM lsif_uploads u WHERE u.num_references IS NOT NULL ORDER BY u.id`)

		if numReferences, err := basestore.ScanInts(store.Query(context.Background(), query)); err != nil {
			t.Fatalf("unexpected error querying uploads: %s", err)
		} else if diff := cmp.Diff(expectedNumReferences, numReferences); diff != "" {
			t.Errorf("unexpected num references (-want +got):\n%s", diff)
		}
	}

	if err := store.Exec(context.Background(), sqlf.Sprintf("INSERT INTO repo (id, name) VALUES (42, 'foo'), (43, 'bar')")); err != nil {
		t.Fatalf("unexpected error inserting repo: %s", err)
	}

	for i := 0; i < n; i++ {
		if err := store.Exec(context.Background(), sqlf.Sprintf(
			"INSERT INTO lsif_uploads (repository_id, commit, state, indexer, num_parts, uploaded_parts) VALUES (%s, %s, 'completed', 'lsif-go', 0, '{}')",
			42+i/(n/2), // 50% id=42, 50% id=43
			fmt.Sprintf("%040d", i),
		)); err != nil {
			t.Fatalf("unexpected error inserting upload: %s", err)
		}

		if err := store.Exec(context.Background(), sqlf.Sprintf(
			"INSERT INTO lsif_packages (scheme, name, version, dump_id) VALUES ('test', %s, '1.2.3', %s)",
			fmt.Sprintf("pkg-%03d", i),
			i+1,
		)); err != nil {
			t.Fatalf("unexpected error inserting upload: %s", err)
		}

		for j := i - 1; j >= 0; j-- {
			if err := store.Exec(context.Background(), sqlf.Sprintf(
				"INSERT INTO lsif_references (scheme, name, version, filter, dump_id) VALUES ('test', %s, '1.2.3', '', %s)",
				fmt.Sprintf("pkg-%03d", j),
				i+1,
			)); err != nil {
				t.Fatalf("unexpected error inserting upload: %s", err)
			}
		}
	}

	assertProgress(0)

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(0.5)
	assertNumReferences(expectedNumReferences[:n/2])

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(1)
	assertNumReferences(expectedNumReferences)

	if err := migrator.Down(context.Background()); err != nil {
		t.Fatalf("unexpected error performing down migration: %s", err)
	}
	assertProgress(0.5)

	if err := migrator.Down(context.Background()); err != nil {
		t.Fatalf("unexpected error performing down migration: %s", err)
	}
	assertProgress(0)
}
