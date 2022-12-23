package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func (s *store) ScanDocuments(ctx context.Context, id int, f func(path string, ranges map[precise.ID]precise.RangeData) error) (err error) {
	ctx, _, endObservation := s.operations.scanDocuments.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return runQuery(ctx, s.db, sqlf.Sprintf(scanDocumentsQuery, id), func(dbs dbutil.Scanner) error {
		var path string
		var rawRanges []byte
		if err := dbs.Scan(&path, &rawRanges); err != nil {
			return err
		}

		var ranges map[precise.ID]precise.RangeData
		if err := s.serializer.decode(rawRanges, &ranges); err != nil {
			return err
		}

		return f(path, ranges)
	})
}

const scanDocumentsQuery = `
SELECT path, ranges
FROM lsif_data_documents
WHERE dump_id = %s
ORDER BY path
`

func (s *store) ScanResultChunks(ctx context.Context, id int, f func(idx int, resultChunk precise.ResultChunkData) error) (err error) {
	ctx, _, endObservation := s.operations.scanResultChunks.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return runQuery(ctx, s.db, sqlf.Sprintf(scanResultChunksQuery, id), func(dbs dbutil.Scanner) error {
		var idx int
		var rawData []byte
		if err := dbs.Scan(&idx, &rawData); err != nil {
			return err
		}

		var resultChunk precise.ResultChunkData
		if err := s.serializer.decode(rawData, &resultChunk); err != nil {
			return err
		}

		return f(idx, resultChunk)
	})
}

const scanResultChunksQuery = `
SELECT idx, data
FROM lsif_data_result_chunks
WHERE dump_id = %s
ORDER BY idx
`

func (s *store) ScanLocations(ctx context.Context, id int, f func(scheme, identifier, monikerType string, locations []precise.LocationData) error) (err error) {
	ctx, _, endObservation := s.operations.scanLocations.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return runQuery(ctx, s.db, sqlf.Sprintf(scanLocationsQuery, id, id), func(dbs dbutil.Scanner) error {
		var scheme, identifier, monikerType string
		var rawData []byte
		if err := dbs.Scan(&scheme, &identifier, &monikerType, &rawData); err != nil {
			return err
		}

		var locations []precise.LocationData
		if err := s.serializer.decode(rawData, &locations); err != nil {
			return err
		}

		return f(scheme, identifier, monikerType, locations)
	})
}

const scanLocationsQuery = `
WITH
defs AS (
	SELECT
		d.scheme,
		d.identifier,
		'export' AS type,
		d.data
	FROM lsif_data_definitions d
	WHERE d.dump_id = %s
	ORDER BY d.scheme, d.identifier
),
refs AS (
	SELECT
		r.scheme,
		r.identifier,
		'import' AS type,
		r.data
	FROM lsif_data_references r
	WHERE r.dump_id = %s
	ORDER BY r.scheme, r.identifier
)
SELECT * FROM defs UNION ALL SELECT * FROM refs
`

func runQuery(ctx context.Context, store *basestore.Store, query *sqlf.Query, f func(dbutil.Scanner) error) (err error) {
	rows, queryErr := store.Query(ctx, query)
	if queryErr != nil {
		return queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := f(rows); err != nil {
			return err
		}
	}

	return nil
}
