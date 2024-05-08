package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertPathRanks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(observation.TestContextTB(t), db)

	key := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123")

	// Insert and export upload
	insertUploads(t, db, uploadsshared.Upload{ID: 4})
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		VALUES (104, 4, $1, md5('key-4'))
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("failed to insert exported upload: %s", err)
	}

	// Insert definitions
	mockDefinitions := make(chan shared.RankingDefinitions, 3)
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:         4,
		ExportedUploadID: 104,
		SymbolChecksum:   hash("foo"),
		DocumentPath:     "foo.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:         4,
		ExportedUploadID: 104,
		SymbolChecksum:   hash("bar"),
		DocumentPath:     "bar.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:         4,
		ExportedUploadID: 104,
		SymbolChecksum:   hash("foo"),
		DocumentPath:     "foo.go",
	}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Insert references
	mockReferences := make(chan [16]byte, 3)
	mockReferences <- hash("foo")
	mockReferences <- hash("bar")
	mockReferences <- hash("baz")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 104, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
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

	// Test InsertPathCountInputs
	if _, _, err := store.InsertPathCountInputs(ctx, key, 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	// Insert repos
	if _, err := db.ExecContext(ctx, `INSERT INTO repo (id, name) VALUES (1, 'deadbeef')`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	// Update metadata to trigger reducer
	if _, err := db.ExecContext(ctx, `UPDATE codeintel_ranking_progress SET reducer_started_at = NOW()`); err != nil {
		t.Fatalf("failed to update metadata: %s", err)
	}

	// Finally! Test InsertPathRanks
	if _, numInputsProcessed, err := store.InsertPathRanks(ctx, key, 10); err != nil {
		t.Fatalf("unexpected error inserting path ranks: %s", err)
	} else if numInputsProcessed != 1 {
		t.Errorf("unexpected number of inputs processed. want=%d have=%d", 1, numInputsProcessed)
	}

	// Need to run this again prior to checking document ranks as we have to close out
	// the progress record by processing *no* records after the last batch.

	if _, numInputsProcessed, err := store.InsertPathRanks(ctx, key, 10); err != nil {
		t.Fatalf("unexpected error inserting path ranks: %s", err)
	} else if numInputsProcessed != 0 {
		t.Fatalf("expected no more work to be available")
	}

	// Check actual ranks
	ranks, _, err := store.GetDocumentRanks(ctx, api.RepoName("n-50"))
	if err != nil {
		t.Fatalf("unexpected error getting document ranks")
	}

	expectedRanks := map[string]float64{
		"foo.go": 2,
		"bar.go": 1,
	}
	if diff := cmp.Diff(expectedRanks, ranks); diff != "" {
		t.Errorf("unexpected ranks (-want +got):\n%s", diff)
	}
}

func TestVacuumStaleRanks(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(observation.TestContextTB(t), db)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (name) VALUES ('bar'), ('baz'), ('bonk'), ('foo1'), ('foo2'), ('foo3'), ('foo4'), ('foo5')`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	key1 := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "123")
	key2 := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "234")
	key3 := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "345")
	key4 := rankingshared.NewDerivativeGraphKey(mockRankingGraphKey, "456")

	// Insert metadata to rank progress by completion date
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_progress(graph_key, max_export_id, mappers_started_at, reducer_completed_at)
		VALUES
			($1, 1000, NOW() - '80 second'::interval, NOW() - '70 second'::interval),
			($2, 1000, NOW() - '60 second'::interval, NOW() - '50 second'::interval),
			($3, 1000, NOW() - '40 second'::interval, NOW() - '30 second'::interval),
			($4, 1000, NOW() - '20 second'::interval, NULL)
	`,
		key1, key2, key3, key4,
	); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}

	for r, key := range map[string]string{
		"foo1": key1,
		"foo2": key1,
		"foo3": key1,
		"foo4": key1,
		"foo5": key1,
		"bar":  key2,
		"baz":  key3,
		"bonk": key4,
	} {
		if err := setDocumentRanks(ctx, basestore.NewWithHandle(db.Handle()), api.RepoName(r), nil, key); err != nil {
			t.Fatalf("failed to insert document ranks: %s", err)
		}
	}

	assertNames := func(expectedNames []string) {
		store := basestore.NewWithHandle(db.Handle())

		names, err := basestore.ScanStrings(store.Query(ctx, sqlf.Sprintf(`
			SELECT r.name
			FROM repo r
			JOIN codeintel_path_ranks pr ON pr.repository_id = r.id
			ORDER BY r.name
		`)))
		if err != nil {
			t.Fatalf("failed to fetch names: %s", err)
		}

		if diff := cmp.Diff(expectedNames, names); diff != "" {
			t.Errorf("unexpected names (-want +got):\n%s", diff)
		}
	}

	// assert initial names
	assertNames([]string{"bar", "baz", "bonk", "foo1", "foo2", "foo3", "foo4", "foo5"})

	// remove sufficiently stale records associated with other ranking keys
	_, rankRecordsDeleted, err := store.VacuumStaleRanks(ctx, key4)
	if err != nil {
		t.Fatalf("unexpected error vacuuming stale ranks: %s", err)
	}
	if expected := 6; rankRecordsDeleted != expected {
		t.Errorf("unexpected number of rank records deleted. want=%d have=%d", expected, rankRecordsDeleted)
	}

	// stale graph keys have been removed
	assertNames([]string{"baz", "bonk"})
}

//
//

func setDocumentRanks(ctx context.Context, db *basestore.Store, repoName api.RepoName, ranks map[string]float64, derivativeGraphKey string) error {
	serialized, err := json.Marshal(ranks)
	if err != nil {
		return err
	}

	return db.Exec(ctx, sqlf.Sprintf(setDocumentRanksQuery, derivativeGraphKey, repoName, serialized))
}

const setDocumentRanksQuery = `
INSERT INTO codeintel_path_ranks AS pr (graph_key, repository_id, payload)
VALUES (%s, (SELECT id FROM repo WHERE name = %s), %s)
ON CONFLICT (graph_key, repository_id) DO
UPDATE SET payload = EXCLUDED.payload
`
