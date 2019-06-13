package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type DiscussionThreadConnection interface {
	Nodes(context.Context) ([]*discussionThreadResolver, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type discussionThreadConnectionWithListOptions struct {
	options db.DiscussionThreadsListOptions

	once    sync.Once
	threads []*types.DiscussionThread
	err     error
}

func NewDiscussionThreadConnectionWithListOptions(options db.DiscussionThreadsListOptions) DiscussionThreadConnection {
	return &discussionThreadConnectionWithListOptions{options: options}
}

func (r *discussionThreadConnectionWithListOptions) compute(ctx context.Context) ([]*types.DiscussionThread, error) {
	r.once.Do(func() {
		r.threads, r.err = db.DiscussionThreads.List(ctx, &r.options)
	})
	return r.threads, r.err
}

func (r *discussionThreadConnectionWithListOptions) Nodes(ctx context.Context) ([]*discussionThreadResolver, error) {
	threads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	var l []*discussionThreadResolver
	for _, thread := range threads {
		l = append(l, &discussionThreadResolver{t: thread})
	}
	return l, nil
}

func (r *discussionThreadConnectionWithListOptions) TotalCount(ctx context.Context) (int32, error) {
	withoutLimit := r.options
	withoutLimit.LimitOffset = nil
	count, err := db.DiscussionThreads.Count(ctx, &withoutLimit)
	return int32(count), err
}

func (r *discussionThreadConnectionWithListOptions) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	threads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.options.LimitOffset != nil && len(threads) > r.options.Limit), nil
}
