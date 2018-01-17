package graphqlbackend

import (
	"context"
	"errors"
	"sync"
	"time"

	graphql "github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

func (r *siteResolver) Repositories(args *struct {
	connectionArgs
	Query    *string
	Cloning  bool
	Enabled  bool
	Disabled bool
}) (*repositoryConnectionResolver, error) {
	if args.Cloning && !args.Enabled {
		return nil, errors.New("mutually exclusive arguments: cloning, !enabled")
	}
	if args.Cloning && args.Disabled {
		return nil, errors.New("mutually exclusive arguments: cloning, disabled")
	}

	opt := db.ReposListOptions{
		Enabled:  args.Enabled,
		Disabled: args.Disabled,
	}
	if args.Query != nil {
		opt.Query = *args.Query
	}
	args.connectionArgs.set(&opt.ListOptions)
	return &repositoryConnectionResolver{
		opt:     opt,
		cloning: args.Cloning,
	}, nil
}

type repositoryConnectionResolver struct {
	opt     db.ReposListOptions
	cloning bool

	// cache results because they is used by multiple fields
	once  sync.Once
	repos []*sourcegraph.Repo
	err   error
}

func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*sourcegraph.Repo, error) {
	r.once.Do(func() {
		opt2 := r.opt
		opt2.PerPage++ // so we can detect if there is a next page
		repos, err := backend.Repos.List(ctx, opt2)
		r.err = err
		if repos != nil {
			r.repos = repos.Repos
		}
	})
	return r.repos, r.err
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*repositoryResolver, error) {
	if r.cloning {
		repos, err := r.resolveCloning(ctx)
		if err != nil {
			return nil, err
		}
		var l []*repositoryResolver
		for _, repoURI := range repos {
			if len(l) == r.opt.PerPageOrDefault() {
				break
			}
			repo, err := backend.Repos.GetByURI(ctx, repoURI)
			if err != nil {
				// Ignore ErrRepoNotFound, which might occur if the gitserver is shared by
				// multiple sites or has git repositories on it that have since been removed from
				// the frontend.
				if err != db.ErrRepoNotFound {
					return nil, err
				}
			}
			if repo != nil {
				l = append(l, &repositoryResolver{repo: repo})
			}
		}
		return l, nil
	}

	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*repositoryResolver, 0, len(repos))
	for i, repo := range repos {
		if i == r.opt.PerPageOrDefault() {
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

	if r.cloning {
		repos, err := r.resolveCloning(ctx)
		return i32ptr(int32(len(repos))), err
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
	if r.cloning {
		return nil, errors.New("pageInfo is not supported with cloning: true")
	}

	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return &pageInfo{hasNextPage: len(repos) > r.opt.PerPageOrDefault()}, nil
}

func (r *repositoryConnectionResolver) resolveCloning(ctx context.Context) (repos []string, err error) {
	if envvar.SourcegraphDotComMode() {
		return nil, nil
	}

	// 🚨 SECURITY: Only site admins can list cloning-in-progress repos, because there's no
	// good reason why non-site-admins should be able to do so.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	// First, find out what repos are currently being cloned.
	return gitserver.DefaultClient.ListCloning(ctx)
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
