package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/keegancsmith/sqlf"
	pkgerrors "github.com/pkg/errors"
	persistence "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization"
	jsonserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/serialization/json"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
)

var ErrNoMetadata = errors.New("no rows in meta table")

type sqliteReader struct {
	db         *sqlx.DB
	serializer serialization.Serializer
}

var _ persistence.Reader = &sqliteReader{}

func NewReader(filename string) (_ persistence.Reader, err error) {
	db, err := sqlx.Open("sqlite3_with_pcre", filename)
	if err != nil {
		return nil, err
	}

	return &sqliteReader{
		db:         db,
		serializer: jsonserializer.New(),
	}, nil
}

func (r *sqliteReader) ReadMeta(ctx context.Context) (lsifVersion string, sourcegraphVersion string, numResultChunks int, _ error) {
	query := `SELECT lsifVersion, sourcegraphVersion, numResultChunks FROM meta LIMIT 1`

	if err := r.queryRow(ctx, sqlf.Sprintf(query)).Scan(&lsifVersion, &sourcegraphVersion, &numResultChunks); err != nil {
		if err == sql.ErrNoRows {
			return "", "", 0, ErrNoMetadata
		}

		return "", "", 0, err
	}

	return lsifVersion, sourcegraphVersion, numResultChunks, nil
}

func (r *sqliteReader) ReadDocument(ctx context.Context, path string) (types.DocumentData, bool, error) {
	query := `SELECT data FROM documents WHERE path = %s LIMIT 1`

	data, err := scanBytes(r.queryRow(ctx, sqlf.Sprintf(query, path)))
	if err != nil {
		if err == sql.ErrNoRows {
			return types.DocumentData{}, false, nil
		}
	}

	documentData, err := r.serializer.UnmarshalDocumentData(data)
	if err != nil {
		return types.DocumentData{}, false, pkgerrors.Wrap(err, "serializer.UnmarshalDocumentData")
	}
	return documentData, true, nil
}

func (r *sqliteReader) ReadResultChunk(ctx context.Context, id int) (types.ResultChunkData, bool, error) {
	query := `SELECT data FROM resultChunks WHERE id = %s LIMIT 1`

	data, err := scanBytes(r.queryRow(ctx, sqlf.Sprintf(query, id)))
	if err != nil {
		if err == sql.ErrNoRows {
			return types.ResultChunkData{}, false, nil
		}
	}

	resultChunkData, err := r.serializer.UnmarshalResultChunkData(data)
	if err != nil {
		return types.ResultChunkData{}, false, pkgerrors.Wrap(err, "serializer.UnmarshalResultChunkData")
	}
	return resultChunkData, true, nil
}

func (r *sqliteReader) ReadDefinitions(ctx context.Context, scheme, identifier string, skip, take int) ([]types.DefinitionReferenceRow, int, error) {
	var query *sqlf.Query
	if take == 0 && skip == 0 {
		query = sqlf.Sprintf(`
			SELECT `+strings.Join(definitionReferenceColumns, ", ")+`
			FROM definitions
			WHERE scheme = %s AND identifier = %s
		`, scheme, identifier)
	} else {
		query = sqlf.Sprintf(`
			SELECT `+strings.Join(definitionReferenceColumns, ", ")+`
			FROM definitions
			WHERE scheme = %s AND identifier = %s
			LIMIT %d OFFSET %d
		`, scheme, identifier, take, skip)
	}

	rows, err := scanDefinitionReferenceRows(r.query(ctx, query))
	if err != nil {
		return nil, 0, err
	}

	countQuery := `
		SELECT COUNT(*) FROM definitions
		WHERE scheme = %s AND identifier = %s
	`

	count, err := scanInt(r.queryRow(ctx, sqlf.Sprintf(countQuery, scheme, identifier)))
	if err != nil {
		return nil, 0, err
	}

	return rows, count, err
}

func (r *sqliteReader) ReadReferences(ctx context.Context, scheme, identifier string, skip, take int) ([]types.DefinitionReferenceRow, int, error) {
	var query *sqlf.Query
	if take == 0 && skip == 0 {
		query = sqlf.Sprintf(`
			SELECT `+strings.Join(definitionReferenceColumns, ", ")+`
			FROM "references"
			WHERE scheme = %s AND identifier = %s
		`, scheme, identifier)
	} else {
		query = sqlf.Sprintf(`
			SELECT `+strings.Join(definitionReferenceColumns, ", ")+`
			FROM "references"
			WHERE scheme = %s AND identifier = %s
			LIMIT %s OFFSET %d
		`, scheme, identifier, take, skip)
	}

	rows, err := scanDefinitionReferenceRows(r.query(ctx, query))
	if err != nil {
		return nil, 0, err
	}

	countQuery := `
		SELECT COUNT(*) FROM "references"
		WHERE scheme = %s AND identifier = %s
	`

	count, err := scanInt(r.queryRow(ctx, sqlf.Sprintf(countQuery, scheme, identifier)))
	if err != nil {
		return nil, 0, err
	}

	return rows, count, err
}

func (r *sqliteReader) Close() error {
	return r.db.Close()
}

var definitionReferenceColumns = []string{
	"scheme",
	"identifier",
	"documentPath",
	"startLine",
	"startCharacter",
	"endLine",
	"endCharacter",
}

// query performs QueryContext on the underlying connection.
func (r *sqliteReader) query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return r.db.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}

// queryRow performs QueryRowContext on the underlying connection.
func (r *sqliteReader) queryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return r.db.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
}

// scanBytes populates a byte slice value from the given scanner.
func scanBytes(scanner *sql.Row) (value []byte, err error) {
	err = scanner.Scan(&value)
	return value, err
}

// scanInt populates an integer value from the given scanner.
func scanInt(scanner *sql.Row) (value int, err error) {
	err = scanner.Scan(&value)
	return value, err
}

// scanDefinitionReferenceRow populates a DefinitionReferenceRow value from the given scanner.
func scanDefinitionReferenceRow(rows *sql.Rows) (row types.DefinitionReferenceRow, err error) {
	err = rows.Scan(
		&row.Scheme,
		&row.Identifier,
		&row.URI,
		&row.StartLine,
		&row.StartCharacter,
		&row.EndLine,
		&row.EndCharacter,
	)
	return row, err
}

// scanDefinitionReferenceRows reads the given set of definition/reference rows and returns
// a slice of resulting values. This method should be called directly with the return value
// of `*db.query`.
func scanDefinitionReferenceRows(rows *sql.Rows, err error) ([]types.DefinitionReferenceRow, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dumps []types.DefinitionReferenceRow
	for rows.Next() {
		dump, err := scanDefinitionReferenceRow(rows)
		if err != nil {
			return nil, err
		}

		dumps = append(dumps, dump)
	}

	return dumps, nil
}
