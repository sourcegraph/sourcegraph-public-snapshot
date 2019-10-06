package rules

import "github.com/sourcegraph/sourcegraph/internal/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "rules"
}
