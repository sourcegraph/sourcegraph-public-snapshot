package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type Mocks struct {
	Get func(ctx context.Context) (*globalstatedb.State, error)
}

var Mock = Mocks{}
