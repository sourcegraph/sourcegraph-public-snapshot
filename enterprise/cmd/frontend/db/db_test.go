package db

import "github.com/sourcegraph/sourcegraph/cmd/internal/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "enterprisedb"
}
