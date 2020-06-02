package v5

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	jsonserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization/json"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
)

type MigrationStep func(context.Context, *store.Store, serialization.Serializer, serialization.Serializer) error

// Migrate v5: BLAH
func Migrate(ctx context.Context, s *store.Store, serializer serialization.Serializer) error {
	steps := []MigrationStep{
		reencodeDocuments,
		reencodeResultChunks,
		reencodeDefinitions,
		reencodeReferences,
	}

	deserializer := jsonserializer.New()

	for _, step := range steps {
		if err := step(ctx, s, deserializer, serializer); err != nil {
			return err
		}
	}

	return nil
}

// swapTables deletes the original table and replaces it with the temporary table.
func swapTables(ctx context.Context, s *store.Store, tableName, tempTableName string) error {
	queries := []*sqlf.Query{
		sqlf.Sprintf(`DROP TABLE "` + tableName + `"`),
		sqlf.Sprintf(`ALTER TABLE "` + tempTableName + `" RENAME TO "` + tableName + `"`),
	}

	for _, query := range queries {
		if err := s.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}
