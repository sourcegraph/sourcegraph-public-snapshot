package sqlite

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Database interface {
	Close() error
	Transact(ctx context.Context) (Database, error)
	Done(err error) error

	Search(ctx context.Context, args types.SearchArgs) ([]result.Symbol, error)

	getCommit(ctx context.Context) (string, bool, error)
	createSchema(ctx context.Context) error
	insertMeta(ctx context.Context, commitID string) error
	createIndexes(ctx context.Context) error
	updateMeta(ctx context.Context, commitID string) error
	deletePaths(ctx context.Context, paths []string) error
	writeSymbols(ctx context.Context, symbols <-chan result.Symbol) error
}

type database struct {
	db *sqlx.DB
	*basestore.Store
}

func NewDatabase(dbFile string) (Database, error) {
	db, err := sqlx.Open("sqlite3_with_regexp", dbFile)
	if err != nil {
		return nil, err
	}

	return &database{
		db:    db,
		Store: basestore.NewWithDB(db, sql.TxOptions{}),
	}, nil
}

func WithDatabase(dbFile string, callback func(db Database) error) error {
	db, err := NewDatabase(dbFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			// TODO
		}
	}()

	return callback(db)
}

func WithTransaction(ctx context.Context, dbFile string, callback func(db Database) error) error {
	return WithDatabase(dbFile, func(db Database) (err error) {
		tx, err := db.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		return callback(tx)
	})
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) Transact(ctx context.Context) (Database, error) {
	tx, err := d.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &database{db: d.db, Store: tx}, nil
}

func (w *database) Search(ctx context.Context, args types.SearchArgs) (res []result.Symbol, err error) {
	var conditions []*sqlf.Query
	conditions = append(conditions, makeCondition("name", args.Query, args.IsCaseSensitive)...)
	for _, includePattern := range args.IncludePatterns {
		conditions = append(conditions, makeCondition("path", includePattern, args.IsCaseSensitive)...)
	}
	conditions = append(conditions, negateAll(makeCondition("path", args.ExcludePattern, args.IsCaseSensitive))...)

	if len(conditions) == 0 {
		conditions = append(conditions, sqlf.Sprintf("TRUE"))
	}

	sqlQuery := sqlf.Sprintf(searchQuery, sqlf.Join(conditions, "AND"), args.First)

	rows, err := w.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var symbol result.Symbol
		if err := rows.Scan(
			&symbol.Name,
			&symbol.Path,
			&symbol.Line,
			&symbol.Kind,
			&symbol.Language,
			&symbol.Parent,
			&symbol.ParentKind,
			&symbol.Signature,
			&symbol.Pattern,
			&symbol.FileLimited,
		); err != nil {
			return nil, err
		}

		res = append(res, symbol)
	}

	return res, nil
}

const searchQuery = `
SELECT
	name,
	path,
	line,
	kind,
	language,
	parent,
	parentkind,
	signature,
	pattern,
	filelimited
FROM symbols
WHERE %s
LIMIT %s
`

func (w *database) getCommit(ctx context.Context) (string, bool, error) {
	return basestore.ScanFirstString(w.Query(ctx, sqlf.Sprintf(`SELECT revision FROM meta`)))
}

func (w *database) createSchema(ctx context.Context) error {
	for _, createTableQuery := range []string{createMetaTableQuery, createSymbolsTableQuery} {
		if err := w.Exec(ctx, sqlf.Sprintf(createTableQuery)); err != nil {
			return err
		}
	}

	return nil
}

const createMetaTableQuery = `
CREATE TABLE IF NOT EXISTS meta (
	id INTEGER PRIMARY KEY CHECK (id = 0),
	revision TEXT NOT NULL
)
`

// The column names are the lowercase version of fields in `symbolInDB`
// because sqlx lowercases struct fields by default. See
// http://jmoiron.github.io/sqlx/#query
const createSymbolsTableQuery = `
CREATE TABLE IF NOT EXISTS symbols (
	name VARCHAR(256) NOT NULL,
	namelowercase VARCHAR(256) NOT NULL,
	path VARCHAR(4096) NOT NULL,
	pathlowercase VARCHAR(4096) NOT NULL,
	line INT NOT NULL,
	kind VARCHAR(255) NOT NULL,
	language VARCHAR(255) NOT NULL,
	parent VARCHAR(255) NOT NULL,
	parentkind VARCHAR(255) NOT NULL,
	signature VARCHAR(255) NOT NULL,
	pattern VARCHAR(255) NOT NULL,
	filelimited BOOLEAN NOT NULL
)
`

func (w *database) insertMeta(ctx context.Context, commitID string) error {
	return w.Exec(ctx, sqlf.Sprintf(`INSERT INTO meta (id, revision) VALUES (0, %s)`, commitID))
}

func (w *database) createIndexes(ctx context.Context) error {
	createIndexStatements := []string{
		`CREATE INDEX name_index ON symbols(name)`,
		`CREATE INDEX path_index ON symbols(path)`,
		`CREATE INDEX namelowercase_index ON symbols(namelowercase)`,
		`CREATE INDEX pathlowercase_index ON symbols(pathlowercase)`,
	}

	for _, createIndex := range createIndexStatements {
		if err := w.Exec(ctx, sqlf.Sprintf(createIndex)); err != nil {
			return err
		}
	}

	return nil
}

func (w *database) updateMeta(ctx context.Context, commitID string) error {
	return w.Exec(ctx, sqlf.Sprintf(`UPDATE meta SET revision = %s`, commitID))
}

func (w *database) deletePaths(ctx context.Context, paths []string) error {
	return w.Exec(ctx, sqlf.Sprintf(`DELETE FROM symbols WHERE path = ANY(%s)`, pq.Array(paths)))
}

var symbolsColumnNames = []string{
	"name",
	"namelowercase",
	"path",
	"pathlowercase",
	"line",
	"kind",
	"language",
	"parent",
	"parentkind",
	"signature",
	"pattern",
	"filelimited",
}

func (w *database) writeSymbols(ctx context.Context, symbols <-chan result.Symbol) (err error) {
	rows := make(chan []interface{})

	go func() {
		defer close(rows)

		for symbol := range symbols {
			rows <- []interface{}{
				symbol.Name,
				strings.ToLower(symbol.Name),
				symbol.Path,
				strings.ToLower(symbol.Path),
				symbol.Line,
				symbol.Kind,
				symbol.Language,
				symbol.Parent,
				symbol.ParentKind,
				symbol.Signature,
				symbol.Pattern,
				symbol.FileLimited,
			}
		}
	}()

	return batch.InsertValues(ctx, w.Handle().DB(), "symbols", symbolsColumnNames, rows)
}
