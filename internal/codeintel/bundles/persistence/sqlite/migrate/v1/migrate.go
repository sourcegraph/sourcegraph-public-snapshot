package v1

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
)

// Migrate v1: Create a schema_version table that will explicitly declare a bundle's format.
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	return s.ExecAll(
		ctx,
		sqlf.Sprintf(`CREATE TABLE schema_version ("version" TEXT NOT NULL)`),
		sqlf.Sprintf(`INSERT INTO schema_version (version) VALUES (%s);`, "v00001"),
	)
}
