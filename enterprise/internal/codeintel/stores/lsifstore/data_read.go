package lsifstore

import (
	"context"
	"database/sql"

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

func (s *Store) PathsWithPrefix(ctx context.Context, bundleID int, prefix string) (_ []string, err error) {
	ctx, endObservation := s.operations.pathsWithPrefix.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("prefix", prefix),
	}})
	defer endObservation(1, observation.Args{})

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

func (s *Store) ReadDocument(ctx context.Context, bundleID int, path string) (_ DocumentData, _ bool, err error) {
	ctx, endObservation := s.operations.readDocument.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
	}})
	defer endObservation(1, observation.Args{})

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

func (s *Store) ReadResultChunk(ctx context.Context, bundleID int, id int) (_ ResultChunkData, _ bool, err error) {
	ctx, endObservation := s.operations.readResultChunk.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

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

// TODO(sqs): rename readDefinitionReferences to readDefinitionReferencesForMoniker
func (s *Store) readMonikerLocations(ctx context.Context, bundleID int, tableName string, skip, take int) (_ []MonikerLocations, err error) {
	scanMonikerLocations := func(rows *sql.Rows, queryErr error) (_ []MonikerLocations, err error) {
		if queryErr != nil {
			return nil, queryErr
		}
		defer func() { err = basestore.CloseRows(rows, err) }()

		var values []MonikerLocations
		for rows.Next() {
			var (
				value MonikerLocations
				data  []byte // raw locations data
			)
			if err := rows.Scan(&value.Scheme, &value.Identifier, &data); err != nil {
				return nil, err
			}

			locations, err := s.serializer.UnmarshalLocations([]byte(data))
			if err != nil {
				return nil, err
			}
			value.Locations = locations

			values = append(values, value)
		}

		return values, nil
	}

	if tableName == "definitions" {
		tableName = "lsif_data_definitions"
	} else {
		panic("TODO(sqs)")
	}

	return scanMonikerLocations(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT scheme, identifier, data FROM "`+tableName+`" WHERE dump_id = %s AND (identifier LIKE %s) LIMIT %d OFFSET %d`,
			bundleID,
			"%:%", // TODO(sqs): hack, omit "local" monikers from lsif-node with
			take,
			skip,
		),
	))
}

func (s *Store) ReadSymbols(ctx context.Context, bundleID int) (_ []SymbolData, err error) {
	ctx, endObservation := s.operations.readSymbols.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	datas, err := basestore.ScanStrings(s.Store.Query(
		ctx,
		sqlf.Sprintf(
			`SELECT data FROM lsif_data_symbols WHERE dump_id = %s`,
			bundleID,
		),
	))
	if err != nil {
		return nil, err
	}

	symbols := make([]SymbolData, len(datas))
	for i, data := range datas {
		symbols[i], err = s.serializer.UnmarshalSymbol([]byte(data))
		if err != nil {
			return nil, err
		}
	}

	return symbols, nil
}
