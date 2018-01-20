package graphqlbackend

import (
	"context"
	"sync"
	"time"

	graphql "github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func (r *siteResolver) Repositories(args *struct {
	connectionArgs
	Query    *string
	Enabled  bool
	Disabled bool
}) (*repositoryConnectionResolver, error) {
	opt := db.ReposListOptions{
		Enabled:  args.Enabled,
		Disabled: args.Disabled,
	}
	if args.Query != nil {
		opt.Query = *args.Query
	}
	args.connectionArgs.set(&opt.LimitOffset)
	return &repositoryConnectionResolver{
		opt: opt,
	}, nil
}

type repositoryConnectionResolver struct {
	opt db.ReposListOptions

	// cache results because they is used by multiple fields
	once  sync.Once
	repos []*api.Repo
	err   error
}

func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*api.Repo, error) {
	r.once.Do(func() {
		opt2 := r.opt
		opt2.Limit++ // so we can detect if there is a next page
		r.repos, r.err = backend.Repos.List(ctx, opt2)
	})
	return r.repos, r.err
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*repositoryResolver, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*repositoryResolver, 0, len(repos))
	for i, repo := range repos {
		if i == r.opt.Limit {
			break
		}
		resolvers = append(resolvers, &repositoryResolver{repo: repo})
	}
	return resolvers, nil
}

func (r *repositoryConnectionResolver) TotalCount(ctx context.Context, args *struct {
	Precise bool
}) (countptr *int32, err error) {
	i32ptr := func(v int32) *int32 {
		return &v
	}

	if args.Precise {
		// Only site admins can perform precise counts, because it is a slow operation.
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
	}

	// Counting repositories is slow on Sourcegraph.com. Don't wait very long for an exact count.
	if !args.Precise && envvar.SourcegraphDotComMode() {
		if len(r.opt.Query) < 4 {
			return nil, nil
		}

		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		defer func() {
			if ctx.Err() == context.DeadlineExceeded {
				countptr = nil
				err = nil
			}
		}()
	}

	count, err := db.Repos.Count(ctx, r.opt)
	return i32ptr(int32(count)), err
}

func (r *repositoryConnectionResolver) PageInfo(ctx context.Context) (*pageInfo, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return &pageInfo{hasNextPage: len(repos) > r.opt.Limit}, nil
}

func (r *schemaResolver) SetRepositoryEnabled(ctx context.Context, args *struct {
	Repository graphql.ID
	Enabled    bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can enable/disable repositories, because it's a site-wide
	// and semi-destructive action.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := unmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}
	if err := db.Repos.SetEnabled(ctx, id, args.Enabled); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) DeleteRepository(ctx context.Context, args *struct {
	Repository graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete repositories, because it's a site-wide
	// and semi-destructive action.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := unmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}
	if err := db.Repos.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
