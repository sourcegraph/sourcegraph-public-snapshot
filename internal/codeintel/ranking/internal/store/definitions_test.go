package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertDefinition(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	// Insert uploads
	insertUploads(t, db, uploadsshared.Upload{ID: 4})

	// Insert exported uploads
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_exports (id, upload_id, graph_key, upload_key)
		VALUES (104, 4, $1, md5('key-4'))
	`,
		mockRankingGraphKey,
	); err != nil {
		t.Fatalf("unexpected error inserting exported upload record: %s", err)
	}

	expectedDefinitions := []shared.RankingDefinitions{
		{
			UploadID:         4,
			ExportedUploadID: 104,
			SymbolChecksum:   hash("foo"),
			DocumentPath:     "foo.go",
		},
		{
			UploadID:         4,
			ExportedUploadID: 104,
			SymbolChecksum:   hash("bar"),
			DocumentPath:     "bar.go",
		},
		{
			UploadID:         4,
			ExportedUploadID: 104,
			SymbolChecksum:   hash("foo"),
			DocumentPath:     "foo.go",
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

//
//

func getRankingDefinitions(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (_ []shared.RankingDefinitions, err error) {
	query := fmt.Sprintf(`
		SELECT cre.upload_id, cre.id, rd.symbol_checksum, rd.document_path
		FROM codeintel_ranking_definitions rd
		JOIN codeintel_ranking_exports cre ON cre.id = rd.exported_upload_id
		WHERE rd.graph_key = '%s'
	`,
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
		var exportedUploadID int
		var symbolChecksum []byte
		var documentPath string
		err = rows.Scan(&uploadID, &exportedUploadID, &symbolChecksum, &documentPath)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, shared.RankingDefinitions{
			UploadID:         uploadID,
			ExportedUploadID: exportedUploadID,
			SymbolChecksum:   castToChecksum(symbolChecksum),
			DocumentPath:     documentPath,
		})
	}

	return definitions, nil
}
