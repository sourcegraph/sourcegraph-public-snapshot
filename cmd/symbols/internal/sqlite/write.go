package sqlite

import (
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/jmoiron/sqlx"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
)

type DatabaseWriter interface {
	WriteDBFile(ctx context.Context, args types.SearchArgs, tempDBFile string) error
}

type databaseWriter struct {
	path            string
	gitserverClient GitserverClient
	parser          parser.Parser
}

func NewDatabaseWriter(
	path string,
	gitserverClient GitserverClient,
	parser parser.Parser,
) DatabaseWriter {
	return &databaseWriter{
		path:            path,
		gitserverClient: gitserverClient,
		parser:          parser,
	}
}

func (w *databaseWriter) WriteDBFile(ctx context.Context, args types.SearchArgs, tempDBFile string) error {
	newest, err := findNewestFile(filepath.Join(w.path, diskcache.EncodeKeyComponent(string(args.Repo))))
	if err != nil {
		return err
	}

	if newest == "" {
		// There are no existing SQLite DBs to reuse, so write a completely new one.
		err := WriteAllSymbolsToNewDB(ctx, w.parser, tempDBFile, args.Repo, args.CommitID)
		if err != nil {
			if err == context.Canceled {
				log15.Error("Unable to parse repository symbols within the context", "repo", args.Repo, "commit", args.CommitID, "query", args.Query)
			}
			return err
		}
	} else {
		// Copy the existing DB to a new DB and update the new DB
		err = copyFile(newest, tempDBFile)
		if err != nil {
			return err
		}

		err = updateSymbols(ctx, w.gitserverClient, w.parser, tempDBFile, args.Repo, args.CommitID)
		if err != nil {
			if err == context.Canceled {
				log15.Error("updateSymbols: unable to parse repository symbols within the context", "repo", args.Repo, "commit", args.CommitID, "query", args.Query)
			}
			return err
		}
	}

	return nil
}

// WriteAllSymbolsToNewDB fetches the repo@commit from gitserver, parses all the
// symbols, and writes them to the blank database file `dbFile`.
func WriteAllSymbolsToNewDB(ctx context.Context, parser parser.Parser, dbFile string, repoName api.RepoName, commitID api.CommitID) (err error) {
	db, err := sqlx.Open("sqlite3_with_regexp", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	store := basestore.NewWithDB(db, sql.TxOptions{})

	if err := createSchema(ctx, commitID, store); err != nil {
		return err
	}

	if err := writeSymbols(ctx, parser, repoName, commitID, nil, store); err != nil {
		return err
	}

	if err := createIndexes(ctx, commitID, store); err != nil {
		return err
	}

	return nil
}

// updateSymbols adds/removes rows from the DB based on a `git diff` between the meta.revision within the
// DB and the given commitID.
func updateSymbols(ctx context.Context, gitserverClient GitserverClient, parser parser.Parser, dbFile string, repoName api.RepoName, commitID api.CommitID) (err error) {
	db, err := sqlx.Open("sqlite3_with_regexp", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	store := basestore.NewWithDB(db, sql.TxOptions{})

	// Read old commit
	metaQuery := `SELECT revision FROM meta`
	oldCommit, _, err := basestore.ScanFirstString(store.Query(ctx, sqlf.Sprintf(metaQuery)))
	if err != nil {
		return err
	}

	// git diff
	changes, err := gitserverClient.GitDiff(ctx, repoName, api.CommitID(oldCommit), commitID)
	if err != nil {
		return err
	}

	tx, err := store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Write new commit
	updateMetaQuery := `UPDATE meta SET revision = %s`
	if err := tx.Exec(ctx, sqlf.Sprintf(updateMetaQuery, commitID)); err != nil {
		return err
	}

	paths := append(changes.Deleted, changes.Modified...)

	deleteQuery := "DELETE FROM symbols WHERE path = ANY(%s)"
	if err := tx.Exec(ctx, sqlf.Sprintf(deleteQuery, pq.Array(paths))); err != nil {
		return err
	}

	return writeSymbols(ctx, parser, repoName, commitID, paths, tx)
}

func createSchema(ctx context.Context, commitID api.CommitID, tx *basestore.Store) error {
	createMetaTableQuery := `
		CREATE TABLE IF NOT EXISTS meta (
		id INTEGER PRIMARY KEY CHECK (id = 0),
		revision TEXT NOT NULL
	)`
	if err := tx.Exec(ctx, sqlf.Sprintf(createMetaTableQuery)); err != nil {
		return err
	}

	// The column names are the lowercase version of fields in `symbolInDB`
	// because sqlx lowercases struct fields by default. See
	// http://jmoiron.github.io/sqlx/#query
	createSymbolsTableQuery := `
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
	if err := tx.Exec(
		ctx,
		sqlf.Sprintf(createSymbolsTableQuery)); err != nil {
		return err
	}

	insertMetaRowQuery := `
		INSERT INTO meta (id, revision) VALUES (0, %s)
	`
	if err := tx.Exec(ctx, sqlf.Sprintf(insertMetaRowQuery, string(commitID))); err != nil {
		return err
	}

	return nil
}

func createIndexes(ctx context.Context, commitID api.CommitID, tx *basestore.Store) error {
	createIndexQuery1 := `CREATE INDEX name_index ON symbols(name);`
	if err := tx.Exec(ctx, sqlf.Sprintf(createIndexQuery1)); err != nil {
		return err
	}

	createIndexQuery2 := `CREATE INDEX path_index ON symbols(path);`
	if err := tx.Exec(ctx, sqlf.Sprintf(createIndexQuery2)); err != nil {
		return err
	}

	// `*lowercase_index` enables indexed case insensitive queries.
	createIndexQuery3 := `CREATE INDEX namelowercase_index ON symbols(namelowercase);`
	if err := tx.Exec(ctx, sqlf.Sprintf(createIndexQuery3)); err != nil {
		return err
	}

	createIndexQuery4 := `CREATE INDEX pathlowercase_index ON symbols(pathlowercase);`
	if err := tx.Exec(ctx, sqlf.Sprintf(createIndexQuery4)); err != nil {
		return err
	}

	return nil
}

func writeSymbols(ctx context.Context, parser parser.Parser, repoName api.RepoName, commitID api.CommitID, paths []string, store *basestore.Store) (err error) {
	tx, err := store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	symbols, err := parser.Parse(ctx, repoName, commitID, paths)
	if err != nil {
		return err
	}

	for symbol := range symbols {
		symbolInDBValue := types.SymbolToSymbolInDB(symbol)

		// TODO - use bulk loader instead
		insertQuery := `
			INSERT INTO symbols (
				name,
				namelowercase,
				path,
				pathlowercase,
				line,
				kind,
				language,
				parent,
				parentkind,
				signature,
				pattern,
				filelimited
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`
		if err := tx.Exec(
			ctx,
			sqlf.Sprintf(
				insertQuery,
				symbolInDBValue.Name,
				symbolInDBValue.NameLowercase,
				symbolInDBValue.Path,
				symbolInDBValue.PathLowercase,
				symbolInDBValue.Line,
				symbolInDBValue.Kind,
				symbolInDBValue.Language,
				symbolInDBValue.Parent,
				symbolInDBValue.ParentKind,
				symbolInDBValue.Signature,
				symbolInDBValue.Parent,
				symbolInDBValue.FileLimited,
			),
		); err != nil {
			return err
		}
	}

	return nil
}

// findNewestFile lists the directory and returns the newest file's path, prepended with dir.
func findNewestFile(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", nil
	}

	var mostRecentTime time.Time
	newest := ""
	for _, fi := range files {
		if fi.Type().IsRegular() {
			if !strings.HasSuffix(fi.Name(), ".zip") {
				continue
			}

			info, err := fi.Info()
			if err != nil {
				return "", err
			}

			if newest == "" || info.ModTime().After(mostRecentTime) {
				mostRecentTime = info.ModTime()
				newest = filepath.Join(dir, fi.Name())
			}
		}
	}

	return newest, nil
}

// copyFile is like the cp command.
func copyFile(from string, to string) error {
	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	toFile, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	return err
}
