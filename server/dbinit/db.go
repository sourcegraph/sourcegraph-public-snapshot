package dbinit

import (
	"src.sourcegraph.com/sourcegraph/server/internal/store/pgsql"
	sgxcli "src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	// Set the database schema to the global variable accessible from `serve_cmd.go`
	// for initiliazing the database at startup.
	sgxcli.DBSchema = pgsql.Schema
}
