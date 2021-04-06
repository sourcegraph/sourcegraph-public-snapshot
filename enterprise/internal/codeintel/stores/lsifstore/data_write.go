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

	var count uint32

	inserter := func(inserter *batch.Inserter) error {
		for v := range documents {
			data, err := s.serializer.MarshalDocumentData(v.Document)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, bundleID, v.Path, data, CurrentDocumentSchemaVersion, len(v.Document.Diagnostics)); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}

		return nil
	}

	if err := withBatchInserter(ctx, s.Handle().DB(), "lsif_data_documents", []string{"dump_id", "path", "data", "schema_version", "num_diagnostics"}, inserter); err != nil {
		return err
	}
	traceLog(log.Int("count", int(count)))
	return nil
}

func (s *Store) WriteResultChunks(ctx context.Context, bundleID int, resultChunks chan semantic.IndexedResultChunkData) (err error) {
	ctx, traceLog, endObservation := s.operations.writeResultChunks.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	var count uint32

	inserter := func(inserter *batch.Inserter) error {
		for v := range resultChunks {
			data, err := s.serializer.MarshalResultChunkData(v.ResultChunk)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, bundleID, v.Index, data); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}

		return nil
	}

	if err := withBatchInserter(ctx, s.Handle().DB(), "lsif_data_result_chunks", []string{"dump_id", "idx", "data"}, inserter); err != nil {
		return err
	}
	traceLog(log.Int("count", int(count)))
	return nil
}

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
	var count uint32

	inserter := func(inserter *batch.Inserter) error {
		for v := range monikerLocations {
			data, err := s.serializer.MarshalLocations(v.Locations)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, bundleID, v.Scheme, v.Identifier, data, version, len(v.Locations)); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}

		return nil
	}

	err := withBatchInserter(ctx, s.Handle().DB(), tableName, []string{"dump_id", "scheme", "identifier", "data", "schema_version", "num_locations"}, inserter)
	return int(count), err
}

func withBatchInserter(ctx context.Context, db dbutil.DB, tableName string, columns []string, f func(inserter *batch.Inserter) error) (err error) {
	return goroutine.RunWorkers(goroutine.SimplePoolWorker(func() error {
		inserter := batch.NewInserter(ctx, db, tableName, columns...)
		defer func() {
			if flushErr := inserter.Flush(ctx); flushErr != nil {
				err = multierror.Append(err, errors.Wrap(flushErr, "inserter.Flush"))
			}
		}()

		return f(inserter)
	}))
}
