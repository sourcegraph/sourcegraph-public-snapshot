package confdb

import "sourcegraph.com/pkg/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "confdb"
}
