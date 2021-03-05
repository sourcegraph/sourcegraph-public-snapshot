package httpapi

import "github.com/sourcegraph/sourcegraph/internal/database/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "httpapidb"
}
