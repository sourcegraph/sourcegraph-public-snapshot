package modl

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// a cursor is either an sqlx.Db or an sqlx.Tx
type handle interface {
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	/*
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	*/
}

// an implmentation of handle which traces using dbmap
type tracingHandle struct {
	d *DbMap
	h handle
}

func (t *tracingHandle) Select(dest interface{}, query string, args ...interface{}) error {
	t.d.trace(query, args...)
	return t.h.Select(dest, query, args...)
}

func (t *tracingHandle) Get(dest interface{}, query string, args ...interface{}) error {
	t.d.trace(query, args...)
	return t.h.Get(dest, query, args...)
}

func (t *tracingHandle) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	t.d.trace(query, args...)
	return t.h.Queryx(query, args...)
}

func (t *tracingHandle) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	t.d.trace(query, args...)
	return t.h.QueryRowx(query, args...)
}

func (t *tracingHandle) Exec(query string, args ...interface{}) (sql.Result, error) {
	t.d.trace(query, args...)
	return t.h.Exec(query, args...)
}
