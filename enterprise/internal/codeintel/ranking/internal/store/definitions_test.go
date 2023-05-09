package store

import (
	"context"
	"fmt"
	"testing"

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

func TestInsertDefinition(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	expectedDefinitions := []shared.RankingDefinitions{
		{
			UploadID:     1,
			SymbolName:   "foo",
			DocumentPath: "foo.go",
		},
		{
			UploadID:     1,
			SymbolName:   "bar",
			DocumentPath: "bar.go",
		},
		{
			UploadID:     1,
			SymbolName:   "foo",
			DocumentPath: "foo.go",
		},
	}

	// Insert definitions
	mockDefinitions := make(chan shared.RankingDefinitions, len(expectedDefinitions))
	for _, def := range expectedDefinitions {
		mockDefinitions <- def
	}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Test definitions were inserted
	definitions, err := getRankingDefinitions(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting definitions: %s", err)
	}

	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

func TestVacuumAbandonedDefinitions(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	symbols := []string{}
	for j := 0; j < 30; j++ {
		symbols = append(symbols, fmt.Sprintf("s%d", j+1))
	}

	mockDefinitions1 := make(chan shared.RankingDefinitions, len(symbols))
	mockDefinitions2 := make(chan shared.RankingDefinitions, len(symbols))
	mockDefinitions3 := make(chan shared.RankingDefinitions, len(symbols))
	for _, symbol := range symbols {
		mockDefinitions1 <- shared.RankingDefinitions{UploadID: 1, SymbolName: symbol, DocumentPath: "foo.go"}
		mockDefinitions2 <- shared.RankingDefinitions{UploadID: 1, SymbolName: symbol, DocumentPath: "foo.go"}
		mockDefinitions3 <- shared.RankingDefinitions{UploadID: 1, SymbolName: symbol, DocumentPath: "foo.go"}
	}
	close(mockDefinitions1)
	close(mockDefinitions2)
	close(mockDefinitions3)

	// Insert definitions
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey+"-abandoned", mockDefinitions1); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockDefinitions2); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey+"-abandoned", mockDefinitions3); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	assertCounts := func(expectedDefinitionRecords int) {
		store := basestore.NewWithHandle(db.Handle())

		numDefinitionRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_ranking_definitions`)))
		if err != nil {
			t.Fatalf("failed to definition records: %s", err)
		}
		if expectedDefinitionRecords != numDefinitionRecords {
			t.Fatalf("unexpected number of definition records. want=%d have=%d", expectedDefinitionRecords, numDefinitionRecords)
		}
	}

	// assert initial count
	assertCounts(3 * 30)

	// remove records associated with other ranking keys
	if _, err := store.VacuumAbandonedDefinitions(ctx, mockRankingGraphKey, 50); err != nil {
		t.Fatalf("unexpected error vacuuming abandoned definitions: %s", err)
	}

	// only 10 records of stale derivative graph key remain (batch size of 50, but 2*30 could be deleted)
	assertCounts(1*30 + 10)
}

func TestSoftDeleteStaleDefinitionsAndReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	insertUploads(t, db,
		uploadsshared.Upload{ID: 1},
		uploadsshared.Upload{ID: 2},
		uploadsshared.Upload{ID: 3},
	)

	mockDefinitions := make(chan shared.RankingDefinitions, 5)
	mockDefinitions <- shared.RankingDefinitions{UploadID: 1, SymbolName: "foo", DocumentPath: "foo.go"}
	mockDefinitions <- shared.RankingDefinitions{UploadID: 1, SymbolName: "bar", DocumentPath: "bar.go"}
	mockDefinitions <- shared.RankingDefinitions{UploadID: 2, SymbolName: "foo", DocumentPath: "foo.go"}
	mockDefinitions <- shared.RankingDefinitions{UploadID: 2, SymbolName: "bar", DocumentPath: "bar.go"}
	mockDefinitions <- shared.RankingDefinitions{UploadID: 3, SymbolName: "baz", DocumentPath: "baz.go"}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

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

	assertCounts := func(expectedNumDefinitions, expectedNumSymbolReferences int) {
		definitions, err := getRankingDefinitions(ctx, t, db, mockRankingGraphKey)
		if err != nil {
			t.Fatalf("failed to get ranking definitions: %s", err)
		}
		if len(definitions) != expectedNumDefinitions {
			t.Errorf("unexpected number of definitions. want=%d have=%d", expectedNumDefinitions, len(definitions))
		}

		references, err := getRankingReferences(ctx, t, db, mockRankingGraphKey)
		if err != nil {
			t.Fatalf("failed to get ranking references: %s", err)
		}
		numSymbolReferences := 0
		for _, ref := range references {
			numSymbolReferences += len(ref.SymbolNames)
		}
		if numSymbolReferences != expectedNumSymbolReferences {
			t.Errorf("unexpected number of symbol references. want=%d have=%d", expectedNumSymbolReferences, len(references))
		}
	}

	// assert initial count
	assertCounts(5, 7)

	// make upload 2 visible at tip (1 and 3 are not)
	insertVisibleAtTip(t, db, 50, 2)

	// remove definitions for non-visible uploads
	_, numStaleDefinitionRecordsDeleted, err := store.SoftDeleteStaleDefinitions(ctx, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error vacuuming stale definitions: %s", err)
	}
	if expected := 3; numStaleDefinitionRecordsDeleted != expected {
		t.Errorf("unexpected number of definition records deleted. want=%d have=%d", expected, numStaleDefinitionRecordsDeleted)
	}

	// remove references for non-visible uploads
	if _, _, err := store.SoftDeleteStaleReferences(ctx, mockRankingGraphKey); err != nil {
		t.Fatalf("unexpected error vacuuming stale references: %s", err)
	}

	// only upload 2's entries remain
	assertCounts(2, 3)
}

func TestVacuumDeletedDefinitions(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123)

	// TODO - setup

	_, err := store.VacuumDeletedDefinitions(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error vacuuming deleted definitions: %s", err)
	}

	// TODO - assertions
}

//
//

func getRankingDefinitions(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (_ []shared.RankingDefinitions, err error) {
	query := fmt.Sprintf(
		`SELECT upload_id, symbol_name, document_path FROM codeintel_ranking_definitions WHERE graph_key = '%s' AND deleted_at IS NULL`,
		graphKey,
	)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var definitions []shared.RankingDefinitions
	for rows.Next() {
		var uploadID int
		var symbolName string
		var documentPath string
		err = rows.Scan(&uploadID, &symbolName, &documentPath)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, shared.RankingDefinitions{
			UploadID:     uploadID,
			SymbolName:   symbolName,
			DocumentPath: documentPath,
		})
	}

	return definitions, nil
}
