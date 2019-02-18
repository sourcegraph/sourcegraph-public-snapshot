package symbols

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/keegancsmith/sqlf"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"golang.org/x/net/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// maxFileSize is the limit on file size in bytes. Only files smaller than this are processed.
const maxFileSize = 1 << 19 // 512KB

func (s *Service) handleSearch(w http.ResponseWriter, r *http.Request) {
	var args protocol.SearchArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.search(r.Context(), args)
	if err != nil {
		if err == context.Canceled && r.Context().Err() == context.Canceled {
			return // client went away
		}
		log15.Error("Symbol search failed", "args", args, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) search(ctx context.Context, args protocol.SearchArgs) (result *protocol.SearchResult, err error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	log15.Debug("Symbol search", "repo", args.Repo, "query", args.Query)

	span, ctx := opentracing.StartSpanFromContext(ctx, "search")
	span.SetTag("repo", args.Repo)
	span.SetTag("commitID", args.CommitID)
	span.SetTag("query", args.Query)
	span.SetTag("first", args.First)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	tr := trace.New("symbols.search", fmt.Sprintf("args:%+v", args))
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	db, err := s.getDB(ctx, args)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	result = &protocol.SearchResult{}
	res, err := filterSymbols(ctx, db, args)
	if err != nil {
		return nil, err
	}
	result.Symbols = res
	return result, nil
}

func (s *Service) getDB(ctx context.Context, args protocol.SearchArgs) (*sqlx.DB, error) {
	var err error

	repoAtCommit := repoAtCommit{
		RepoName: args.Repo,
		CommitID: args.CommitID,
	}

	err = os.MkdirAll(filepath.Dir(s.dbFilename(repoAtCommit)), os.ModePerm)
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Open("sqlite3_with_pcre", s.dbFilename(repoAtCommit))
	if err != nil {
		return nil, err
	}

	inFlightMutex.Lock()
	if _, ok := inFlight[repoAtCommit]; !ok {
		inFlight[repoAtCommit] = &indexingState{
			Mutex: &sync.Mutex{},
			Error: nil,
		}
	}
	inFlight[repoAtCommit].Mutex.Lock()
	inFlightMutex.Unlock()

	if _, err := db.Exec(`SELECT 1 FROM symbols LIMIT 1;`); err != nil {
		bgCtx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		symbols, err := s.parseUncached(bgCtx, args.Repo, args.CommitID)
		if err != nil {
			cancel()
			return nil, err
		}
		inFlight[repoAtCommit].Error = s.writeSymbols(db, symbols)
		cancel()
	}
	inFlight[repoAtCommit].Mutex.Unlock()

	if inFlight[repoAtCommit].Error != nil {
		return nil, inFlight[repoAtCommit].Error
	}

	return db, nil
}

func filterSymbols(ctx context.Context, db *sqlx.DB, args protocol.SearchArgs) (res []protocol.Symbol, err error) {
	start := time.Now()
	defer func() {
		fmt.Printf("filterSymbolsSqlite %.3f %s\n", time.Now().Sub(start).Seconds(), args.Query)
	}()
	span, _ := opentracing.StartSpanFromContext(ctx, "filterSymbols")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	const maxFirst = 500
	if args.First < 0 || args.First > maxFirst {
		args.First = maxFirst
	}

	mkConds := func(column string, regex string) []*sqlf.Query {
		conditions := []*sqlf.Query{}

		if regex == "" {
			return conditions
		}

		if strings.HasPrefix(regex, "^") && strings.HasSuffix(regex, "$") {
			// It looks like the user is asking for exact matches, so use `=`
			// for speed. Checking for `^...$` isn't 100% accurate, but it
			// covers 99.9% of cases.
			rhs := strings.TrimSuffix(strings.TrimPrefix(regex, "^"), "$")
			if !args.IsCaseSensitive {
				rhs = strings.ToLower(rhs)
				column = column + "lowercase"
			}
			conditions = append(conditions, sqlf.Sprintf(column+" = %s", rhs))
		} else {
			if !args.IsCaseSensitive {
				regex = "(?i:" + regex + ")"
			}
			conditions = append(conditions, sqlf.Sprintf(column+" REGEXP %s", regex))
		}

		return conditions
	}

	negateAll := func(oldConditions []*sqlf.Query) []*sqlf.Query {
		newConditions := []*sqlf.Query{}

		for _, oldCondition := range oldConditions {
			newConditions = append(newConditions, sqlf.Sprintf("NOT %s", oldCondition))
		}

		return newConditions
	}

	var conds []*sqlf.Query
	conds = append(conds, mkConds("name", args.Query)...)
	for _, i := range args.IncludePatterns {
		conds = append(conds, mkConds("path", i)...)
	}
	conds = append(conds, negateAll(mkConds("path", args.ExcludePattern))...)

	var sqlQuery *sqlf.Query
	if len(conds) == 0 {
		sqlQuery = sqlf.Sprintf("SELECT * FROM symbols LIMIT %s", args.First)
	} else {
		sqlQuery = sqlf.Sprintf("SELECT * FROM symbols WHERE %s LIMIT %s", sqlf.Join(conds, "AND"), args.First)
	}

	var symbolsInDB []symbolInDB
	err = db.Select(&symbolsInDB, sqlQuery.Query(sqlf.PostgresBindVar), sqlQuery.Args()...)
	if err != nil {
		return nil, err
	}

	for _, symbolInDB := range symbolsInDB {
		res = append(res, symbolInDBToSymbol(symbolInDB))
	}

	span.SetTag("hits", len(res))
	return res, nil
}

// The version of the schema symbols databases. This is included in the database
// filenames to prevent a newer version of the symbols service attempting to
// read from a database created by an older (and likely incompatible) symbols
// service. Increment this when you change the database schema.
const symbolsDbVersion = 1

// Returns the filename of the database corresponding to the given repo@commit.
// There is a 1-1 correspondence between database files and repo@commits.
func (s *Service) dbFilename(repoAtCommit repoAtCommit) string {
	return path.Join(s.Path, fmt.Sprintf("v%d-%s@%s.sqlite", symbolsDbVersion, strings.Replace(string(repoAtCommit.RepoName), "/", "-", -1), repoAtCommit.CommitID))
}

// A repo@commit.
type repoAtCommit struct {
	RepoName api.RepoName
	CommitID api.CommitID
}

// The state of indexing a particular repo@commit into sqlite.
type indexingState struct {
	Mutex *sync.Mutex
	Error error
}

var inFlightMutex = &sync.Mutex{}
var inFlight = map[repoAtCommit]*indexingState{}

// symbolInDB is a code symbol as represented in the sqlite database. It's the
// same as `Symbol`, but with namelowercase and pathlowercase, which enable
// indexed case insensitive queries.
type symbolInDB struct {
	Name          string
	NameLowercase string
	Path          string
	PathLowercase string
	Line          int
	Kind          string
	Language      string
	Parent        string
	ParentKind    string
	Signature     string
	Pattern       string

	FileLimited bool
}

func symbolToSymbolInDB(symbol protocol.Symbol) symbolInDB {
	return symbolInDB{
		Name:          symbol.Name,
		NameLowercase: strings.ToLower(symbol.Name),
		Path:          symbol.Path,
		PathLowercase: strings.ToLower(symbol.Path),
		Line:          symbol.Line,
		Kind:          symbol.Kind,
		Language:      symbol.Language,
		Parent:        symbol.Parent,
		ParentKind:    symbol.ParentKind,
		Signature:     symbol.Signature,
		Pattern:       symbol.Pattern,

		FileLimited: symbol.FileLimited,
	}
}

func symbolInDBToSymbol(symbolInDB symbolInDB) protocol.Symbol {
	return protocol.Symbol{
		Name:       symbolInDB.Name,
		Path:       symbolInDB.Path,
		Line:       symbolInDB.Line,
		Kind:       symbolInDB.Kind,
		Language:   symbolInDB.Language,
		Parent:     symbolInDB.Parent,
		ParentKind: symbolInDB.ParentKind,
		Signature:  symbolInDB.Signature,
		Pattern:    symbolInDB.Pattern,

		FileLimited: symbolInDB.FileLimited,
	}
}

func (s *Service) writeSymbols(db *sqlx.DB, symbols []protocol.Symbol) error {
	// Wrapping all database modifications in a transaction ensures that if the
	// `symbols` table exists, then it contains all symbols for the repo@commit.
	// This allows a goroutine servicing a query to determine whether or not a
	// repo@commit has already been indexed by checking for the existence of the
	// `symbols` table.
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// sqlx lowercases struct fields by default.
	// http://jmoiron.github.io/sqlx/#query
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS symbols (
			name VARCHAR(256) NOT NULL,
			namelowercase VARCHAR(256) NOT NULL,
			path VARCHAR(4096) NOT NULL,
			pathlowercase VARCHAR(256) NOT NULL,
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

	for _, symbol := range symbols {
		symbolInDBValue := symbolToSymbolInDB(symbol)
		_, err := tx.NamedExec(
			fmt.Sprintf(
				"INSERT INTO symbols %s VALUES %s",
				"( name,  namelowercase,  path,  pathlowercase,  line,  kind,  language,  parent,  parentkind,  signature,  pattern,  filelimited)",
				"(:name, :namelowercase, :path, :pathlowercase, :line, :kind, :language, :parent, :parentkind, :signature, :pattern, :filelimited)"),
			&symbolInDBValue)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
