package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var ErrNoMetadata = errors.New("no rows in meta table")

func (s *Store) ReadMeta(ctx context.Context, bundleID int) (_ MetaData, err error) {
	ctx, endObservation := s.operations.readMeta.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	numResultChunks, ok, err := basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(readMetaQuery, bundleID)))
	if err != nil {
		return MetaData{}, err
	}
	if !ok {
		return MetaData{}, ErrNoMetadata
	}

	return MetaData{NumResultChunks: numResultChunks}, nil
}

const readMetaQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_read.go:ReadMeta
SELECT num_result_chunks FROM lsif_data_metadata WHERE dump_id = %s
`

func (s *Store) PathsWithPrefix(ctx context.Context, bundleID int, prefix string) (_ []string, err error) {
	ctx, endObservation := s.operations.pathsWithPrefix.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("prefix", prefix),
	}})
	defer endObservation(1, observation.Args{})

	paths, err := basestore.ScanStrings(s.Store.Query(ctx, sqlf.Sprintf(pathsWithPrefixQuery, bundleID, prefix+"%")))
	if err != nil {
		return nil, err
	}

	return paths, nil
}

const pathsWithPrefixQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_read.go:PathsWithPrefix
SELECT path FROM lsif_data_documents WHERE dump_id = %s AND path LIKE %s ORDER BY path
`

func (s *Store) ReadDocument(ctx context.Context, bundleID int, path string) (_ DocumentData, _ bool, err error) {
	ctx, endObservation := s.operations.readDocument.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

	data, ok, err := basestore.ScanFirstString(s.Store.Query(ctx, sqlf.Sprintf(readDocumentQuery, bundleID, path)))
	if err != nil || !ok {
		return DocumentData{}, false, err
	}

	documentData, err := s.serializer.UnmarshalDocumentData([]byte(data))
	if err != nil {
		return DocumentData{}, false, err
	}

	return documentData, true, nil
}

const readDocumentQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_read.go:ReadDocument
SELECT data FROM lsif_data_documents WHERE dump_id = %s AND path = %s LIMIT 1
`

func (s *Store) ReadResultChunk(ctx context.Context, bundleID int, id int) (_ ResultChunkData, _ bool, err error) {
	ctx, endObservation := s.operations.readResultChunk.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	data, ok, err := basestore.ScanFirstString(s.Store.Query(ctx, sqlf.Sprintf(readResultChunkQuery, bundleID, id)))
	if err != nil || !ok {
		return ResultChunkData{}, false, err
	}

	resultChunkData, err := s.serializer.UnmarshalResultChunkData([]byte(data))
	if err != nil {
		return ResultChunkData{}, false, err
	}

	return resultChunkData, true, nil
}

const readResultChunkQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_read.go:ReadResultChunk
SELECT data FROM lsif_data_result_chunks WHERE dump_id = %s AND idx = %s LIMIT 1
`

func (s *Store) ReadDefinitions(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) (_ []LocationData, _ int, err error) {
	ctx, endObservation := s.operations.readDefinitions.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("scheme", scheme),
		log.String("identifier", identifier),
		log.Int("skip", skip),
		log.Int("take", take),
	}})
	defer endObservation(1, observation.Args{})

	return s.readDefinitionReferences(ctx, bundleID, "lsif_data_definitions", scheme, identifier, skip, take)
}

func (s *Store) ReadReferences(ctx context.Context, bundleID int, scheme, identifier string, skip, take int) (_ []LocationData, _ int, err error) {
	ctx, endObservation := s.operations.readReferences.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("scheme", scheme),
		log.String("identifier", identifier),
		log.Int("skip", skip),
		log.Int("take", take),
	}})
	defer endObservation(1, observation.Args{})

	return s.readDefinitionReferences(ctx, bundleID, "lsif_data_references", scheme, identifier, skip, take)
}

func (s *Store) readDefinitionReferences(ctx context.Context, bundleID int, tableName, scheme, identifier string, skip, take int) (_ []LocationData, _ int, err error) {
	data, ok, err := basestore.ScanFirstString(s.Store.Query(ctx, sqlf.Sprintf(readDefinitionReferencesQuery, sqlf.Sprintf(tableName), bundleID, scheme, identifier)))
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

const readDefinitionReferencesQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_read.go:readDefinitionReferences
SELECT data FROM %s WHERE dump_id = %s AND scheme = %s AND identifier = %s LIMIT 1
`
