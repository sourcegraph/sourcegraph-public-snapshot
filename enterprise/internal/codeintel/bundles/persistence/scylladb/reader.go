package scylladb

import (
	"context"
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	gobserializer "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization/gob"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
)

type reader struct {
	dumpID     int
	serializer serialization.Serializer
}

var _ persistence.Reader = &reader{}

func NewReader(dumpID int) persistence.Reader {
	return &reader{
		dumpID:     dumpID,
		serializer: gobserializer.New(),
	}
}

func (r *reader) ReadMeta(ctx context.Context) (types.MetaData, error) {
	var numResultChunks int

	if err := session.Query(
		`SELECT num_result_chunks FROM metadata WHERE dump_id = ? LIMIT 1`,
		r.dumpID,
	).Scan(&numResultChunks); err != nil {
		return types.MetaData{}, err
	}

	return types.MetaData{NumResultChunks: numResultChunks}, nil
}

func (r *reader) PathsWithPrefix(ctx context.Context, prefix string) (px []string, _ error) {
	iter := session.Query(
		`SELECT path FROM documents WHERE dump_id = ?`,
		r.dumpID,
	).Iter()

	var path string
	for iter.Scan(&path) {
		if strings.HasPrefix(path, prefix) {
			px = append(px, path)
		}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return px, nil
}

func (r *reader) ReadDocument(ctx context.Context, path string) (types.DocumentData, bool, error) {
	var data string
	if err := session.Query(
		`SELECT data FROM documents WHERE dump_id = ? AND path = ? LIMIT 1`,
		r.dumpID,
		path,
	).Scan(&data); err != nil {
		if err == gocql.ErrNotFound {
			err = nil
		}
		return types.DocumentData{}, false, err
	}

	documentData, err := r.serializer.UnmarshalDocumentData([]byte(data))
	if err != nil {
		return types.DocumentData{}, false, err
	}

	return documentData, true, nil
}

func (r *reader) ReadResultChunk(ctx context.Context, id int) (types.ResultChunkData, bool, error) {
	var data string
	if err := session.Query(
		`SELECT data FROM result_chunks WHERE dump_id = ? AND idx = ? LIMIT 1`,
		r.dumpID,
		id,
	).Scan(&data); err != nil {
		if err == gocql.ErrNotFound {
			err = nil
		}
		return types.ResultChunkData{}, false, err
	}

	resultChunkData, err := r.serializer.UnmarshalResultChunkData([]byte(data))
	if err != nil {
		return types.ResultChunkData{}, false, err
	}

	return resultChunkData, true, nil
}

func (r *reader) ReadDefinitions(ctx context.Context, scheme, identifier string, skip, take int) ([]types.Location, int, error) {
	return r.defref(ctx, "definitions", scheme, identifier, skip, take)
}

func (r *reader) ReadReferences(ctx context.Context, scheme, identifier string, skip, take int) ([]types.Location, int, error) {
	return r.defref(ctx, "references", scheme, identifier, skip, take)
}

func (r *reader) defref(ctx context.Context, tableName, scheme, identifier string, skip, take int) ([]types.Location, int, error) {
	locations, err := r.readDefinitionReferences(ctx, tableName, scheme, identifier)
	if err != nil {
		return nil, 0, err
	}

	if skip == 0 && take == 0 {
		// Pagination is disabled, return full result set
		return locations, len(locations), nil
	}

	lo := skip
	if lo >= len(locations) {
		// Skip lands past result set, return nothing
		return nil, len(locations), nil
	}

	hi := skip + take
	if hi >= len(locations) {
		hi = len(locations)
	}

	return locations[lo:hi], len(locations), nil
}

func (r *reader) readDefinitionReferences(ctx context.Context, tableName, scheme, identifier string) ([]types.Location, error) {
	iter := session.Query(
		fmt.Sprintf(`SELECT data FROM %s WHERE dump_id = ? AND scheme = ? AND identifier = ?`, tableName),
		r.dumpID,
		scheme,
		identifier,
	).Iter()

	var data string
	var allLocations []types.Location
	for iter.Scan(&data) {
		locations, err := r.serializer.UnmarshalLocations([]byte(data))
		if err != nil {
			return nil, err
		}

		allLocations = append(allLocations, locations...)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return allLocations, nil
}

func (r *reader) Close() error {
	return nil
}
