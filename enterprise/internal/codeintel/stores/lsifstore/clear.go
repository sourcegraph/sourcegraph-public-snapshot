package lsifstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var tableNames = []string{
	"lsif_data_metadata",
	"lsif_data_documents",
	"lsif_data_result_chunks",
	"lsif_data_definitions",
	"lsif_data_references",
}

func (s *Store) Clear(ctx context.Context, bundleIDs ...int) (err error) {
	var stringIDs []string
	for _, bundleID := range bundleIDs {
		stringIDs = append(stringIDs, fmt.Sprintf("%d", bundleID))
	}

	ctx, endObservation := s.operations.clear.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("bundleIDs", strings.Join(stringIDs, ", ")),
	}})
	defer endObservation(1, observation.Args{})

	if len(bundleIDs) == 0 {
		return nil
	}

	var ids []*sqlf.Query
	for _, bundleID := range bundleIDs {
		ids = append(ids, sqlf.Sprintf("%d", bundleID))
	}

	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	for _, tableName := range tableNames {
		if err := tx.Exec(ctx, sqlf.Sprintf(`DELETE FROM "`+tableName+`" WHERE dump_id IN (%s)`, sqlf.Join(ids, ","))); err != nil {
			return err
		}
	}

	return nil
}
