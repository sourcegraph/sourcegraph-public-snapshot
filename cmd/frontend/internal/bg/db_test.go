package bg

import "github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "bgdb"
}
