package postgres

import (
	"context"
	"runtime"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/util"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

// NumWriterRoutines is the number of goroutines launched to write database records.
var NumWriterRoutines = runtime.NumCPU() * 2

func (w *reader) WriteMeta(ctx context.Context, meta types.MetaData) error {
	inserter := NewBatchInserter(w.Handle().DB(), "lsif_data_metadata", "dump_id", "num_result_chunks")

	if err := inserter.Insert(ctx, w.dumpID, meta.NumResultChunks); err != nil {
		return err
	}

	if err := inserter.Flush(ctx); err != nil {
		return errors.Wrap(err, "inserter.Flush")
	}

	return nil
}

func (w *reader) WriteDocuments(ctx context.Context, documents chan persistence.KeyedDocumentData) error {
	return util.InvokeN(NumWriterRoutines, func() error {
		inserter := NewBatchInserter(w.Handle().DB(), "lsif_data_documents", "dump_id", "path", "data")

		for v := range documents {
			data, err := w.serializer.MarshalDocumentData(v.Document)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, w.dumpID, v.Path, data); err != nil {
				return err
			}
		}

		if err := inserter.Flush(ctx); err != nil {
			return errors.Wrap(err, "inserter.Flush")
		}

		return nil
	})
}

func (w *reader) WriteResultChunks(ctx context.Context, resultChunks chan persistence.IndexedResultChunkData) error {
	return util.InvokeN(NumWriterRoutines, func() error {
		inserter := NewBatchInserter(w.Handle().DB(), "lsif_data_result_chunks", "dump_id", "idx", "data")

		for v := range resultChunks {
			data, err := w.serializer.MarshalResultChunkData(v.ResultChunk)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, w.dumpID, v.Index, data); err != nil {
				return err
			}
		}

		if err := inserter.Flush(ctx); err != nil {
			return errors.Wrap(err, "inserter.Flush")
		}

		return nil
	})
}

func (w *reader) WriteDefinitions(ctx context.Context, monikerLocations chan types.MonikerLocations) error {
	return util.InvokeN(NumWriterRoutines, func() error {
		inserter := NewBatchInserter(w.Handle().DB(), "lsif_data_definitions", "dump_id", "scheme", "identifier", "data")

		for v := range monikerLocations {
			data, err := w.serializer.MarshalLocations(v.Locations)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, w.dumpID, v.Scheme, v.Identifier, data); err != nil {
				return err
			}
		}

		if err := inserter.Flush(ctx); err != nil {
			return errors.Wrap(err, "inserter.Flush")
		}

		return nil
	})
}

func (w *reader) WriteReferences(ctx context.Context, monikerLocations chan types.MonikerLocations) error {
	return util.InvokeN(NumWriterRoutines, func() error {
		inserter := NewBatchInserter(w.Handle().DB(), "lsif_data_references", "dump_id", "scheme", "identifier", "data")

		for v := range monikerLocations {
			data, err := w.serializer.MarshalLocations(v.Locations)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, w.dumpID, v.Scheme, v.Identifier, data); err != nil {
				return err
			}
		}

		if err := inserter.Flush(ctx); err != nil {
			return errors.Wrap(err, "inserter.Flush")
		}

		return nil
	})
}

func (w *reader) Close(err error) error {
	return err
}
