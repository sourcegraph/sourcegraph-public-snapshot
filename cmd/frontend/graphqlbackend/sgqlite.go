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
	registerOnce.Do(func() {
		sql.Register("sqlite3_with_extensions", &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.CreateModule("search", &searchModule{db: r.db})
			},
		})
	})

	dblite, err := sql.Open("sqlite3_with_extensions", ":memory:")
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

	var graphqlRows [][]graphqlValue
	for rows.Next() {
		graphqlVals := make([]graphqlValue, len(cols))
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

type graphqlValue struct {
	value interface{}
}

func (graphqlValue) ImplementsGraphQLType(name string) bool {
	return name == "JSONValue"
}

func (v graphqlValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *graphqlValue) UnmarshalGraphQL(input interface{}) error {
	*v = graphqlValue{value: input}
	return nil
}

func (v *graphqlValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.value)
}

// Scan satisfies SQL scanner interface
func (j *graphqlValue) Scan(src interface{}) error {
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
	case nil:
	default:
		return errors.Errorf("invalid type %T", src)
	}
	return nil
}

type ExecutionResult struct {
	columnNames []string
	rows        [][]graphqlValue
}

func (e *ExecutionResult) ColumnNames() []string {
	return e.columnNames
}

func (e *ExecutionResult) Rows() [][]graphqlValue {
	return e.rows
}

type searchModule struct {
	db database.DB
}

func (m *searchModule) EponymousOnlyModule() {}

func (m *searchModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			result_type TEXT,
			repo_id INT,
			repo_name TEXT,
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
	db      database.DB
	index   int
	query   string
	results result.Matches
}

func (vc *searchResultCursor) Column(c *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		switch vc.results[vc.index].(type) {
		case *result.RepoMatch:
			c.ResultText("repo")
		case *result.FileMatch:
			c.ResultText("file")
		case *result.CommitMatch:
			c.ResultText("commit")
		default:
			return errors.New("unknown type")
		}
	case 1:
		c.ResultInt(int(vc.results[vc.index].RepoName().ID))
	case 2:
		c.ResultText(string(vc.results[vc.index].RepoName().Name))
	case 3:
		c.ResultText(vc.query)
	}
	return nil
}

func (vc *searchResultCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	if len(vals) == 0 {
		return nil
	}
	vc.query = vals[0].(string)
	// TODO how to capture context, db, and args from request
	agg := streaming.NewAggregatingStream()
	imp, err := NewSearchImplementer(context.Background(), vc.db, &SearchArgs{
		Query:   vc.query,
		Stream:  agg,
		Version: "V2",
	})
	if err != nil {
		fmt.Printf("%s\n", err)
		return errors.Wrap(err, "newSearchImplementor")
	}
	_, err = imp.Results(context.Background())
	if err != nil {
		fmt.Printf("%s\n", err)
		return errors.Wrap(err, "searchImplementor.Results")
	}
	vc.index = 0
	vc.results = agg.Results
	return nil
}

func (vc *searchResultCursor) Next() error {
	vc.index++
	return nil
}

func (vc *searchResultCursor) EOF() bool {
	return vc.index >= len(vc.results)
}

func (vc *searchResultCursor) Rowid() (int64, error) {
	return int64(vc.index), nil
}

func (vc *searchResultCursor) Close() error {
	return nil
}
