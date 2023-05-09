package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// Insert references
	mockReferences := make(chan string, 3)
	mockReferences <- "foo"
	mockReferences <- "bar"
	mockReferences <- "baz"
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 1, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test references were inserted
	references, err := getRankingReferences(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting references: %s", err)
	}

	expectedReferences := []shared.RankingReferences{
		{
			UploadID:    1,
			SymbolNames: []string{"foo", "bar", "baz"},
		},
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

func TestVacuumAbandonedReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	mockReferences1 := make(chan string, 30)
	mockReferences2 := make(chan string, 30)
	mockReferences3 := make(chan string, 30)
	for j := 0; j < 30; j++ {
		mockReferences1 <- fmt.Sprintf("s%d", j+1)
		mockReferences2 <- fmt.Sprintf("s%d", j+1)
		mockReferences3 <- fmt.Sprintf("s%d", j+1)
	}
	close(mockReferences1)
	close(mockReferences2)
	close(mockReferences3)

	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey+"-abandoned", 1, 1, mockReferences1); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, 1, 1, mockReferences2); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey+"-abandoned", 1, 1, mockReferences3); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	assertCounts := func(expectedReferenceRecords int) {
		store := basestore.NewWithHandle(db.Handle())

		numReferenceRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`
			WITH symbols AS (
				SELECT unnest(symbol_names) AS symbol_name
				FROM codeintel_ranking_references
			)
			SELECT COUNT(*) FROM symbols
		`)))
		if err != nil {
			t.Fatalf("failed to definition records: %s", err)
		}
		if expectedReferenceRecords != numReferenceRecords {
			t.Fatalf("unexpected number of references records. want=%d have=%d", expectedReferenceRecords, numReferenceRecords)
		}
	}

	// assert initial count
	assertCounts(3 * 30)

	// remove records associated with other ranking keys
	if _, err := store.VacuumAbandonedReferences(ctx, mockRankingGraphKey, 50); err != nil {
		t.Fatalf("unexpected error vacuuming abandoned references: %s", err)
	}

	// only 10 records of stale graph key remain (batch size of 50, but 2*30 could be deleted)
	assertCounts(1*30 + 10)
}

func TestVacuumDeletedReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123)

	// TODO - setup

	_, err := store.VacuumDeletedReferences(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error vacuuming deleted references: %s", err)
	}

	// TODO - assertions
}

//
//

func getRankingReferences(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (_ []shared.RankingReferences, err error) {
	query := fmt.Sprintf(
		`SELECT upload_id, symbol_names FROM codeintel_ranking_references WHERE graph_key = '%s' AND deleted_at IS NULL`,
		graphKey,
	)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

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

	return references, nil
}
