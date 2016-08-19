package traceutil

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"

	opentracing "github.com/opentracing/opentracing-go"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
)

// SQLExecutor records the timings of SQL queries and
// associates them with the span that originated them.
type SQLExecutor struct {
	gorp.SqlExecutor
	Context context.Context
}

func (x SQLExecutor) record(query string, args []interface{}) opentracing.Span {
	shortenedQuery := query
	if len(shortenedQuery) > 30 {
		shortenedQuery = query[:30] + "..."
	}
	_, file, line, _ := runtime.Caller(3)
	span, _ := opentracing.StartSpanFromContext(x.Context, fmt.Sprintf("SQL: %s", shortenedQuery))
	span.SetTag("Query", dbutil.UnbindQuery(query, args...))
	span.SetTag("File", fmt.Sprintf("%s:%d", file, line))
	return span
}

func (x SQLExecutor) Get(i interface{}, keys ...interface{}) (interface{}, error) {
	span := x.record(fmt.Sprintf("Get %T", i), keys)
	defer span.Finish()
	return x.SqlExecutor.Get(i, keys...)
}

func (x SQLExecutor) Insert(list ...interface{}) error {
	span := x.record("Insert", list)
	defer span.Finish()
	return x.SqlExecutor.Insert(list...)
}

func (x SQLExecutor) Update(list ...interface{}) (int64, error) {
	span := x.record("Update", list)
	defer span.Finish()
	return x.SqlExecutor.Update(list...)
}

func (x SQLExecutor) Delete(list ...interface{}) (int64, error) {
	span := x.record("Delete", list)
	defer span.Finish()
	return x.SqlExecutor.Delete(list...)
}

func (x SQLExecutor) Exec(query string, args ...interface{}) (sql.Result, error) {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.Exec(query, args...)
}

func (x SQLExecutor) Select(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.Select(i, query, args...)
}

func (x SQLExecutor) SelectInt(query string, args ...interface{}) (int64, error) {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.SelectInt(query, args...)
}

func (x SQLExecutor) SelectNullInt(query string, args ...interface{}) (sql.NullInt64, error) {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.SelectNullInt(query, args...)
}

func (x SQLExecutor) SelectFloat(query string, args ...interface{}) (float64, error) {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.SelectFloat(query, args...)
}

func (x SQLExecutor) SelectNullFloat(query string, args ...interface{}) (sql.NullFloat64, error) {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.SelectNullFloat(query, args...)
}

func (x SQLExecutor) SelectStr(query string, args ...interface{}) (string, error) {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.SelectStr(query, args...)
}

func (x SQLExecutor) SelectNullStr(query string, args ...interface{}) (sql.NullString, error) {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.SelectNullStr(query, args...)
}

func (x SQLExecutor) SelectOne(holder interface{}, query string, args ...interface{}) error {
	span := x.record(query, args)
	defer span.Finish()
	return x.SqlExecutor.SelectOne(holder, query, args...)
}

// UnderlyingSQLExecutor implements db.SQLExecutorWrapper.
func (x SQLExecutor) UnderlyingSQLExecutor() gorp.SqlExecutor { return x.SqlExecutor }
