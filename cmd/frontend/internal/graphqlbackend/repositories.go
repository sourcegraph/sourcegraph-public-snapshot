package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

func (r *siteResolver) Repositories(args *struct {
	connectionArgs
	Query           *string
	Cloning         bool
	IncludeDisabled bool
}) (*repositoryConnectionResolver, error) {
	if args.Cloning && args.IncludeDisabled {
		return nil, errors.New("mutually exclusive arguments: cloning, includeDisabled")
	}

	opt := db.ReposListOptions{
		IncludeDisabled: args.IncludeDisabled,
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

	reposList, err := backend.Repos.List(ctx, &r.opt)
	if err != nil {
		return nil, err
	}

	var l []*repositoryResolver
	for _, repo := range reposList.Repos {
		l = append(l, &repositoryResolver{
			repo: repo,
		})
	}
	return l, nil
}

func (r *repositoryConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	if r.cloning {
		repos, err := r.resolveCloning(ctx)
		return int32(len(repos)), err
	}

	count, err := db.Repos.Count(ctx, r.opt)
	return int32(count), err
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
