package testdb

import (
	"github.com/sqs/modl"
	"src.sourcegraph.com/sourcegraph/util/dbutil2"
)

// NewHandle creates new test DB handles.
//
// NOTE: You must call done() when your test is finished, so that the
// DB can be reused. If the entire test process only calls
// NewHandle once, it's OK to not call done.
func NewHandle(schema *dbutil2.Schema) (main modl.SqlExecutor, done func()) {
	return pristineDBs(schema)
}
