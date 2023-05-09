package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	t1 := time.Now().Add(-time.Minute * 10)
	t2 := time.Now().Add(-time.Minute * 20)

	insertUploads(t, db,
		uploadsshared.Upload{ID: 42, RepositoryID: 50},
		uploadsshared.Upload{ID: 43, RepositoryID: 51},
		uploadsshared.Upload{ID: 90, RepositoryID: 52},
		uploadsshared.Upload{ID: 91, RepositoryID: 53, FinishedAt: &t1}, // younger
		uploadsshared.Upload{ID: 92, RepositoryID: 53, FinishedAt: &t2}, // older
		uploadsshared.Upload{ID: 93, RepositoryID: 54, Root: "lib/", Indexer: "test"},
		uploadsshared.Upload{ID: 94, RepositoryID: 54, Root: "lib/", Indexer: "test"},
	)

	// Insert definitions
	mockDefinitions := make(chan shared.RankingDefinitions, 4)
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:     42,
		SymbolName:   "foo",
		DocumentPath: "foo.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:     42,
		SymbolName:   "bar",
		DocumentPath: "bar.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:     43,
		SymbolName:   "baz",
		DocumentPath: "baz.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:     43,
		SymbolName:   "bonk",
		DocumentPath: "bonk.go",
	}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Insert metadata to trigger mapper
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_progress(graph_key, max_definition_id, max_reference_id, max_path_id, mappers_started_at)
		VALUES ($1,  1000, 1000, 1000, NOW())
	`,
		rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123),
	); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}

	//
	// Basic test case

	mockReferences := make(chan string, 2)
	mockReferences <- "foo"
	mockReferences <- "bar"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 90, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	if _, _, err := store.InsertPathCountInputs(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123), 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	// Same ID split over two batches
	mockReferences = make(chan string, 1)
	mockReferences <- "baz"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 90, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	mockReferences = make(chan string, 1)
	mockReferences <- "bonk"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 91, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// duplicate of 91 (should not be processed as it's older)
	mockReferences = make(chan string, 4)
	mockReferences <- "foo"
	mockReferences <- "bar"
	mockReferences <- "baz"
	mockReferences <- "bonk"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 92, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs: should process existing rows
	if _, _, err := store.InsertPathCountInputs(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123), 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	//
	// Incremental insertion test case

	mockReferences = make(chan string, 2)
	mockReferences <- "foo"
	mockReferences <- "bar"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 93, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs: should process unprocessed rows only
	if _, _, err := store.InsertPathCountInputs(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123), 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	//
	// No-op test case

	mockReferences = make(chan string, 4)
	mockReferences <- "foo"
	mockReferences <- "bar"
	mockReferences <- "baz"
	mockReferences <- "bonk"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 94, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs: should do nothing, 94 covers the same project as 93
	if _, _, err := store.InsertPathCountInputs(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123), 1000); err != nil {
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// Creates repository 50
	insertUploads(t, db, uploadsshared.Upload{ID: 1})

	// Insert metadata to trigger mapper
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_progress(graph_key, max_definition_id, max_reference_id, max_path_id, mappers_started_at)
		VALUES ($1,  1000, 1000, 1000, NOW())
	`,
		rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123),
	); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}

	mockUploadID := 1
	mockPathNames := make(chan string, 3)
	mockPathNames <- "foo.go"
	mockPathNames <- "bar.go"
	mockPathNames <- "baz.go"
	close(mockPathNames)
	if err := store.InsertInitialPathRanks(ctx, mockUploadID, mockPathNames, 2, mockRankingGraphKey); err != nil {
		t.Fatalf("unexpected error inserting initial path counts: %s", err)
	}

	if _, _, err := store.InsertInitialPathCounts(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123), 1000); err != nil {
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

func TestVacuumStaleGraphs(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	mockReferences := make(chan string, 2)
	mockReferences <- "foo"
	mockReferences <- "bar"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 1, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	mockReferences = make(chan string, 3)
	mockReferences <- "foo"
	mockReferences <- "bar"
	mockReferences <- "baz"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 2, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	mockReferences = make(chan string, 2)
	mockReferences <- "bar"
	mockReferences <- "baz"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 3, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	for _, graphKey := range []string{
		rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123),
		rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 456),
		rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 789),
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_ranking_references_processed (graph_key, codeintel_ranking_reference_id)
			SELECT $1, id FROM codeintel_ranking_references
		`, graphKey); err != nil {
			t.Fatalf("failed to insert ranking references processed: %s", err)
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_ranking_path_counts_inputs (repository_id, document_path, count, graph_key)
			SELECT 50, '', 100, $1 FROM generate_series(1, 30)
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
	if _, err := store.VacuumStaleGraphs(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 456), 50); err != nil {
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
		FROM codeintel_ranking_path_counts_inputs
		WHERE graph_key LIKE %s || '%%'
		GROUP BY repository_id, document_path
		ORDER BY repository_id, document_path
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
