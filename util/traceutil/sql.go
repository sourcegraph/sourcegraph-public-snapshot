package traceutil

import (
	"database/sql"
	"time"

	"github.com/sqs/modl"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/sqltrace"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil"
)

// SQLExecutor records the timings of SQL queries in appdash and
// associates them with the span that originated them.
type SQLExecutor struct {
	modl.SqlExecutor
	*appdash.Recorder
}

func (x SQLExecutor) record(start time.Time, query string, args []interface{}) {
	rec := x.Recorder.Child()
	rec.Name("SQL")
	rec.Event(sqltrace.SQLEvent{
		SQL:        dbutil.UnbindQuery(query, args...),
		ClientSend: start,
		ClientRecv: time.Now(),
	})
}

func (x SQLExecutor) Exec(query string, args ...interface{}) (sql.Result, error) {
	defer x.record(time.Now(), query, args)
	return x.SqlExecutor.Exec(query, args...)
}

func (x SQLExecutor) Select(dest interface{}, query string, args ...interface{}) error {
	defer x.record(time.Now(), query, args)
	return x.SqlExecutor.Select(dest, query, args...)
}

// UnderlyingSQLExecutor implements db.SQLExecutorWrapper.
func (x SQLExecutor) UnderlyingSQLExecutor() modl.SqlExecutor { return x.SqlExecutor }
