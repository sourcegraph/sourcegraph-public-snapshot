package db

import (
	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
)

func init() {
	dbtesting.BeforeTest = append(dbtesting.BeforeTest, func() { Mocks = MockStores{} })
}
