package bg

import "sourcegraph.com/pkg/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "bgdb"
}
