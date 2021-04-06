package lsifstore

import (
	"context"
	"sync/atomic"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// CurrentDocumentSchemaVersion is the schema version used for new lsif_data_documents rows.
const CurrentDocumentSchemaVersion = 2

// CurrentDefinitionsSchemaVersion is the schema version used for new lsif_data_definitions rows.
const CurrentDefinitionsSchemaVersion = 2

// CurrentReferencesSchemaVersion is the schema version used for new lsif_data_references rows.
const CurrentReferencesSchemaVersion = 2

func (s *Store) WriteMeta(ctx context.Context, bundleID int, meta semantic.MetaData) (err error) {
	ctx, endObservation := s.operations.writeMeta.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	return s.Exec(ctx, sqlf.Sprintf("INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES (%s, %s)", bundleID, meta.NumResultChunks))
}

func (s *Store) WriteDocuments(ctx context.Context, bundleID int, documents chan semantic.KeyedDocumentData) (err error) {
	ctx, traceLog, endObservation := s.operations.writeDocuments.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	// Create temporary table symmetric to lsif_data_documents without the dump id or schema version
	if err := s.Exec(ctx, sqlf.Sprintf(writeDocumentsTemporaryTableQuery)); err != nil {
		return err
	}

	var count uint32
	db := s.Handle().DB()
	columns := []string{"path", "data", "num_diagnostics"}
	inserter := func(inserter *batch.Inserter) error {
		for v := range documents {
			data, err := s.serializer.MarshalDocumentData(v.Document)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, v.Path, data, len(v.Document.Diagnostics)); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}

		return nil
	}

	// Bulk insert all the unique column values into the temporary table
	if err := withBatchInserter(ctx, db, "t_lsif_data_documents", columns, inserter); err != nil {
		return err
	}

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id and schema version here since it is the same for all rows
	// in this operation.
	if err := s.Exec(ctx, sqlf.Sprintf(writeDocumentsInsertQuery, bundleID, CurrentDocumentSchemaVersion)); err != nil {
		return err
	}
	traceLog(log.Int("count", int(count)))
	return nil
}

const writeDocumentsTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write.go:WriteDocuments
CREATE TEMPORARY TABLE t_lsif_data_documents (
	path text NOT NULL,
	data bytea NOT NULL,
	num_diagnostics integer NOT NULL
) ON COMMIT DROP
`

const writeDocumentsInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write.go:WriteDocuments
INSERT INTO lsif_data_documents (dump_id, schema_version, path, data, num_diagnostics)
SELECT %s, %s, source.path, source.data, source.num_diagnostics
FROM t_lsif_data_documents source
`

func (s *Store) WriteResultChunks(ctx context.Context, bundleID int, resultChunks chan semantic.IndexedResultChunkData) (err error) {
	ctx, traceLog, endObservation := s.operations.writeResultChunks.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	// Create temporary table symmetric to lsif_data_documents without the dump id
	if err := s.Exec(ctx, sqlf.Sprintf(writeResultChunksTemporaryTableQuery)); err != nil {
		return err
	}

	var count uint32
	db := s.Handle().DB()
	columns := []string{"idx", "data"}
	inserter := func(inserter *batch.Inserter) error {
		for v := range resultChunks {
			data, err := s.serializer.MarshalResultChunkData(v.ResultChunk)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, v.Index, data); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}

		return nil
	}

	// Bulk insert all the unique column values into the temporary table
	if err := withBatchInserter(ctx, db, "t_lsif_data_result_chunks", columns, inserter); err != nil {
		return err
	}

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id here since it is the same for all rows in this operation.
	if err := s.Exec(ctx, sqlf.Sprintf(writeResultChunksInsertQuery, bundleID)); err != nil {
		return err
	}

	traceLog(log.Int("count", int(count)))
	return nil
}

const writeResultChunksTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write.go:WriteResultChunks
CREATE TEMPORARY TABLE t_lsif_data_result_chunks (
	idx integer NOT NULL,
	data bytea NOT NULL
) ON COMMIT DROP
`

const writeResultChunksInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write.go:WriteResultChunks
INSERT INTO lsif_data_result_chunks (dump_id, idx, data)
SELECT %s, source.idx, source.data
FROM t_lsif_data_result_chunks source
`

func (s *Store) WriteDefinitions(ctx context.Context, bundleID int, monikerLocations chan semantic.MonikerLocations) (err error) {
	ctx, traceLog, endObservation := s.operations.writeDefinitions.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	count, err := s.writeDefinitionReferences(ctx, bundleID, "lsif_data_definitions", CurrentDefinitionsSchemaVersion, monikerLocations)
	if err != nil {
		return err
	}
	traceLog(log.Int("count", count))
	return nil
}

func (s *Store) WriteReferences(ctx context.Context, bundleID int, monikerLocations chan semantic.MonikerLocations) (err error) {
	ctx, traceLog, endObservation := s.operations.writeReferences.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	count, err := s.writeDefinitionReferences(ctx, bundleID, "lsif_data_references", CurrentReferencesSchemaVersion, monikerLocations)
	if err != nil {
		return err
	}
	traceLog(log.Int("count", count))
	return nil
}

func (s *Store) writeDefinitionReferences(ctx context.Context, bundleID int, tableName string, version int, monikerLocations chan semantic.MonikerLocations) (int, error) {
	// Create temporary table symmetric to the given target table without the dump id or schema version
	if err := s.Exec(ctx, sqlf.Sprintf(writeDefinitionsReferencesTemporaryTableQuery, sqlf.Sprintf(tableName))); err != nil {
		return 0, err
	}

	var count uint32
	db := s.Handle().DB()
	columns := []string{"scheme", "identifier", "data", "num_locations"}
	inserter := func(inserter *batch.Inserter) error {
		for v := range monikerLocations {
			data, err := s.serializer.MarshalLocations(v.Locations)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, v.Scheme, v.Identifier, data, len(v.Locations)); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}

		return nil
	}

	// Bulk insert all the unique column values into the temporary table
	if err := withBatchInserter(ctx, db, "t_"+tableName, columns, inserter); err != nil {
		return 0, err
	}

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id and schema version here since it is the same for all rows
	// in this operation.
	if err := s.Exec(ctx, sqlf.Sprintf(writeDefinitionReferencesInsertQuery, sqlf.Sprintf(tableName), bundleID, version, sqlf.Sprintf(tableName))); err != nil {
		return 0, err
	}

	return int(count), nil
}

const writeDefinitionsReferencesTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write.go:writeDefinitionReferences
CREATE TEMPORARY TABLE t_%s (
	scheme text NOT NULL,
	identifier text NOT NULL,
	data bytea NOT NULL,
	num_locations integer NOT NULL
) ON COMMIT DROP
`

const writeDefinitionReferencesInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write.go:writeDefinitionReferences
INSERT INTO %s (dump_id, schema_version, scheme, identifier, data, num_locations)
SELECT %s, %s, source.scheme, source.identifier, source.data, source.num_locations
FROM t_%s source
`

// withBatchInserter runs batch.WithInserter in as a number of goroutines proportional to
// the maximum number of CPUs that can be executing simultaneously.
func withBatchInserter(ctx context.Context, db dbutil.DB, tableName string, columns []string, f func(inserter *batch.Inserter) error) (err error) {
	return goroutine.RunWorkers(goroutine.SimplePoolWorker(func() error {
		return batch.WithInserter(ctx, db, tableName, columns, f)
	}))
}
