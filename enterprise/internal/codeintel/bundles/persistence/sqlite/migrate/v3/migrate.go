package v3

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/store"
)

// Migrate v3: Rename the resultChunks table to result_chunks.
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	return s.Exec(ctx, sqlf.Sprintf(`ALTER TABLE resultChunks RENAME TO result_chunks`))
}
