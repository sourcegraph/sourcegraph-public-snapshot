package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertReferences(t *testing.T) {
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

	// Insert references
	mockReferences := make(chan [16]byte, 3)
	mockReferences <- hash("foo")
	mockReferences <- hash("bar")
	mockReferences <- hash("baz")
	close(mockReferences)
	if err := store.InsertReferencesForRanking(ctx, mockRankingGraphKey, mockRankingBatchSize, 104, mockReferences); err != nil {
		t.Fatalf("unexpected error inserting references: %s", err)
	}

	// Test references were inserted
	references, err := getRankingReferences(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting references: %s", err)
	}

	expectedReferences := []shared.RankingReferences{
		{
			UploadID:         4,
			ExportedUploadID: 104,
			SymbolChecksums:  [][16]byte{hash("foo"), hash("bar"), hash("baz")},
		},
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

//
//

func getRankingReferences(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (_ []shared.RankingReferences, err error) {
	query := fmt.Sprintf(`
		SELECT cre.upload_id, cre.id, rd.symbol_checksums
		FROM codeintel_ranking_references rd
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

	var references []shared.RankingReferences
	for rows.Next() {
		var uploadID int
		var exportedUploadID int
		var symbolChecksums [][]byte
		err = rows.Scan(&uploadID, &exportedUploadID, pq.Array(&symbolChecksums))
		if err != nil {
			return nil, err
		}

		references = append(references, shared.RankingReferences{
			UploadID:         uploadID,
			ExportedUploadID: exportedUploadID,
			SymbolChecksums:  castToChecksums(symbolChecksums),
		})
	}

	return references, nil
}
