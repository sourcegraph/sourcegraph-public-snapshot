package db

import (
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func init() {
	dbtesting.BeforeTest = append(dbtesting.BeforeTest, func() { Mocks = MockStores{} })
}
