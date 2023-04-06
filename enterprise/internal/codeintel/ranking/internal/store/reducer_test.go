package store

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertUploads(t, db, uploadsshared.Upload{ID: 1})

	// Insert definitions
	mockDefinitions := make(chan shared.RankingDefinitions, 3)
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:     1,
		SymbolName:   "foo",
		DocumentPath: "foo.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:     1,
		SymbolName:   "bar",
		DocumentPath: "bar.go",
	}
	mockDefinitions <- shared.RankingDefinitions{
		UploadID:     1,
		SymbolName:   "foo",
		DocumentPath: "foo.go",
	}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Insert references
	mockReferences := make(chan string, 3)
	mockReferences <- "foo"
	mockReferences <- "bar"
	mockReferences <- "baz"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 1, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs
	if _, _, err := store.InsertPathCountInputs(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123), 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	// Insert repos
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO repo (id, name) VALUES (1, 'deadbeef')`)); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	// Finally! Test InsertPathRanks
	numPathRanksInserted, numInputsProcessed, err := store.InsertPathRanks(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123), 10)
	if err != nil {
		t.Fatalf("unexpected error inserting path ranks: %s", err)
	}

	if numPathRanksInserted != 2 {
		t.Errorf("unexpected number of path ranks inserted. want=%d have=%d", 2, numPathRanksInserted)
	}

	if numInputsProcessed != 1 {
		t.Errorf("unexpected number of inputs processed. want=%d have=%d", 1, numInputsProcessed)
	}
}

func TestVacuumStaleRanks(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (name) VALUES ('bar'), ('baz'), ('bonk'), ('foo1'), ('foo2'), ('foo3'), ('foo4'), ('foo5')`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	for r, key := range map[string]string{
		"foo1": rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123),
		"foo2": rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123),
		"foo3": rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123),
		"foo4": rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123),
		"foo5": rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123),
		"bar":  rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 234),
		"baz":  rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 345),
		"bonk": rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 456),
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
	_, rankRecordsDeleted, err := store.VacuumStaleRanks(ctx, rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 456))
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

func setDocumentRanks(ctx context.Context, db *basestore.Store, repoName api.RepoName, ranks map[string]float64, graphKey string) error {
	serialized, err := json.Marshal(ranks)
	if err != nil {
		return err
	}

	return db.Exec(ctx, sqlf.Sprintf(setDocumentRanksQuery, repoName, serialized, graphKey))
}

const setDocumentRanksQuery = `
INSERT INTO codeintel_path_ranks AS pr (repository_id, payload, graph_key)
VALUES ((SELECT id FROM repo WHERE name = %s), %s, %s)
ON CONFLICT (repository_id) DO
UPDATE
	SET payload = EXCLUDED.payload
`
