package executors

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/services/executors/store"
	postgres "github.com/sourcegraph/sourcegraph/internal/services/executors/store/db"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Executor interface {
	List(ctx context.Context, query string, active bool, offset int, limit int) ([]types.Executor, int, error)
	GetByID(ctx context.Context, id int) (types.Executor, bool, error)
	UpsertHeartbeat(ctx context.Context, executor types.Executor) error
	DeleteInactiveHeartbeats(ctx context.Context, minAge time.Duration) error
	GetByHostname(ctx context.Context, hostname string) (types.Executor, bool, error)
}

func New(db dbutil.DB) Executor {
	return &executorService{store: postgres.New(db)}
}

type executorService struct {
	store store.Store
}

func (s *executorService) List(ctx context.Context, query string, active bool, offset int, limit int) ([]types.Executor, int, error) {
	args := store.ExecutorStoreListOptions{
		Query:  query,
		Active: active,
		Offset: offset,
		Limit:  limit,
	}

	return s.store.List(ctx, args)
}

func (s *executorService) GetByID(ctx context.Context, id int) (types.Executor, bool, error) {
	return s.store.GetByID(ctx, id)
}

func (s *executorService) UpsertHeartbeat(ctx context.Context, executor types.Executor) error {
	return s.store.UpsertHeartbeat(ctx, executor)
}

func (s *executorService) DeleteInactiveHeartbeats(ctx context.Context, minAge time.Duration) error {
	return s.store.DeleteInactiveHeartbeats(ctx, minAge)
}

func (s *executorService) GetByHostname(ctx context.Context, hostname string) (types.Executor, bool, error) {
	return s.store.GetByHostname(ctx, hostname)
}
