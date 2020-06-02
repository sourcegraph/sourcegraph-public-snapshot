package v5

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

func reencodeDefinitions(ctx context.Context, s *store.Store, deserializer, serializer serialization.Serializer) error {
	return reencodeDefinitionReferences(ctx, s, "definitions", deserializer, serializer)
}

func reencodeReferences(ctx context.Context, s *store.Store, deserializer, serializer serialization.Serializer) error {
	return reencodeDefinitionReferences(ctx, s, "references", deserializer, serializer)
}

func reencodeDefinitionReferences(ctx context.Context, s *store.Store, tableName string, deserializer, serializer serialization.Serializer) error {
	if err := s.Exec(ctx, sqlf.Sprintf(`CREATE TABLE "t_`+tableName+`" ("scheme" text NOT NULL, "identifier" text NOT NULL, "data" blob NOT NULL, PRIMARY KEY (scheme, identifier))`)); err != nil {
		return err
	}

	rows, err := s.Query(ctx, sqlf.Sprintf(`SELECT scheme, identifier, data FROM "`+tableName+`"`))
	if err != nil {
		return err
	}
	defer func() {
		err = store.CloseRows(rows, err)
	}()

	inserter := sqliteutil.NewBatchInserter(s, fmt.Sprintf("t_%s", tableName), "scheme", "identifier", "data")

	for rows.Next() {
		var scheme string
		var identifier string
		var data []byte
		if err := rows.Scan(&scheme, &identifier, &data); err != nil {
			return err
		}

		locations, err := deserializer.UnmarshalLocations(data)
		if err != nil {
			return err
		}

		newData, err := serializer.MarshalLocations(locations)
		if err != nil {
			return err
		}

		if err := inserter.Insert(ctx, scheme, identifier, newData); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	return swapTables(ctx, s, tableName, fmt.Sprintf("t_%s", tableName))
}
