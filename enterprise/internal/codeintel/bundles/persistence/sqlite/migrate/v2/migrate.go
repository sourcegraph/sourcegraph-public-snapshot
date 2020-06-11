package v2

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/store"
)

// Migrate v2: Rename meta.numResultChunks to meta.num_result_chunks and drop the version columns.
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	queries := []*sqlf.Query{
		sqlf.Sprintf(`CREATE TABLE t_meta (num_result_chunks int NOT NULL)`),
		sqlf.Sprintf(`INSERT INTO t_meta (num_result_chunks) SELECT numResultChunks FROM meta`),
		sqlf.Sprintf(`DROP TABLE meta`),
		sqlf.Sprintf(`ALTER TABLE t_meta RENAME TO meta`),
	}

	for _, query := range queries {
		if err := s.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}
