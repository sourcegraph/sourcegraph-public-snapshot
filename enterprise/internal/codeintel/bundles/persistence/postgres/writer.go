package postgres

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

type writer struct {
	dumpID     int
	serializer serialization.Serializer
	writer     *batchWriter
}

func (w *reader) WriteMeta(ctx context.Context, meta types.MetaData) error {
	w.writer.Write(
		`INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES ($1, $2)`,
		w.dumpID,
		meta.NumResultChunks,
	)

	return nil
}

func (w *reader) WriteDocuments(ctx context.Context, documents chan persistence.KeyedDocumentData) error {
	for v := range documents {
		data, err := w.serializer.MarshalDocumentData(v.Document)
		if err != nil {
			return err
		}

		w.writer.Write(
			`INSERT INTO lsif_data_documents (dump_id, path, data) VALUES ($1, $2, $3)`,
			w.dumpID,
			v.Path,
			data,
		)
	}

	return nil
}

func (w *reader) WriteResultChunks(ctx context.Context, resultChunks chan persistence.IndexedResultChunkData) error {
	for v := range resultChunks {
		data, err := w.serializer.MarshalResultChunkData(v.ResultChunk)
		if err != nil {
			return err
		}

		w.writer.Write(
			`INSERT INTO lsif_data_result_chunks (dump_id, idx, data) VALUES ($1, $2, $3)`,
			w.dumpID,
			v.Index,
			data,
		)
	}

	return nil
}

func (w *reader) WriteDefinitions(ctx context.Context, monikerLocations chan types.MonikerLocations) error {
	for v := range monikerLocations {
		data, err := w.serializer.MarshalLocations(v.Locations)
		if err != nil {
			return err
		}

		w.writer.Write(
			`INSERT INTO lsif_data_definitions (dump_id, scheme, identifier, data) VALUES ($1, $2, $3, $4)`,
			w.dumpID,
			v.Scheme,
			v.Identifier,
			data,
		)
	}

	return nil
}

func (w *reader) WriteReferences(ctx context.Context, monikerLocations chan types.MonikerLocations) error {
	for v := range monikerLocations {
		data, err := w.serializer.MarshalLocations(v.Locations)
		if err != nil {
			return err
		}

		w.writer.Write(
			`INSERT INTO lsif_data_references (dump_id, scheme, identifier, data) VALUES ($1, $2, $3, $4)`,
			w.dumpID,
			v.Scheme,
			v.Identifier,
			data,
		)
	}

	return nil
}

func (w *reader) Close(err error) error {
	return w.writer.Flush()
}
