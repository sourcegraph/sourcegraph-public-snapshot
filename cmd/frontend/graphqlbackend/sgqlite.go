package graphqlbackend

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mattn/go-sqlite3"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var registerOnce sync.Once

func (r *schemaResolver) Execute(ctx context.Context, args struct{ Query string }) (*ExecutionResult, error) {
	// TODO this is a bit hacky because it captures the database handle from the first request.
	// Not a deal-breaker, but also not great. I haven't yet figured out a way to pass request-scoped
	// through SQLite.
	registerOnce.Do(func() {
		sql.Register("sqlite3_with_extensions", &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.CreateModule("search", &searchModule{db: r.db})
			},
		})
	})

	// Using memory right now because it's easy, but we could also create a temp file which should
	// help ease any issues with holding the whole result set in memory for sorting/filtering/etc.
	dblite, err := sql.Open("sqlite3_with_extensions", "file:/tmp/test.db?_sync=OFF&_journal=OFF&_query_only=TRUE")
	if err != nil {
		return nil, err
	}
	defer dblite.Close()

	rows, err := dblite.QueryContext(ctx, args.Query)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var graphqlRows [][]jsonVal
	for rows.Next() {
		// make one scan target per column, then wrap them in interfaces for Scan
		graphqlVals := make([]jsonVal, len(cols))
		ifaces := make([]interface{}, len(graphqlVals))
		for i := range graphqlVals {
			ifaces[i] = &graphqlVals[i]
		}
		if err := rows.Scan(ifaces...); err != nil {
			return nil, err
		}
		graphqlRows = append(graphqlRows, graphqlVals)
	}

	return &ExecutionResult{columnNames: cols, rows: graphqlRows}, nil
}

// A copy of JSONValue that implements the SQL scanner interface
type jsonVal struct {
	value interface{}
}

func (jsonVal) ImplementsGraphQLType(name string) bool {
	return name == "JSONValue"
}

func (v jsonVal) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *jsonVal) UnmarshalGraphQL(input interface{}) error {
	*v = jsonVal{value: input}
	return nil
}

func (v *jsonVal) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.value)
}

// Scan satisfies SQL scanner interface
func (j *jsonVal) Scan(src interface{}) error {
	// These are all the types the SQLite driver is documented to return
	switch v := src.(type) {
	case int:
		j.value = v
	case int64:
		i := int(v)
		j.value = i
	case float64:
		j.value = v
	case bool:
		j.value = v
	case []byte:
		s := string(v)
		j.value = s
	case string:
		j.value = v
	case time.Time:
		j.value = v
	default:
		j.value = nil
	}
	return nil
}

type ExecutionResult struct {
	columnNames []string
	rows        [][]jsonVal
}

func (e *ExecutionResult) ColumnNames() []string {
	return e.columnNames
}

func (e *ExecutionResult) Rows() [][]jsonVal {
	return e.rows
}

type searchModule struct {
	db database.DB
}

// EponymousOnlyModule is a maker that lets us treat the table like a table-valued function
func (m *searchModule) EponymousOnlyModule() {}

func (m *searchModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	// The hidden `query` column is populated by the "function" call.
	// `from search('TODO')` is equivalent to `from search where query = 'TODO'`
	// because hidden the args to a function call are connected in order of the
	// hidden columns on the table.
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			result_type TEXT,
			repo_id INT,
			repo_name TEXT,
			file_name TEXT,
			query HIDDEN TEXT
		)`, args[0]))
	if err != nil {
		return nil, errors.Wrap(err, "create table")
	}
	return &searchTable{db: m.db}, nil
}

func (m *searchModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *searchModule) DestroyModule() {}

type searchTable struct {
	db database.DB
}

func (t *searchTable) Open() (sqlite3.VTabCursor, error) {
	return &searchResultCursor{db: t.db}, nil
}

func (v *searchTable) BestIndex(csts []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	// This marks the 'query' column as used so sqlite knows that
	// it doesn't need to do the filtering itself.
	// TODO don't mark other WHERE x = y clauses as used
	used := make([]bool, len(csts))
	for c, cst := range csts {
		if cst.Usable && cst.Op == sqlite3.OpEQ {
			used[c] = true
		}
	}

	return &sqlite3.IndexResult{
		IdxNum: 0,
		IdxStr: "default",
		Used:   used,
	}, nil
}

func (v *searchTable) Disconnect() error { return nil }
func (v *searchTable) Destroy() error    { return nil }

type searchResultCursor struct {
	db    database.DB
	query string

	wg     sync.WaitGroup
	cancel context.CancelFunc

	resultChan chan streaming.SearchEvent
	done       bool

	batch    result.Matches
	batchIdx int
	curRowID int64
}

func (vc *searchResultCursor) Column(c *sqlite3.SQLiteContext, col int) error {
	// This is where we convert the current result into columns.
	// Column number are in the order the table was defined
	// TODO add additional columns here
	switch col {
	case 0: // result_type
		switch vc.batch[vc.batchIdx].(type) {
		case *result.RepoMatch:
			c.ResultText("repo")
		case *result.FileMatch:
			c.ResultText("file")
		case *result.CommitMatch:
			c.ResultText("commit")
		default:
			return errors.New("unknown type")
		}
	case 1: // repo_id
		c.ResultInt(int(vc.batch[vc.batchIdx].RepoName().ID))
	case 2: // repo_name
		c.ResultText(string(vc.batch[vc.batchIdx].RepoName().Name))
	case 3: // file_name
		if fileMatch, ok := vc.batch[vc.batchIdx].(*result.FileMatch); ok {
			c.ResultText(fileMatch.Path)
		} else {
			// null if not a result.FileMatch
			c.ResultNull()
		}
	case 4: // query (hidden)
		c.ResultText(vc.query)
	}
	return nil
}

func (vc *searchResultCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	vc.cancel = cancel

	// Filter is is what tells the ResultCursor about the WHERE clauses (basically).
	// Its meaning is coupled with BestIndex.
	// TODO figure out exactly how BestIndex and Filter need to communicate.
	if len(vals) == 0 {
		return nil
	}
	// We expect the only argument right now to be the query passed to `search()`
	// TODO we should also figure out a way to support other WHERE clauses
	vc.query = vals[0].(string)
	// TODO how to capture context, db, and args from request
	resultChan := make(chan streaming.SearchEvent, 32)
	agg := streaming.StreamFunc(func(e streaming.SearchEvent) {
		resultChan <- e
	})
	imp, err := NewSearchImplementer(ctx, vc.db, &SearchArgs{
		Query:   vc.query,
		Stream:  agg,
		Version: "V2",
	})
	if err != nil {
		return errors.Wrap(err, "newSearchImplementor")
	}

	vc.wg = sync.WaitGroup{}
	vc.wg.Add(1)
	go func() {
		defer vc.wg.Done()
		defer close(resultChan)

		_, err = imp.Results(ctx)
		if err != nil {
			// TODO collect this error
			panic(err)
		}
	}()

	vc.resultChan = resultChan
	vc.batch = nil
	vc.batchIdx, vc.curRowID = -1, -1
	return vc.Next()
}

func (vc *searchResultCursor) Next() error {
	// Increment counters
	vc.batchIdx++
	vc.curRowID++

	// Read events from the channel until we get results
	for vc.batchIdx >= len(vc.batch) {
		event, ok := <-vc.resultChan
		if !ok {
			vc.done = true
			return nil
		}
		vc.batch = event.Results
		vc.batchIdx = 0
	}
	return nil
}

func (vc *searchResultCursor) EOF() bool {
	return vc.done
}

func (vc *searchResultCursor) Rowid() (int64, error) {
	return vc.curRowID, nil
}

func (vc *searchResultCursor) Close() error {
	vc.cancel()
	vc.wg.Wait()
	return nil
}
