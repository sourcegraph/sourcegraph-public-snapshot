package db

import "sourcegraph.com/pkg/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "enterprisedb"
}
