package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization"
	gobserializer "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/serialization/gob"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

type reader struct {
	dumpID     int
	serializer serialization.Serializer
	writer     *batchWriter
}

var _ persistence.Store = &reader{}

func NewStore(dumpID int) persistence.Store {
	return &reader{
		dumpID:     dumpID,
		serializer: gobserializer.New(),
		writer:     newBatchWriter(),
	}
}

// TODO
func (r *reader) Transact(ctx context.Context) (persistence.Store, error) { return r, nil }
func (r *reader) Done(err error) error                                    { return err }
func (r *reader) CreateTables(ctx context.Context) error                  { return nil }

func (r *reader) ReadMeta(ctx context.Context) (_ types.MetaData, err error) {
	rows, err := dbconn.Global.Query(
		`SELECT num_result_chunks FROM lsif_data_metadata WHERE dump_id = $1 LIMIT 1`,
		r.dumpID,
	)
	if err != nil {
		return types.MetaData{}, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	if !rows.Next() {
		return types.MetaData{}, fmt.Errorf("missing metadata")
	}

	var numResultChunks int
	if rows.Scan(&numResultChunks); err != nil {
		return types.MetaData{}, err
	}

	if err := rows.Err(); err != nil {
		return types.MetaData{}, err
	}

	return types.MetaData{NumResultChunks: numResultChunks}, nil
}

func (r *reader) PathsWithPrefix(ctx context.Context, prefix string) (px []string, err error) {
	rows, err := dbconn.Global.Query(
		`SELECT path FROM lsif_data_documents WHERE dump_id = $1`,
		r.dumpID,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}

		if strings.HasPrefix(path, prefix) {
			px = append(px, path)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return px, nil
}

func (r *reader) ReadDocument(ctx context.Context, path string) (types.DocumentData, bool, error) {
	rows, err := dbconn.Global.Query(
		`SELECT data FROM lsif_data_documents WHERE dump_id = $1 AND path = $2 LIMIT 1`,
		r.dumpID,
		path,
	)
	if err != nil {
		return types.DocumentData{}, false, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	if !rows.Next() {
		return types.DocumentData{}, false, nil
	}

	var data string
	if err := rows.Scan(&data); err != nil {
		return types.DocumentData{}, false, err
	}

	if err := rows.Err(); err != nil {
		return types.DocumentData{}, false, err
	}

	documentData, err := r.serializer.UnmarshalDocumentData([]byte(data))
	if err != nil {
		return types.DocumentData{}, false, err
	}

	return documentData, true, nil
}

func (r *reader) ReadResultChunk(ctx context.Context, id int) (types.ResultChunkData, bool, error) {
	rows, err := dbconn.Global.Query(
		`SELECT data FROM lsif_data_result_chunks WHERE dump_id = $1 AND idx = $2 LIMIT 1`,
		r.dumpID,
		id,
	)
	if err != nil {
		return types.ResultChunkData{}, false, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	if !rows.Next() {
		return types.ResultChunkData{}, false, nil
	}

	var data string
	if err := rows.Scan(&data); err != nil {
		return types.ResultChunkData{}, false, err
	}

	if err := rows.Err(); err != nil {
		return types.ResultChunkData{}, false, err
	}

	resultChunkData, err := r.serializer.UnmarshalResultChunkData([]byte(data))
	if err != nil {
		return types.ResultChunkData{}, false, err
	}

	return resultChunkData, true, nil
}

func (r *reader) ReadDefinitions(ctx context.Context, scheme, identifier string, skip, take int) ([]types.Location, int, error) {
	return r.defref(ctx, "lsif_data_definitions", scheme, identifier, skip, take)
}

func (r *reader) ReadReferences(ctx context.Context, scheme, identifier string, skip, take int) ([]types.Location, int, error) {
	return r.defref(ctx, "lsif_data_references", scheme, identifier, skip, take)
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

func (r *reader) readDefinitionReferences(ctx context.Context, tableName, scheme, identifier string) (_ []types.Location, err error) {
	rows, err := dbconn.Global.Query(
		fmt.Sprintf(`SELECT data FROM %s WHERE dump_id = $1 AND scheme = $2 AND identifier = $3`, tableName),
		r.dumpID,
		scheme,
		identifier,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	var allLocations []types.Location

	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}

		locations, err := r.serializer.UnmarshalLocations([]byte(data))
		if err != nil {
			return nil, err
		}

		allLocations = append(allLocations, locations...)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return allLocations, nil
}
