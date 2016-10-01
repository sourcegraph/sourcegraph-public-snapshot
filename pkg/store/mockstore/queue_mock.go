package mockstore

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

type Queue struct {
	Enqueue func(ctx context.Context, j *store.Job) error
	LockJob func(ctx context.Context) (*store.LockedJob, error)
	Stats   func(ctx context.Context) (map[string]store.QueueStats, error)
}
