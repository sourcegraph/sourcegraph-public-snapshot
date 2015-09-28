package dbutil

import (
	"database/sql"
	"log"

	"github.com/sqs/modl"
)

// A NoOpExecutor implements modl.SqlExecutor but does not issue SQL
// queries. Instead, it logs them to the console. It is useful when
// you are trying to eliminate SQL queries incrementally.
type NoOpExecutor struct {
	modl.SqlExecutor
}

func (t *NoOpExecutor) Get(i interface{}, keys ...interface{}) error {
	t.logOp("Get", keys)
	return nil
}

func (t *NoOpExecutor) Insert(list ...interface{}) error {
	t.logOp("Insert", list)
	return nil
}

func (t *NoOpExecutor) Update(list ...interface{}) (int64, error) {
	t.logOp("Update", list)
	return 0, nil
}

func (t *NoOpExecutor) Delete(list ...interface{}) (int64, error) {
	t.logOp("Delete", list)
	return 0, nil
}

func (t *NoOpExecutor) Exec(query string, args ...interface{}) (sql.Result, error) {
	t.logOp("Exec", query)
	return nil, nil
}

func (t *NoOpExecutor) Select(i interface{}, query string, args ...interface{}) error {
	t.logOp("Select", query)
	return nil
}

func (t *NoOpExecutor) SelectOne(i interface{}, query string, args ...interface{}) error {
	t.logOp("SelectOne", query)
	return nil
}

func (t *NoOpExecutor) Commit() error {
	t.logOp("Commit", nil)
	return nil
}

func (t *NoOpExecutor) Rollback() error {
	t.logOp("Rollback", nil)
	return nil
}

func (t *NoOpExecutor) Prepare(sql string) (*sql.Stmt, error) {
	t.logOp("Prepare", sql)
	return nil, nil
}

func (t *NoOpExecutor) logOp(op string, v interface{}) {
	yellow := func(s string) string {
		return "\x1b[33m" + s + "\x1b[0m"
	}
	log.Printf(yellow("sql[noop]: ")+"%s %v", op, v)
}

var _ modl.SqlExecutor = &NoOpExecutor{}
