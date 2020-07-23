package v1

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/store"
)

// Migrate v1: Create a schema_version table that will explicitly declare a bundle's format.
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	queries := []*sqlf.Query{
		sqlf.Sprintf(`CREATE TABLE schema_version ("version" INT NOT NULL)`),
		sqlf.Sprintf(`INSERT INTO schema_version (version) VALUES (%s);`, 1),
	}

	for _, query := range queries {
		if err := s.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}
