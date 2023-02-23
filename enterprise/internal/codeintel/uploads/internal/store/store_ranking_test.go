package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const (
	mockRankingGraphKey    = "mockDev"
	mockRankingBatchNumber = 10
)

func TestInsertReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// Insert references
	mockReferences := []shared.RankingReferences{
		{UploadID: 1, SymbolNames: []string{"foo", "bar", "baz"}},
	}

	for _, reference := range mockReferences {
		// Test InsertReferencesForRanking
		if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, reference); err != nil {
			t.Fatalf("unexpected error inserting references: %s", err)
		}
	}

	// Test references where inserted
	references, err := getRankingReferences(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting references: %s", err)
	}

	if diff := cmp.Diff(mockReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

func TestInsertDefinition(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// Insert defintions
	mockDefinitions := []shared.RankingDefintions{
		{
			UploadID:     1,
			SymbolName:   "foo",
			Repository:   "deadbeef",
			DocumentPath: "foo.go",
		},
		{
			UploadID:     1,
			SymbolName:   "bar",
			Repository:   "deadbeef",
			DocumentPath: "bar.go",
		},
		{
			UploadID:     1,
			SymbolName:   "foo",
			Repository:   "deadbeef",
			DocumentPath: "foo.go",
		},
	}

	// Test InsertDefinitionsForRanking
	if err := store.InsertDefintionsForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Test definitions where inserted
	definitions, err := getRankingDefinitions(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting definitions: %s", err)
	}

	if diff := cmp.Diff(mockDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

func TestInsertPathRanks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// Insert defintions
	mockDefinitions := []shared.RankingDefintions{
		{
			UploadID:     1,
			SymbolName:   "foo",
			Repository:   "deadbeef",
			DocumentPath: "foo.go",
		},
		{
			UploadID:     1,
			SymbolName:   "bar",
			Repository:   "deadbeef",
			DocumentPath: "bar.go",
		},
		{
			UploadID:     1,
			SymbolName:   "foo",
			Repository:   "deadbeef",
			DocumentPath: "foo.go",
		},
	}
	if err := store.InsertDefintionsForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Insert references
	mockReferences := shared.RankingReferences{
		UploadID: 1,
		SymbolNames: []string{
			mockDefinitions[0].SymbolName,
			mockDefinitions[1].SymbolName,
			mockDefinitions[2].SymbolName,
		},
	}
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs
	if err := store.InsertPathCountInputs(ctx, mockRankingGraphKey, 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	// Insert repos
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO repo (id, name) VALUES (1, 'deadbeef')`)); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	// Finally! Test InsertPathRanks
	numPathRanksInserted, numInputsProcessed, err := store.InsertPathRanks(ctx, mockRankingGraphKey, 10)
	if err != nil {
		t.Fatalf("unexpected error inserting path ranks: %s", err)
	}

	if numPathRanksInserted != 2 {
		t.Fatalf("unexpected number of path ranks inserted. want=%d have=%f", 2, numPathRanksInserted)
	}

	if numInputsProcessed != 1 {
		t.Fatalf("unexpected number of inputs processed. want=%d have=%f", 1, numInputsProcessed)
	}
}

func TestInsertPathCountInputs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// Insert defintions
	mockDefinitions := []shared.RankingDefintions{
		{
			UploadID:     1,
			SymbolName:   "foo",
			Repository:   "deadbeef",
			DocumentPath: "foo.go",
		},
		{
			UploadID:     1,
			SymbolName:   "bar",
			Repository:   "deadbeef",
			DocumentPath: "bar.go",
		},
		{
			UploadID:     1,
			SymbolName:   "foo",
			Repository:   "deadbeef",
			DocumentPath: "foo.go",
		},
	}
	if err := store.InsertDefintionsForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Insert references
	mockReferences := shared.RankingReferences{
		UploadID: 1,
		SymbolNames: []string{
			mockDefinitions[0].SymbolName,
			mockDefinitions[1].SymbolName,
			mockDefinitions[2].SymbolName,
		},
	}
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test InsertPathCountInputs
	if err := store.InsertPathCountInputs(ctx, mockRankingGraphKey, 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	// Test path count inputs where inserted
	repository, documentPath, count, err := getRankingPathCountsInputs(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting path count inputs: %s", err)
	}

	if repository != "deadbeef" {
		t.Fatalf("unexpected repository. want=%s have=%s", "deadbeef", repository)
	}

	if documentPath != "foo.go" {
		t.Fatalf("unexpected document path. want=%s have=%s", "foo.go", documentPath)
	}

	if count != 2 {
		t.Fatalf("unexpected count. want=%d have=%d", 2, count)
	}
}

func getRankingDefinitions(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) ([]shared.RankingDefintions, error) {
	query := fmt.Sprintf(
		`SELECT upload_id, symbol_name, repository, document_path FROM codeintel_ranking_definitions WHERE graph_key = '%s'`,
		graphKey,
	)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var definitions []shared.RankingDefintions
	for rows.Next() {
		var uploadID int
		var symbolName string
		var repository string
		var documentPath string
		err = rows.Scan(&uploadID, &symbolName, &repository, &documentPath)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, shared.RankingDefintions{
			UploadID:     uploadID,
			SymbolName:   symbolName,
			Repository:   repository,
			DocumentPath: documentPath,
		})
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	return definitions, nil
}

func getRankingReferences(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) ([]shared.RankingReferences, error) {
	query := fmt.Sprintf(
		`SELECT upload_id, symbol_names FROM codeintel_ranking_references WHERE graph_key = '%s'`,
		graphKey,
	)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var references []shared.RankingReferences
	for rows.Next() {
		var uploadID int
		var symbolNames []string
		err = rows.Scan(&uploadID, pq.Array(&symbolNames))
		if err != nil {
			return nil, err
		}
		references = append(references, shared.RankingReferences{
			UploadID:    uploadID,
			SymbolNames: symbolNames,
		})
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	return references, nil
}

func getRankingPathCountsInputs(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (repository, documentPath string, count int, err error) {
	query := fmt.Sprintf(
		`SELECT repository, document_path, count FROM codeintel_ranking_path_counts_inputs WHERE graph_key = '%s'`,
		graphKey,
	)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", "", 0, err
	}

	for rows.Next() {
		err = rows.Scan(&repository, &documentPath, &count)
		if err != nil {
			return "", "", 0, err
		}
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	return repository, documentPath, count, nil
}
