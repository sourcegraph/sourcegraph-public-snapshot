package globalstatedb

import "github.com/sourcegraph/sourcegraph/internal/database/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "globalstatedb"
}
