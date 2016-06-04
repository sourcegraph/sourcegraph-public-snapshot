package testdb

import (
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
)

// NewHandle creates new test DB handles.
//
// NOTE: You must call done() when your test is finished, so that the
// DB can be reused. If the entire test process only calls
// NewHandle once, it's OK to not call done.
func NewHandle(dbName string, schema *dbutil2.Schema) (main gorp.SqlExecutor, done func()) {
	return pristineDBs(dbName, schema)
}
