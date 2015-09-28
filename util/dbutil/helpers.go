package dbutil

import (
	"database/sql"
	"reflect"

	"github.com/sqs/modl"
)

// SelectBool executes the given query, which should be a SELECT statement for a single
// int column, and returns the value of the first row returned.  If no rows are
// found, zero is returned.
func SelectBool(e modl.SqlExecutor, query string, args ...interface{}) (bool, error) {
	var h bool
	err := selectVal(e, &h, query, args...)
	if err != nil {
		return false, err
	}
	return h, nil
}

// SelectInt executes the given query, which should be a SELECT statement for a single
// integer column, and returns the value of the first row returned.  If no rows are
// found, sql.ErrNoRows is returned.
func SelectInt(e modl.SqlExecutor, query string, args ...interface{}) (int64, error) {
	var h int64
	err := selectVal(e, &h, query, args...)
	if err != nil {
		return 0, err
	}
	return h, nil
}

// SelectString executes the given query, which should be a SELECT statement for a single
// char/varchar column, and returns the value of the first row returned.  If no rows are
// found, an empty string is returned.
func SelectString(e modl.SqlExecutor, query string, args ...interface{}) (string, error) {
	var h string
	err := selectVal(e, &h, query, args...)
	if err != nil {
		return "", err
	}
	return h, nil
}

// SelectNullStr executes the given query, which should be a SELECT statement for a single
// char/varchar column, and returns the value of the first row returned.  If no rows are
// found, the empty sql.NullString is returned.
func SelectNullStr(e modl.SqlExecutor, query string, args ...interface{}) (sql.NullString, error) {
	var h sql.NullString
	err := selectVal(e, &h, query, args...)
	if err != nil {
		return h, err
	}
	return h, nil
}

func selectVal(e modl.SqlExecutor, v interface{}, query string, args ...interface{}) error {
	e = GetUnderlyingSQLExecutor(e)

	var row *sql.Row
	switch e := e.(type) {
	case *modl.DbMap:
		row = e.Dbx.QueryRow(query, args...)
	case *modl.Transaction:
		row = e.Tx.QueryRow(query, args...)
	default:
		panic("selectVal: unknown type: " + reflect.TypeOf(e).String())
	}

	if row != nil {
		err := row.Scan(v)
		if err != nil {
			return err
		}
	}

	return nil
}
