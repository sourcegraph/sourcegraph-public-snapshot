package db

import "sourcegraph.com/pkg/db/dbtesting"

func init() {
	dbtesting.BeforeTest = append(dbtesting.BeforeTest, func() { Mocks = MockStores{} })
}
