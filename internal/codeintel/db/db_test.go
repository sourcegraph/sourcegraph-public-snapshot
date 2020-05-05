package db

import "github.com/sourcegraph/sourcegraph/internal/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "precise-code-intel-api-server"
}
