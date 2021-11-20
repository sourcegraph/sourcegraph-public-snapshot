package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
)

type lifecycleHookConnectionResolver struct {
	store *store.Store
	opts  store.ListLifecycleHookOpts

	once  sync.Once
	hooks []graphqlbackend.BatchChangesLifecycleHookResolver
	next  int64
	err   error
}

var _ graphqlbackend.BatchChangesLifecycleHookConnectionResolver = &lifecycleHookConnectionResolver{}

func (r *lifecycleHookConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.BatchChangesLifecycleHookResolver, error) {
	hooks, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	return hooks, nil
}

func (r *lifecycleHookConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := store.CountLifecycleHooksOpts{
		IncludeExpired: r.opts.IncludeExpired,
	}

	count, err := r.store.CountLifecycleHooks(ctx, opts)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

func (r *lifecycleHookConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next == 0 {
		return graphqlutil.HasNextPage(false), nil
	}
	return graphqlutil.NextPageCursor(strconv.FormatInt(next, 10)), nil
}

func (r *lifecycleHookConnectionResolver) compute(ctx context.Context) ([]graphqlbackend.BatchChangesLifecycleHookResolver, int64, error) {
	r.once.Do(func() {
		hooks, next, err := r.store.ListLifecycleHooks(ctx, r.opts)
		if err != nil {
			r.err = err
			return
		}

		r.next = next
		r.hooks = make([]graphqlbackend.BatchChangesLifecycleHookResolver, len(hooks))
		for i := range hooks {
			r.hooks[i] = &lifecycleHookResolver{
				store: r.store,
				hook:  hooks[i],
			}
		}
	})

	return r.hooks, r.next, r.err
}
