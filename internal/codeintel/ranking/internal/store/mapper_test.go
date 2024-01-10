package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertPathCountInputs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123")

	// Insert and export uploads
	insertUploads(t, db,
		uploadsshared.Upload{ID: 42, RepositoryID: 50},
		uploadsshared.Upload{ID: 43, RepositoryID: 51},
		uploadsshared.Upload{ID: 90, RepositoryID: 52},
		uploadsshared.Upload{ID: 91, RepositoryID: 53}, // older   (by ID order)
		uploadsshared.Upload{ID: 92, RepositoryID: 53}, // younger (by ID order)
		uploadsshared.Upload{ID: 93, RepositoryID: 54, Root: "lib/", Indexer: "test"},
		uploadsshared.Upload{ID: 94, RepositoryID: 54, Root: "lib/", Indexer: "test"},
	)
	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{42, 43, 90, 91, 92, 93, 94}'::integer[]))
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		SELECT id + 100, id, $1, (SELECT md5(u.repository_id::text || ':' || u.root) FROM lsif_uploads u WHERE u.id = v.id) FROM v AS v(id)
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("failed to insert exported upload: %s", err)
	}

	// Insert definitions
	mockDefinitions := make(chan shared.RankingDefinitions, 4)
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:         42,
		ExportedUploadID: 142,
		SymbolChecksum:   hash("foo"),
		DocumentPath:     "foo.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:         42,
		ExportedUploadID: 142,
		SymbolChecksum:   hash("bar"),
		DocumentPath:     "bar.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:         43,
		ExportedUploadID: 143,
		SymbolChecksum:   hash("baz"),
		DocumentPath:     "baz.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:         43,
		ExportedUploadID: 143,
		SymbolChecksum:   hash("bonk"),
		DocumentPath:     "bonk.go",
	}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Insert metadata to trigger mapper
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_progress(graph_key, max_export_id, mappers_started_at)
		VALUES ($1, 1000, NOW())
	`,
		key,
	); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}

	//
	// Basic test case

	mockReferences := make(chan [16]byte, 2)
	mockReferences <- hash("foo")
	mockReferences <- hash("bar")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 190, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	if _, _, err := store.InsertPathCountInputs(ctx, key, 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	// Same ID split over two batches
	mockReferences = make(chan [16]byte, 1)
	mockReferences <- hash("baz")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 190, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Duplicate of 92 (below) - should not be processed as it's older
	mockReferences = make(chan [16]byte, 4)
	mockReferences <- hash("foo")
	mockReferences <- hash("bar")
	mockReferences <- hash("baz")
	mockReferences <- hash("bonk")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 191, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	mockReferences = make(chan [16]byte, 1)
	mockReferences <- hash("bonk")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 192, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs: should process existing rows
	if _, _, err := store.InsertPathCountInputs(ctx, key, 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	//
	// Incremental insertion test case

	mockReferences = make(chan [16]byte, 2)
	mockReferences <- hash("foo")
	mockReferences <- hash("bar")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 193, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs: should process unprocessed rows only
	if _, _, err := store.InsertPathCountInputs(ctx, key, 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	//
	// No-op test case

	mockReferences = make(chan [16]byte, 4)
	mockReferences <- hash("foo")
	mockReferences <- hash("bar")
	mockReferences <- hash("baz")
	mockReferences <- hash("bonk")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 194, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs: should do nothing, 94 covers the same project as 93
	if _, _, err := store.InsertPathCountInputs(ctx, key, 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	//
	// Assertions

	// Test path count inputs were inserted
	inputs, err := getPathCountsInputs(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting path count inputs: %s", err)
	}

	expectedInputs := []pathCountsInput{
		{RepositoryID: 50, DocumentPath: "bar.go", Count: 2},
		{RepositoryID: 50, DocumentPath: "foo.go", Count: 2},
		{RepositoryID: 51, DocumentPath: "baz.go", Count: 1},
		{RepositoryID: 51, DocumentPath: "bonk.go", Count: 1},
	}
	if diff := cmp.Diff(expectedInputs, inputs); diff != "" {
		t.Errorf("unexpected path count inputs (-want +got):\n%s", diff)
	}
}

func TestInsertInitialPathCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123")

	// Insert and export upload
	// N.B. This creates repository 50 implicitly
	insertUploads(t, db, uploadsshared.Upload{ID: 4})
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		VALUES (104, 4, $1, md5('key-4'))
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("failed to insert exported upload: %s", err)
	}

	// Insert metadata to trigger mapper
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_progress(graph_key, max_export_id, mappers_started_at)
		VALUES ($1, 1000, NOW())
	`,
		key,
	); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}

	mockPathNames := []string{
		"foo.go",
		"bar.go",
		"baz.go",
	}
	if err := store.InsertInitialPathRanks(ctx, 104, mockPathNames, 2, mockRankingGraphKey); err != nil {
		t.Fatalf("unexpected error inserting initial path counts: %s", err)
	}

	if _, _, err := store.InsertInitialPathCounts(ctx, key, 1000); err != nil {
		t.Fatalf("unexpected error inserting initial path counts: %s", err)
	}

	inputs, err := getPathCountsInputs(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting path count inputs: %s", err)
	}

	expectedInputs := []pathCountsInput{
		{RepositoryID: 50, DocumentPath: "bar.go", Count: 0},
		{RepositoryID: 50, DocumentPath: "baz.go", Count: 0},
		{RepositoryID: 50, DocumentPath: "foo.go", Count: 0},
	}
	if diff := cmp.Diff(expectedInputs, inputs); diff != "" {
		t.Errorf("unexpected path count inputs (-want +got):\n%s", diff)
	}
}

func TestVacuumStaleProcessedReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123")

	// Insert and export uploads
	insertUploads(t, db,
		uploadsshared.Upload{ID: 1},
		uploadsshared.Upload{ID: 2},
		uploadsshared.Upload{ID: 3},
	)
	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		SELECT id + 100, id, $1, md5('key-' || id::text) FROM v AS v(id)
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("failed to insert exported upload: %s", err)
	}

	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_ranking_references (graph_key, symbol_names, symbol_checksums, exported_upload_id)
		SELECT $1, '{}', '{}', id FROM codeintel_ranking_exports
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("failed to insert references: %s", err)
	}

	for _, graphKey := range []string{
		key,
		rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "456"),
		rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "789"),
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_ranking_references_processed (graph_key, codeintel_ranking_reference_id)
			SELECT $1, id FROM codeintel_ranking_references
		`, graphKey); err != nil {
			t.Fatalf("failed to insert ranking processed reference records: %s", err)
		}
	}

	if numRecords, _, err := basestore.ScanFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_ranking_references_processed")); err != nil {
		t.Fatalf("unexpected error counting records: %s", err)
	} else if numRecords != 9 {
		t.Fatalf("unexpected number of records. want=%d have=%d", 9, numRecords)
	}

	if _, err := store.VacuumStaleProcessedReferences(ctx, key, 1000); err != nil {
		t.Fatalf("unexpected error vacuuming processed reference records: %s", err)
	}

	if numRecords, _, err := basestore.ScanFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_ranking_references_processed")); err != nil {
		t.Fatalf("unexpected error counting records: %s", err)
	} else if numRecords != 3 {
		t.Fatalf("unexpected number of records. want=%d have=%d", 3, numRecords)
	}
}

func TestVacuumStaleProcessedPaths(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123")

	// Insert and export uploads
	insertUploads(t, db,
		uploadsshared.Upload{ID: 1},
		uploadsshared.Upload{ID: 2},
		uploadsshared.Upload{ID: 3},
	)
	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		SELECT id + 100, id, $1, md5('key-' || id::text) FROM v AS v(id)
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("failed to insert exported upload: %s", err)
	}

	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_initial_path_ranks (graph_key, document_paths, exported_upload_id)
		SELECT $1, '{}', id FROM codeintel_ranking_exports
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("failed to insert path ranks: %s", err)
	}

	for _, graphKey := range []string{
		key,
		rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "456"),
		rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "789"),
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_initial_path_ranks_processed (graph_key, codeintel_initial_path_ranks_id)
			SELECT $1, id FROM codeintel_initial_path_ranks
		`, graphKey); err != nil {
			t.Fatalf("failed to insert ranking processed reference records: %s", err)
		}
	}

	if numRecords, _, err := basestore.ScanFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_initial_path_ranks_processed")); err != nil {
		t.Fatalf("unexpected error counting records: %s", err)
	} else if numRecords != 9 {
		t.Fatalf("unexpected number of records. want=%d have=%d", 9, numRecords)
	}

	if _, err := store.VacuumStaleProcessedPaths(ctx, key, 1000); err != nil {
		t.Fatalf("unexpected error vacuuming processed path records: %s", err)
	}

	if numRecords, _, err := basestore.ScanFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_initial_path_ranks_processed")); err != nil {
		t.Fatalf("unexpected error counting records: %s", err)
	} else if numRecords != 3 {
		t.Fatalf("unexpected number of records. want=%d have=%d", 3, numRecords)
	}
}

func TestVacuumStaleGraphs(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123")

	// Insert and export uploads
	insertUploads(t, db,
		uploadsshared.Upload{ID: 1},
		uploadsshared.Upload{ID: 2},
		uploadsshared.Upload{ID: 3},
	)
	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		SELECT id + 100, id, $1, md5('key-' || id::text) FROM v AS v(id)
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("failed to insert exported upload: %s", err)
	}

	mockReferences := make(chan [16]byte, 2)
	mockReferences <- hash("foo")
	mockReferences <- hash("bar")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 101, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	mockReferences = make(chan [16]byte, 3)
	mockReferences <- hash("foo")
	mockReferences <- hash("bar")
	mockReferences <- hash("baz")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 102, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	mockReferences = make(chan [16]byte, 2)
	mockReferences <- hash("bar")
	mockReferences <- hash("baz")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 103, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	for _, graphKey := range []string{
		key,
		rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "456"),
		rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "789"),
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_ranking_references_processed (graph_key, codeintel_ranking_reference_id)
			SELECT $1, id FROM codeintel_ranking_references
		`, graphKey); err != nil {
			t.Fatalf("failed to insert ranking references processed: %s", err)
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_ranking_path_counts_inputs (definition_id, count, graph_key)
			SELECT v, 100, $1 FROM generate_series(1, 30) AS v
		`, graphKey); err != nil {
			t.Fatalf("failed to insert ranking path count inputs: %s", err)
		}
	}

	assertCounts := func(expectedInputRecords int) {
		store := basestore.NewWithHandle(db.Handle())

		numInputRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_ranking_path_counts_inputs`)))
		if err != nil {
			t.Fatalf("failed to count input records: %s", err)
		}
		if expectedInputRecords != numInputRecords {
			t.Errorf("unexpected number of input records. want=%d have=%d", expectedInputRecords, numInputRecords)
		}
	}

	// assert initial count
	assertCounts(3 * 30)

	// remove records associated with other ranking keys
	if _, err := store.VacuumStaleGraphs(ctx, rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "456"), 50); err != nil {
		t.Fatalf("unexpected error vacuuming stale graphs: %s", err)
	}

	// only 10 records of stale derivative graph key remain (batch size of 50, but 2*30 could be deleted)
	assertCounts(1*30 + 10)
}

//
//

type pathCountsInput struct {
	RepositoryID int
	DocumentPath string
	Count        int
}

func getPathCountsInputs(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (_ []pathCountsInput, err error) {
	query := sqlf.Sprintf(`
		SELECT repository_id, document_path, SUM(count)
		FROM codeintel_ranking_path_counts_inputs pci
		JOIN codeintel_ranking_definitions rd ON rd.id = pci.definition_id
		JOIN codeintel_ranking_exports eu ON eu.id = rd.exported_upload_id
		JOIN lsif_uploads u ON u.id = eu.upload_id
		WHERE pci.graph_key LIKE %s || '%%'
		GROUP BY u.repository_id, rd.document_path
		ORDER BY u.repository_id, rd.document_path
	`, graphKey)
	rows, err := db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var pathCountsInputs []pathCountsInput
	for rows.Next() {
		var input pathCountsInput
		if err := rows.Scan(&input.RepositoryID, &input.DocumentPath, &input.Count); err != nil {
			return nil, err
		}

		pathCountsInputs = append(pathCountsInputs, input)
	}

	return pathCountsInputs, nil
}
