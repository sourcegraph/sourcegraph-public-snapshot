package db

import dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"

func init() {
	dbtesting.DBNameSuffix = "enterprisedb"
}
