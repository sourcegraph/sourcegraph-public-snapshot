package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertInitialPathRanks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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

	mockPathNames := []string{
		"foo.go",
		"bar.go",
		"baz.go",
	}
	if err := store.InsertInitialPathRanks(ctx, 104, mockPathNames, 2, mockRankingGraphKey); err != nil {
		t.Fatalf("unexpected error inserting initial path counts: %s", err)
	}

	inputs, err := getInitialPathRanks(ctx, t, db, mockRankingGraphKey)
	if err != nil {
		t.Fatalf("unexpected error getting path count inputs: %s", err)
	}

	expectedInputs := []initialPathRanks{
		{UploadID: 4, DocumentPath: "bar.go"},
		{UploadID: 4, DocumentPath: "baz.go"},
		{UploadID: 4, DocumentPath: "foo.go"},
	}
	if diff := cmp.Diff(expectedInputs, inputs); diff != "" {
		t.Errorf("unexpected path count inputs (-want +got):\n%s", diff)
	}
}

//
//

type initialPathRanks struct {
	UploadID     int
	DocumentPath string
}

func getInitialPathRanks(
	ctx context.Context,
	t *testing.T,
	db database.DB,
	graphKey string,
) (pathRanks []initialPathRanks, err error) {
	query := sqlf.Sprintf(`
		SELECT
			s.upload_id,
			s.document_path
		FROM (
			SELECT
				cre.upload_id,
				unnest(pr.document_paths) AS document_path
			FROM codeintel_initial_path_ranks pr
			JOIN codeintel_ranking_exports cre ON cre.id = pr.exported_upload_id
			WHERE pr.graph_key LIKE %s || '%%'
		)s
		GROUP BY s.upload_id, s.document_path
		ORDER BY s.upload_id, s.document_path
	`, graphKey)
	rows, err := db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var input initialPathRanks
		if err := rows.Scan(&input.UploadID, &input.DocumentPath); err != nil {
			return nil, err
		}

		pathRanks = append(pathRanks, input)
	}

	return pathRanks, nil
}
