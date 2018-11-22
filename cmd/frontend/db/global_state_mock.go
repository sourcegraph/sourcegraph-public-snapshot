package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type MockGlobalState struct {
	Get func(ctx context.Context) (*types.GlobalState, error)
}
