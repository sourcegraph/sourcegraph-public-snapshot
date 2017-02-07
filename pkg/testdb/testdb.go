package testdb

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"

// NewHandle creates new test DB handles.
func NewHandle(dbName string, schema *dbutil2.Schema) (main *dbutil2.Handle) {
	return pristineDBs(dbName, schema)
}
