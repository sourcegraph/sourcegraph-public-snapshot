package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// discussionThreadTargetConnectionResolver resolves a list of discussion thread targets.
//
// 🚨 SECURITY: When instantiating an discussionThreadTargetConnectionResolver value, the caller
// MUST check permissions.
type discussionThreadTargetConnectionResolver struct {
	threadID int64
	args     *graphqlutil.ConnectionArgs

	// cache results because they are used by multiple fields
	once    sync.Once
	targets []*types.DiscussionThreadTargetRepo
	err     error
}

func (r *discussionThreadTargetConnectionResolver) compute(ctx context.Context) ([]*types.DiscussionThreadTargetRepo, error) {
	r.once.Do(func() {
		r.targets, r.err = db.DiscussionThreads.ListTargets(ctx, r.threadID)
	})
	return r.targets, r.err
}

func (r *discussionThreadTargetConnectionResolver) Nodes(ctx context.Context) ([]*discussionThreadTargetResolver, error) {
	targets, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.args.First != nil {
		first := int(*r.args.First)
		if first < 0 {
			first = 0
		}
		if first > len(targets) {
			first = len(targets)
		}
		targets = targets[:first]
	}

	var l []*discussionThreadTargetResolver
	for _, t := range targets {
		l = append(l, &discussionThreadTargetResolver{t: t})
	}
	return l, nil
}

func (r *discussionThreadTargetConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	targets, err := r.compute(ctx)
	return int32(len(targets)), err
}

func (r *discussionThreadTargetConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	targets, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	return graphqlutil.HasNextPage(r.args.First != nil && len(targets) > int(*r.args.First)), nil
}
