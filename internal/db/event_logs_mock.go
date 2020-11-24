package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockEventLogs struct {
	LatestPing func(ctx context.Context) (*types.Event, error)
}
