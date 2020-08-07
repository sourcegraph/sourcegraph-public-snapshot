package scylladb

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
		`INSERT INTO metadata (dump_id, num_result_chunks) VALUES (?, ?)`,
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
			`INSERT INTO documents (dump_id, path, data) VALUES (?, ?, ?)`,
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
			`INSERT INTO result_chunks (dump_id, idx, data) VALUES (?, ?, ?)`,
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
			`INSERT INTO definitions (dump_id, scheme, identifier, data) VALUES (?, ?, ?, ?)`,
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
			`INSERT INTO references (dump_id, scheme, identifier, data) VALUES (?, ?, ?, ?)`,
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

// Run scylladb:
// docker run -p 9042:9042 -d scylladb/scylla

// Prepare schema:
// docker exec -it {container-id} cqlsh
//
// cqlsh> CREATE KEYSPACE lsif WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
// cqlsh> USE lsif;
// cqlsh:lsif> CREATE TABLE metadata (dump_id int, num_result_chunks int, PRIMARY KEY (dump_id));
// cqlsh:lsif> CREATE TABLE documents (dump_id int, path text, data blob, PRIMARY KEY (dump_id, path));
// cqlsh:lsif> CREATE TABLE result_chunks (dump_id int, idx int, data blob, PRIMARY KEY (dump_id, idx));
// cqlsh:lsif> CREATE TABLE definitions (dump_id int, scheme text, identifier text, data blob, PRIMARY KEY (dump_id, scheme, identifier));
// cqlsh:lsif> CREATE TABLE references (dump_id int, scheme text, identifier text, data blob, PRIMARY KEY (dump_id, scheme, identifier));
