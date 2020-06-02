package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	pkgerrors "github.com/pkg/errors"
	persistence "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	jsonserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization/json"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/migrate"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

// ErrNoMetadata occurs when there are no rows in the meta table.
var ErrNoMetadata = errors.New("no rows in meta table")

type sqliteReader struct {
	store      *store.Store
	closer     func() error
	serializer serialization.Serializer
}

var _ persistence.Reader = &sqliteReader{}

func NewReader(ctx context.Context, filename string) (_ persistence.Reader, err error) {
	store, closer, err := store.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if closeErr := closer(); closeErr != nil {
				err = multierror.Append(err, closeErr)
			}
		}
	}()

	serializer := jsonserializer.New()

	if err := migrate.Migrate(ctx, store, serializer); err != nil {
		return nil, err
	}

	return &sqliteReader{
		store:      store,
		closer:     closer,
		serializer: serializer,
	}, nil
}

func (r *sqliteReader) ReadMeta(ctx context.Context) (types.MetaData, error) {
	numResultChunks, exists, err := store.ScanFirstInt(r.store.Query(ctx, sqlf.Sprintf(
		`SELECT numResultChunks FROM meta LIMIT 1`,
	)))
	if err != nil {
		return types.MetaData{}, err
	}
	if !exists {
		return types.MetaData{}, ErrNoMetadata
	}

	return types.MetaData{
		NumResultChunks: numResultChunks,
	}, nil
}

func (r *sqliteReader) ReadDocument(ctx context.Context, path string) (types.DocumentData, bool, error) {
	data, exists, err := store.ScanFirstBytes(r.store.Query(ctx, sqlf.Sprintf(
		`SELECT data FROM documents WHERE path = %s LIMIT 1`,
		path,
	)))
	if err != nil || !exists {
		return types.DocumentData{}, false, err
	}

	documentData, err := r.serializer.UnmarshalDocumentData(data)
	if err != nil {
		return types.DocumentData{}, false, pkgerrors.Wrap(err, "serializer.UnmarshalDocumentData")
	}
	return documentData, true, nil
}

func (r *sqliteReader) ReadResultChunk(ctx context.Context, id int) (types.ResultChunkData, bool, error) {
	data, exists, err := store.ScanFirstBytes(r.store.Query(ctx, sqlf.Sprintf(
		`SELECT data FROM resultChunks WHERE id = %s LIMIT 1`,
		id,
	)))
	if err != nil || !exists {
		return types.ResultChunkData{}, false, err
	}

	resultChunkData, err := r.serializer.UnmarshalResultChunkData(data)
	if err != nil {
		return types.ResultChunkData{}, false, pkgerrors.Wrap(err, "serializer.UnmarshalResultChunkData")
	}
	return resultChunkData, true, nil
}

func (r *sqliteReader) ReadDefinitions(ctx context.Context, scheme, identifier string, skip, take int) ([]types.Location, int, error) {
	return r.readDefinitionReferences(ctx, "definitions", scheme, identifier, skip, take)
}

func (r *sqliteReader) ReadReferences(ctx context.Context, scheme, identifier string, skip, take int) ([]types.Location, int, error) {
	return r.readDefinitionReferences(ctx, "references", scheme, identifier, skip, take)
}

func (r *sqliteReader) readDefinitionReferences(ctx context.Context, tableName, scheme, identifier string, skip, take int) ([]types.Location, int, error) {
	var limitOffset string
	if take != 0 && skip != 0 {
		limitOffset = fmt.Sprintf("LIMIT %d OFFSET %d", take, skip)
	}

	locations, err := scanLocations(r.store.Query(ctx, sqlf.Sprintf(
		`SELECT documentPath, startLine, startCharacter, endLine, endCharacter FROM "`+tableName+`" WHERE scheme = %s AND identifier = %s `+limitOffset,
		scheme,
		identifier,
	)))
	if err != nil {
		return nil, 0, err
	}

	count, _, err := store.ScanFirstInt(r.store.Query(ctx, sqlf.Sprintf(
		`SELECT COUNT(*) FROM "`+tableName+`" WHERE scheme = %s AND identifier = %s `,
		scheme,
		identifier,
	)))
	if err != nil {
		return nil, 0, err
	}

	return locations, count, err
}

func (r *sqliteReader) Close() error {
	return r.closer()
}

// scanLocations reads the given set of definition/reference rows and returns a slice of resulting
// values. This method should be called directly with the return value of `*db.query`.
func scanLocations(rows *sql.Rows, queryErr error) (_ []types.Location, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = store.CloseRows(rows, err) }()

	var locations []types.Location
	for rows.Next() {
		var location types.Location
		if err := rows.Scan(
			&location.URI,
			&location.StartLine,
			&location.StartCharacter,
			&location.EndLine,
			&location.EndCharacter,
		); err != nil {
			return nil, err
		}

		locations = append(locations, location)
	}

	return locations, nil
}
