package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertInitialPathRanks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	mockUploadID := 1
	mockPathNames := make(chan string, 3)
	mockPathNames <- "foo.go"
	mockPathNames <- "bar.go"
	mockPathNames <- "baz.go"
	close(mockPathNames)
	if err := store.InsertInitialPathRanks(ctx, mockUploadID, mockPathNames, 2, mockRankingGraphKey); err != nil {
		t.Fatalf("unexpected error inserting initial path counts: %s", err)
	}

	inputs, err := getInitialPathRanks(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting path count inputs: %s", err)
	}

	expectedInputs := []initialPathRanks{
		{UploadID: 1, DocumentPath: "bar.go"},
		{UploadID: 1, DocumentPath: "baz.go"},
		{UploadID: 1, DocumentPath: "foo.go"},
	}
	if diff := cmp.Diff(expectedInputs, inputs); diff != "" {
		t.Errorf("unexpected path count inputs (-want +got):\n%s", diff)
	}
}

func TestVacuumAbandonedInitialPathCounts(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_initial_path_ranks (upload_id, document_paths, graph_key)
		SELECT 50, '{"test"}', $1 FROM generate_series(1, 30)
	`, mockRankingGraphKey); err != nil {
		t.Fatalf("failed to insert ranking path count inputs: %s", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_initial_path_ranks (upload_id, document_paths, graph_key)
		SELECT 50, '{"test"}', $1 FROM generate_series(1, 60)
	`, mockRankingGraphKey+"-abandoned"); err != nil {
		t.Fatalf("failed to insert ranking path count inputs: %s", err)
	}

	assertCounts := func(expectedInitialPathCountRecords int) {
		store := basestore.NewWithHandle(db.Handle())

		numInitialPathCountRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_initial_path_ranks`)))
		if err != nil {
			t.Fatalf("failed to definition records: %s", err)
		}
		if expectedInitialPathCountRecords != numInitialPathCountRecords {
			t.Fatalf("unexpected number of initial path counts records. want=%d have=%d", expectedInitialPathCountRecords, numInitialPathCountRecords)
		}
	}

	// assert initial count
	assertCounts(3 * 30)

	// remove records associated with other ranking keys
	if _, err := store.VacuumAbandonedInitialPathCounts(ctx, mockRankingGraphKey, 50); err != nil {
		t.Fatalf("unexpected error vacuuming initial path counts: %s", err)
	}

	// only 10 records of stale derivative graph key remain (batch size of 50, but 2*30 could be deleted)
	assertCounts(1*30 + 10)
}

func TestSoftDeleteStaleInitialPaths(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	for _, uploadID := range []int{1, 2, 3} {
		insertUploads(t, db, uploadsshared.Upload{ID: uploadID})

		mockPathNames := make(chan string, 3)
		mockPathNames <- "foo.go"
		mockPathNames <- "bar.go"
		mockPathNames <- "baz.go"
		close(mockPathNames)
		if err := store.InsertInitialPathRanks(ctx, uploadID, mockPathNames, 2, mockRankingGraphKey); err != nil {
			t.Errorf("unexpected error vacuuming initial paths: %s", err)
		}
	}

	assertCounts := func(expectedNumRecords int) {
		initialRanks, err := getInitialPathRanks(ctx, t, db, mockRankingGraphKey)
		if err != nil {
			t.Fatalf("failed to get initial ranks: %s", err)
		}
		if len(initialRanks) != expectedNumRecords {
			t.Errorf("unexpected number of initial ranks. want=%d have=%d", expectedNumRecords, len(initialRanks))
		}
	}

	// assert initial count
	assertCounts(9)

	// make upload 2 visible at tip (1 and 3 are not)
	insertVisibleAtTip(t, db, 50, 2)

	// remove path counts for non-visible uploads
	if _, _, err := store.SoftDeleteStaleInitialPaths(ctx, mockRankingGraphKey); err != nil {
		t.Fatalf("unexpected error vacuuming stale initial counts: %s", err)
	}

	// only upload 2's entries remain
	assertCounts(3)
}

func TestVacuumDeletedInitialPaths(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123)

	// TODO - setup

	_, err := store.VacuumDeletedInitialPaths(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error vacuuming deleted initial paths: %s", err)
	}

	// TODO - assertions
}

//
//

type initialPathRanks struct {
	UploadID     int
	DocumentPath string
}

func getInitialPathRanks(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (pathRanks []initialPathRanks, err error) {
	query := sqlf.Sprintf(`
		SELECT upload_id, document_path FROM (
			SELECT
				upload_id,
				unnest(document_paths) AS document_path
			FROM codeintel_initial_path_ranks
			WHERE graph_key LIKE %s || '%%' AND deleted_at IS NULL
		)s
		GROUP BY upload_id, document_path
		ORDER BY upload_id, document_path
	`, graphKey)
	rows, err := db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var input initialPathRanks
		if err := rows.Scan(&input.UploadID, &input.DocumentPath); err != nil {
			return nil, err
		}

		pathRanks = append(pathRanks, input)
	}

	return pathRanks, nil
}
