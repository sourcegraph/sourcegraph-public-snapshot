package graphqlbackend

import (
	"context"
	"errors"
	"strings"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	gogithub "github.com/sourcegraph/go-github/github"

	"sourcegraph.com/sourcegraph/sourcegraph/api"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

var GraphQLSchema *graphql.Schema

func init() {
	var err error
	GraphQLSchema, err = graphql.ParseSchema(api.Schema, &queryResolver{})
	if err != nil {
		panic(err)
	}
}

type nodeResolver interface {
	ID() graphql.ID
	ToRepository() (*repositoryResolver, bool)
	ToCommit() (*commitResolver, bool)
}

type nodeBase struct{}

func (*nodeBase) ToRepository() (*repositoryResolver, bool) {
	return nil, false
}

func (*nodeBase) ToCommit() (*commitResolver, bool) {
	return nil, false
}

type queryResolver struct{}

func (r *queryResolver) Root() *rootResolver {
	return &rootResolver{}
}

func (r *queryResolver) Node(ctx context.Context, args *struct{ ID graphql.ID }) (nodeResolver, error) {
	switch relay.UnmarshalKind(args.ID) {
	case "Repository":
		return repositoryByID(ctx, args.ID)
	case "Commit":
		return commitByID(ctx, args.ID)
	default:
		return nil, errors.New("invalid id")
	}
}

type rootResolver struct{}

func (r *rootResolver) Repository(ctx context.Context, args *struct{ URI string }) (*repositoryResolver, error) {
	if args.URI == "" {
		return nil, nil
	}

	repo, err := resolveRepo(ctx, args.URI)
	if err != nil {
		if err, ok := err.(legacyerr.Error); ok && err.Code == legacyerr.NotFound {
			return nil, nil
		}
		return nil, err
	}
	if err := backend.Repos.RefreshIndex(ctx, repo.URI); err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

func resolveRepo(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	res, err := backend.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{
		Path:   uri,
		Remote: true,
	})
	if err != nil {
		return nil, err
	}

	if res.Repo != 0 {
		return localstore.Repos.Get(ctx, res.Repo)
	}

	// repository exists only remotely, try to clone
	var op *sourcegraph.ReposCreateOp
	if res.RemoteRepo.Origin != nil {
		op = &sourcegraph.ReposCreateOp{
			Op: &sourcegraph.ReposCreateOp_Origin{
				Origin: res.RemoteRepo.Origin,
			},
		}
	} else {
		// Non-GitHub repositories.
		op = &sourcegraph.ReposCreateOp{
			Op: &sourcegraph.ReposCreateOp_New{
				New: &sourcegraph.ReposCreateOp_NewRepo{
					URI:           strings.Replace(res.RemoteRepo.HTTPCloneURL, "https://", "", -1),
					CloneURL:      res.RemoteRepo.HTTPCloneURL,
					DefaultBranch: "master",
					Mirror:        true,
				},
			},
		}
	}

	repo, err := backend.Repos.Create(ctx, op)
	if err != nil {
		if err, ok := err.(legacyerr.Error); ok && err.Code == legacyerr.AlreadyExists { // race condition
			res, err := backend.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{
				Path: uri,
			})
			if err != nil {
				return nil, err
			}
			return localstore.Repos.Get(ctx, res.Repo)
		}
		return nil, err
	}
	return repo, nil
}

func (r *rootResolver) RemoteRepositories(ctx context.Context) ([]*repositoryResolver, error) {
	reposList, err := backend.Repos.List(ctx, &sourcegraph.RepoListOptions{
		RemoteOnly: true,
	})

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

func (r *rootResolver) RemoteStarredRepositories(ctx context.Context) ([]*repositoryResolver, error) {
	starredRepos, err := backend.Repos.ListStarredRepos(ctx, &gogithub.ActivityListStarredOptions{})
	if err != nil {
		return nil, err
	}

	var s []*repositoryResolver
	for _, repo := range starredRepos.Repos {
		s = append(s, &repositoryResolver{
			repo: repo,
		})
	}

	return s, nil
}
