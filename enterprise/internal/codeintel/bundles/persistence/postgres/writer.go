package postgres

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	gobserializer "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization/gob"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

type writer struct {
	dumpID     int
	serializer serialization.Serializer
	writer     *batchWriter
}

var _ persistence.Writer = &writer{}

func NewWriter(dumpID int) persistence.Writer {
	return &writer{
		dumpID:     dumpID,
		serializer: gobserializer.New(),
		writer:     newBatchWriter(),
	}
}

func (w *writer) WriteMeta(ctx context.Context, meta types.MetaData) error {
	w.writer.Write(
		`INSERT INTO lsif_data_metadata (dump_id, num_result_chunks) VALUES ($1, $2)`,
		w.dumpID,
		meta.NumResultChunks,
	)

	return nil
}

func (w *writer) WriteDocuments(ctx context.Context, documents map[string]types.DocumentData) error {
	for path, document := range documents {
		data, err := w.serializer.MarshalDocumentData(document)
		if err != nil {
			return err
		}

		w.writer.Write(
			`INSERT INTO lsif_data_documents (dump_id, path, data) VALUES ($1, $2, $3)`,
			w.dumpID,
			path,
			data,
		)
	}

	return nil
}

func (w *writer) WriteResultChunks(ctx context.Context, resultChunks map[int]types.ResultChunkData) error {
	for idx, resultChunk := range resultChunks {
		data, err := w.serializer.MarshalResultChunkData(resultChunk)
		if err != nil {
			return err
		}

		w.writer.Write(
			`INSERT INTO lsif_data_result_chunks (dump_id, idx, data) VALUES ($1, $2, $3)`,
			w.dumpID,
			idx,
			data,
		)
	}

	return nil
}

func (w *writer) WriteDefinitions(ctx context.Context, monikerLocations []types.MonikerLocations) error {
	for _, v := range monikerLocations {
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

func (w *writer) WriteReferences(ctx context.Context, monikerLocations []types.MonikerLocations) error {
	for _, v := range monikerLocations {
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

func (w *writer) Close(err error) error {
	return w.writer.Flush()
}
