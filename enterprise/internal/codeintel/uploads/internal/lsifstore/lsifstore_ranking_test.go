package lsifstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	lsifstore := New(&observation.TestContext, codeIntelDB)

	mockRankingGraphKey := "mockDev"
	mockRankingBatchNumber := 10

	// Insert repos
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO repo (id, name) VALUES (1, 'deadbeef'), (2, 'alivebeef')`)); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

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
			UploadID:     2,
			SymbolName:   "foo",
			Repository:   "deadbeef",
			DocumentPath: "foo.go",
		},
	}

	// Test InsertDefinitionsForRanking
	if err := lsifstore.InsertDefintionsForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Insert references
	mockReferences := []shared.RankingReferences{
		{
			UploadID: 1,
			SymbolNames: []string{
				mockDefinitions[0].SymbolName,
				mockDefinitions[1].SymbolName,
			},
		},
		{
			UploadID: 2,
			SymbolNames: []string{
				mockDefinitions[2].SymbolName,
			},
		},
	}

	// Test InsertReferencesForRanking
	for _, ref := range mockReferences {
		if err := lsifstore.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, ref); err != nil {
			t.Fatalf("unexpected error inserting references: %s", err)
		}
	}

	getRankingReferencesQuery := fmt.Sprintf(
		`SELECT upload_id, symbol_names FROM codeintel_ranking_references WHERE graph_key = '%s'`,
		mockRankingGraphKey,
	)
	rows, err := codeIntelDB.QueryContext(ctx, getRankingReferencesQuery)
	if err != nil {
		t.Fatalf("failed to query ranking references: %s", err)
	}
	var uploadID int
	var symbolNames []string
	for rows.Next() {
		err = rows.Scan(&uploadID, pq.Array(&symbolNames))
		if err != nil {
			t.Fatalf("failed to scan row: %s", err)
		}
	}

	_ = uploadID
	_ = symbolNames

	// Test InsertPathCountInputs
	if err := lsifstore.InsertPathCountInputs(ctx, mockRankingGraphKey, 10); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}
	query := fmt.Sprintf(
		`SELECT id, repository FROM codeintel_ranking_path_counts_inputs WHERE graph_key = '%s'`,
		mockRankingGraphKey,
	)
	rows, err = codeIntelDB.QueryContext(ctx, query)
	if err != nil {
		t.Fatalf("failed to query path ranks: %s", err)
	}

	var id int
	var repository string
	for rows.Next() {
		err = rows.Scan(&id, &repository)
		if err != nil {
			t.Fatalf("failed to scan row: %s", err)
		}
	}

	_ = id
	_ = repository

	// Finally! Test InsertPathRanks
	numPathRanksInserted, numInputsProcessed, err := lsifstore.InsertPathRanks(ctx, mockRankingGraphKey, 10)
	if err != nil {
		t.Fatalf("unexpected error inserting path ranks: %s", err)
	}

	_ = numPathRanksInserted
	_ = numInputsProcessed
}
