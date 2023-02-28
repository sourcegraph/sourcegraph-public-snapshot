package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const (
	mockRankingGraphKey    = "mockDev" // NOTE: ensure we don't have hyphens so we can validate derivative keys easily
	mockRankingBatchNumber = 10
)

func TestInsertDefinition(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	// Insert definitions
	mockDefinitions := []shared.RankingDefinitions{
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
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}

	// Test definitions were inserted
	definitions, err := getRankingDefinitions(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting definitions: %s", err)
	}

	if diff := cmp.Diff(mockDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

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

	// Test references were inserted
	references, err := getRankingReferences(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting references: %s", err)
	}

	if diff := cmp.Diff(mockReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
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

	// Insert definitions
	mockDefinitions := []shared.RankingDefinitions{
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
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockDefinitions); err != nil {
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
	if _, _, err := store.InsertPathCountInputs(ctx, mockRankingGraphKey+"-123", 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	// Insert repos
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO repo (id, name) VALUES (1, 'deadbeef')`)); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	// Finally! Test InsertPathRanks
	numPathRanksInserted, numInputsProcessed, err := store.InsertPathRanks(ctx, mockRankingGraphKey+"-123", 10)
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

	// Insert definitions
	mockDefinitions := []shared.RankingDefinitions{
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
	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockDefinitions); err != nil {
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
	if _, _, err := store.InsertPathCountInputs(ctx, mockRankingGraphKey+"-123", 1000); err != nil {
		t.Fatalf("unexpected error inserting path count inputs: %s", err)
	}

	// Test path count inputs were inserted
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

func TestVacuumStaleDefinitionsAndReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	mockDefinitions := []shared.RankingDefinitions{
		{UploadID: 1, SymbolName: "foo", Repository: "deadbeef", DocumentPath: "foo.go"},
		{UploadID: 1, SymbolName: "bar", Repository: "deadbeef", DocumentPath: "bar.go"},
		{UploadID: 2, SymbolName: "foo", Repository: "deadbeef", DocumentPath: "foo.go"},
		{UploadID: 2, SymbolName: "bar", Repository: "deadbeef", DocumentPath: "bar.go"},
		{UploadID: 3, SymbolName: "baz", Repository: "deadbeef", DocumentPath: "baz.go"},
	}
	mockReferences := []shared.RankingReferences{
		{UploadID: 1, SymbolNames: []string{"foo"}},
		{UploadID: 1, SymbolNames: []string{"bar"}},
		{UploadID: 2, SymbolNames: []string{"foo"}},
		{UploadID: 2, SymbolNames: []string{"bar"}},
		{UploadID: 2, SymbolNames: []string{"baz"}},
		{UploadID: 3, SymbolNames: []string{"bar"}},
		{UploadID: 3, SymbolNames: []string{"baz"}},
	}

	if err := store.InsertDefinitionsForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, mockDefinitions); err != nil {
		t.Fatalf("unexpected error inserting definitions: %s", err)
	}
	for _, reference := range mockReferences {
		if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, reference); err != nil {
			t.Fatalf("unexpected error inserting references: %s", err)
		}
	}

	assertCounts := func(expectedNumDefinitions, expectedNumReferences int) {
		definitions, err := getRankingDefinitions(ctx, t, db, mockRankingGraphKey)
		if err != nil {
			t.Fatalf("failed to get ranking definitions: %s", err)
		}
		if len(definitions) != expectedNumDefinitions {
			t.Fatalf("unexpected number of definitions. want=%d have=%d", expectedNumDefinitions, len(definitions))
		}

		references, err := getRankingReferences(ctx, t, db, mockRankingGraphKey)
		if err != nil {
			t.Fatalf("failed to get ranking references: %s", err)
		}
		if len(references) != expectedNumReferences {
			t.Fatalf("unexpected number of references. want=%d have=%d", expectedNumReferences, len(references))
		}
	}

	// assert initial count
	assertCounts(5, 7)

	// make upload 2 visible at tip (1 and 3 are not)
	insertVisibleAtTip(t, db, 50, 2)

	// remove definitions and references for non-visible uploads
	numStaleDefinitionRecordsDeleted, numStaleReferenceRecordsDeleted, err := store.VacuumStaleDefinitionsAndReferences(ctx, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error vacuuming stale definitions and references: %s", err)
	}
	if expected := 3; numStaleDefinitionRecordsDeleted != expected {
		t.Fatalf("unexpected number of definition records deleted. want=%d have=%d", expected, numStaleDefinitionRecordsDeleted)
	}
	if expected := 4; numStaleReferenceRecordsDeleted != expected {
		t.Fatalf("unexpected number of reference records deleted. want=%d have=%d", expected, numStaleReferenceRecordsDeleted)
	}

	// only upload 2's entries remain
	assertCounts(2, 3)
}

func TestVacuumStaleGraphs(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	mockReferences := []shared.RankingReferences{
		{UploadID: 1, SymbolNames: []string{"foo"}},
		{UploadID: 1, SymbolNames: []string{"bar"}},
		{UploadID: 2, SymbolNames: []string{"foo"}},
		{UploadID: 2, SymbolNames: []string{"bar"}},
		{UploadID: 2, SymbolNames: []string{"baz"}},
		{UploadID: 3, SymbolNames: []string{"bar"}},
		{UploadID: 3, SymbolNames: []string{"baz"}},
	}
	for _, reference := range mockReferences {
		if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchNumber, reference); err != nil {
			t.Fatalf("unexpected error inserting references: %s", err)
		}
	}

	for _, graphKey := range []string{mockRankingGraphKey + "-123", mockRankingGraphKey + "-456", mockRankingGraphKey + "-789"} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_ranking_references_processed (graph_key, codeintel_ranking_reference_id)
			SELECT $1, id FROM codeintel_ranking_references
	`, graphKey); err != nil {
			t.Fatalf("failed to insert ranking references processed: %s", err)
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_ranking_path_counts_inputs (repository, document_path, count, graph_key)
			SELECT 50, '', 100, $1 FROM generate_series(1, 30)
	`, graphKey); err != nil {
			t.Fatalf("failed to insert ranking path count inputs: %s", err)
		}
	}

	assertCounts := func(expectedMetadataRecords, expectedInputRecords int) {
		store := basestore.NewWithHandle(db.Handle())

		numMetadataRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_ranking_references_processed`)))
		if err != nil {
			t.Fatalf("failed to count metadata records: %s", err)
		}
		if expectedMetadataRecords != numMetadataRecords {
			t.Fatalf("unexpected number of metadata records. want=%d have=%d", expectedMetadataRecords, numMetadataRecords)
		}

		numInputRecords, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_ranking_path_counts_inputs`)))
		if err != nil {
			t.Fatalf("failed to count input records: %s", err)
		}
		if expectedInputRecords != numInputRecords {
			t.Fatalf("unexpected number of input records. want=%d have=%d", expectedInputRecords, numInputRecords)
		}
	}

	// assert initial count
	assertCounts(3*7, 3*30)

	// remove records associated with other ranking keys
	metadataRecordsDeleted, inputRecordsDeleted, err := store.VacuumStaleGraphs(ctx, mockRankingGraphKey+"-456")
	if err != nil {
		t.Fatalf("unexpected error vacuuming stale graphs: %s", err)
	}
	if expected := 2 * 7; metadataRecordsDeleted != expected {
		t.Fatalf("unexpected number of metadata records deleted. want=%d have=%d", expected, metadataRecordsDeleted)
	}
	if expected := 2 * 30; inputRecordsDeleted != expected {
		t.Fatalf("unexpected number of input records deleted. want=%d have=%d", expected, inputRecordsDeleted)
	}

	// only the non-stale derivative graph key remains
	assertCounts(1*7, 1*30)
}

func getRankingDefinitions(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (_ []shared.RankingDefinitions, err error) {
	query := fmt.Sprintf(
		`SELECT upload_id, symbol_name, repository, document_path FROM codeintel_ranking_definitions WHERE graph_key = '%s'`,
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
		var repository string
		var documentPath string
		err = rows.Scan(&uploadID, &symbolName, &repository, &documentPath)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, shared.RankingDefinitions{
			UploadID:     uploadID,
			SymbolName:   symbolName,
			Repository:   repository,
			DocumentPath: documentPath,
		})
	}

	return definitions, nil
}

func getRankingReferences(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (_ []shared.RankingReferences, err error) {
	query := fmt.Sprintf(
		`SELECT upload_id, symbol_names FROM codeintel_ranking_references WHERE graph_key = '%s'`,
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

func getRankingPathCountsInputs(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (repository, documentPath string, count int, err error) {
	query := sqlf.Sprintf(
		`SELECT repository, document_path, count FROM codeintel_ranking_path_counts_inputs WHERE graph_key LIKE %s || '%%'`,
		graphKey,
	)
	rows, err := db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return "", "", 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		err = rows.Scan(&repository, &documentPath, &count)
		if err != nil {
			return "", "", 0, err
		}
	}

	return repository, documentPath, count, nil
}
