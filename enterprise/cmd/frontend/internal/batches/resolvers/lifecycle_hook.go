package resolvers

import (
	"context"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

const lifecycleHookIDKind = "BatchChangesLifecycleHook"

func marshalLifecycleHookID(id int64) graphql.ID {
	return relay.MarshalID(lifecycleHookIDKind, id)
}

func unmarshalLifecycleHookID(id graphql.ID) (hid int64, err error) {
	err = relay.UnmarshalSpec(id, &hid)
	return
}

func (r *Resolver) CreateBatchChangesLifecycleHook(
	ctx context.Context,
	args *graphqlbackend.CreateBatchChangesLifecycleHookArgs,
) (graphqlbackend.BatchChangesLifecycleHookResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	var expiresAt time.Time
	if args.ExpiresAt != nil {
		expiresAt = args.ExpiresAt.Time
	}

	hook := btypes.LifecycleHook{
		ExpiresAt: expiresAt,
		URL:       args.URL,
		Secret:    args.Secret,
	}
	if err := r.store.CreateLifecycleHook(ctx, &hook); err != nil {
		return nil, err
	}

	return &lifecycleHookResolver{
		store: r.store,
		hook:  &hook,
	}, nil
}

func (r *Resolver) DeleteBatchChangesLifecycleHook(
	ctx context.Context,
	args *graphqlbackend.DeleteBatchChangesLifecycleHookArgs,
) (*graphqlbackend.EmptyResponse, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	id, err := unmarshalLifecycleHookID(args.ID)
	if err != nil {
		return nil, err
	}

	if err := r.store.DeleteLifecycleHook(ctx, id); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) BatchChangesLifecycleHooks(
	ctx context.Context,
	args *graphqlbackend.ListBatchChangesLifecycleHooksArgs,
) (graphqlbackend.BatchChangesLifecycleHookConnectionResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		return nil, err
	}

	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}

	opts := store.ListLifecycleHookOpts{
		LimitOpts:      store.LimitOpts{Limit: int(args.First)},
		IncludeExpired: false,
	}

	if args.After != nil {
		cursor, err := strconv.ParseInt(*args.After, 10, 64)
		if err != nil {
			return nil, err
		}
		opts.Cursor = cursor
	}

	if args.IncludeExpired != nil && *args.IncludeExpired {
		opts.IncludeExpired = true
	}

	return &lifecycleHookConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

type lifecycleHookResolver struct {
	store *store.Store
	hook  *btypes.LifecycleHook
}

var _ graphqlbackend.Node = &lifecycleHookResolver{}
var _ graphqlbackend.BatchChangesLifecycleHookResolver = &lifecycleHookResolver{}

func (r *lifecycleHookResolver) ID() graphql.ID {
	return marshalLifecycleHookID(r.hook.ID)
}

func (r *lifecycleHookResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.hook.CreatedAt}
}

func (r *lifecycleHookResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.hook.UpdatedAt}
}

func (r *lifecycleHookResolver) ExpiresAt() *graphqlbackend.DateTime {
	if r.hook.ExpiresAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.hook.ExpiresAt}
}

func (r *lifecycleHookResolver) URL() string {
	return r.hook.URL
}

func (r *lifecycleHookResolver) Secret() string {
	return r.hook.Secret
}
