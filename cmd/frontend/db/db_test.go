package db

import "github.com/sourcegraph/sourcegraph/cmd/internal/db/dbtesting"

func init() {
	dbtesting.BeforeTest = append(dbtesting.BeforeTest, func() { Mocks = MockStores{} })
}
