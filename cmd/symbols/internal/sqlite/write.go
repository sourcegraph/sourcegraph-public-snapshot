package sqlite

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/jmoiron/sqlx"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type DatabaseWriter interface {
	WriteDBFile(ctx context.Context, args types.SearchArgs, tempDBFile string) error
}

type databaseWriter struct {
	gitserverClient GitserverClient
	parser          parser.Parser
	cache           *diskcache.Store
}

func NewDatabaseWriter(
	gitserverClient GitserverClient,
	parser parser.Parser,
	cache *diskcache.Store,
) DatabaseWriter {
	return &databaseWriter{
		gitserverClient: gitserverClient,
		parser:          parser,
		cache:           cache,
	}
}

func (w *databaseWriter) WriteDBFile(ctx context.Context, args types.SearchArgs, tempDBFile string) error {
	newest, err := findNewestFile(filepath.Join(w.cache.Dir, diskcache.EncodeKeyComponent(string(args.Repo))))
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

	// Writing a bunch of rows into sqlite3 is much faster in a transaction.
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS meta (
    		id INTEGER PRIMARY KEY CHECK (id = 0),
			revision TEXT NOT NULL
		)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO meta (id, revision) VALUES (0, ?)`,
		string(commitID))
	if err != nil {
		return err
	}

	// The column names are the lowercase version of fields in `symbolInDB`
	// because sqlx lowercases struct fields by default. See
	// http://jmoiron.github.io/sqlx/#query
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS symbols (
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
		)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX name_index ON symbols(name);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX path_index ON symbols(path);`)
	if err != nil {
		return err
	}

	// `*lowercase_index` enables indexed case insensitive queries.
	_, err = tx.Exec(`CREATE INDEX namelowercase_index ON symbols(namelowercase);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX pathlowercase_index ON symbols(pathlowercase);`)
	if err != nil {
		return err
	}

	insertStatement, err := tx.PrepareNamed(insertQuery)
	if err != nil {
		return err
	}

	return parser.Parse(ctx, repoName, commitID, []string{}, func(symbol result.Symbol) error {
		symbolInDBValue := types.SymbolToSymbolInDB(symbol)
		_, err := insertStatement.Exec(&symbolInDBValue)
		return err
	})
}

// updateSymbols adds/removes rows from the DB based on a `git diff` between the meta.revision within the
// DB and the given commitID.
func updateSymbols(ctx context.Context, gitserverClient GitserverClient, parser parser.Parser, dbFile string, repoName api.RepoName, commitID api.CommitID) (err error) {
	db, err := sqlx.Open("sqlite3_with_regexp", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	// Writing a bunch of rows into sqlite3 is much faster in a transaction.
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Read old commit
	row := tx.QueryRow(`SELECT revision FROM meta`)
	oldCommit := api.CommitID("")
	if err = row.Scan(&oldCommit); err != nil {
		return err
	}

	// Write new commit
	_, err = tx.Exec(`UPDATE meta SET revision = ?`, string(commitID))
	if err != nil {
		return err
	}

	// git diff
	changes, err := gitserverClient.GitDiff(ctx, repoName, oldCommit, commitID)
	if err != nil {
		return err
	}

	deleteStatement, err := tx.Prepare("DELETE FROM symbols WHERE path = ?")
	if err != nil {
		return err
	}

	insertStatement, err := tx.PrepareNamed(insertQuery)
	if err != nil {
		return err
	}

	for _, path := range append(changes.Deleted, changes.Modified...) {
		_, err := deleteStatement.Exec(path)
		return err
	}

	return parser.Parse(ctx, repoName, commitID, append(changes.Added, changes.Modified...), func(symbol result.Symbol) error {
		symbolInDBValue := types.SymbolToSymbolInDB(symbol)
		_, err := insertStatement.Exec(&symbolInDBValue)
		return err
	})
}

const insertQuery = `
	INSERT INTO symbols ( name,  namelowercase,  path,  pathlowercase,  line,  kind,  language,  parent,  parentkind,  signature,  pattern,  filelimited)
	VALUES              (:name, :namelowercase, :path, :pathlowercase, :line, :kind, :language, :parent, :parentkind, :signature, :pattern, :filelimited)`

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
