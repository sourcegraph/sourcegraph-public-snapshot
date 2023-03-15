package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDeleteLsifDataByUploadIds(t *testing.T) {
	logger := logtest.ScopedWith(t, logtest.LoggerOptions{
		Level: log.LevelError,
	})
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, codeIntelDB)

	t.Run("scip", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			query := sqlf.Sprintf("INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tooL_name, tool_version, tool_arguments, protocol_version) VALUES (%s, 'utf8', '', '', '{}', 1)", i+1)

			if _, err := codeIntelDB.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
				t.Fatalf("unexpected error inserting repo: %s", err)
			}
		}

		if err := store.DeleteLsifDataByUploadIds(context.Background(), 2, 4); err != nil {
			t.Fatalf("unexpected error clearing bundle data: %s", err)
		}

		dumpIDs, err := basestore.ScanInts(codeIntelDB.QueryContext(context.Background(), "SELECT upload_id FROM codeintel_scip_metadata"))
		if err != nil {
			t.Fatalf("Unexpected error querying dump identifiers: %s", err)
		}

		if diff := cmp.Diff([]int{1, 3, 5}, dumpIDs); diff != "" {
			t.Errorf("unexpected dump identifiers (-want +got):\n%s", diff)
		}
	})
}
