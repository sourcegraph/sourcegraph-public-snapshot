package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
)

var ErrNoMetadata = errors.New("no rows in meta table")

func (s *Store) ReadMeta(ctx context.Context, bundleID int) (MetaData, error) {
	numResultChunks, ok, err := basestore.ScanFirstInt(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT num_result_chunks FROM lsif_data_metadata WHERE dump_id = %s`,
			bundleID,
		),
	))
	if err != nil {
		return MetaData{}, err
	}
	if !ok {
		return MetaData{}, ErrNoMetadata
	}

	return MetaData{NumResultChunks: numResultChunks}, nil
}

func (s *Store) PathsWithPrefix(ctx context.Context, bundleID int, prefix string) ([]string, error) {
	paths, err := basestore.ScanStrings(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT path FROM lsif_data_documents WHERE dump_id = %s AND path LIKE %s ORDER BY path`,
			bundleID,
			prefix+"%",
		),
	))
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (s *Store) ReadDocument(ctx context.Context, bundleID int, path string) (DocumentData, bool, error) {
	data, ok, err := basestore.ScanFirstString(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT data FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1`,
			bundleID,
			path,
		),
	))
	if err != nil || !ok {
		return DocumentData{}, false, err
	}

	documentData, err := s.serializer.UnmarshalDocumentData([]byte(data))
	if err != nil {
		return DocumentData{}, false, err
	}

	return documentData, true, nil
}

func (s *Store) ReadResultChunk(ctx context.Context, bundleID int, id int) (ResultChunkData, bool, error) {
	data, ok, err := basestore.ScanFirstString(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT data FROM lsif_data_result_chunks WHERE dump_id = %s AND idx = %s LIMIT 1`,
			bundleID,
			id,
		),
	))
	if err != nil || !ok {
		return ResultChunkData{}, false, err
	}

	resultChunkData, err := s.serializer.UnmarshalResultChunkData([]byte(data))
	if err != nil {
		return ResultChunkData{}, false, err
	}

	return resultChunkData, true, nil
}

func (s *Store) ReadDefinitions(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) ([]LocationData, int, error) {
	return s.readDefinitionReferences(ctx, bundleID, "lsif_data_definitions", scheme, identifier, skip, take)
}

func (s *Store) ReadReferences(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) ([]LocationData, int, error) {
	return s.readDefinitionReferences(ctx, bundleID, "lsif_data_references", scheme, identifier, skip, take)
}

func (s *Store) readDefinitionReferences(ctx context.Context, bundleID int, tableName, scheme, identifier string, skip, take int) ([]LocationData, int, error) {
	data, ok, err := basestore.ScanFirstString(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT data FROM "`+tableName+`" WHERE dump_id = %s AND scheme = %s AND identifier = %s LIMIT 1`,
			bundleID,
			scheme,
			identifier,
		),
	))
	if err != nil || !ok {
		return nil, 0, err
	}

	locations, err := s.serializer.UnmarshalLocations([]byte(data))
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
