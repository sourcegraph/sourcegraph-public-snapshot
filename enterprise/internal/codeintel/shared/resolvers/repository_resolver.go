package sharedresolvers

import (
	"context"
	"net/url"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepositoryResolver struct {
	repo *types.Repo
}

func NewRepositoryFromID(ctx context.Context, repoStore database.RepoStore, id int) (*RepositoryResolver, error) {
	repo, err := repoStore.Get(ctx, api.RepoID(id))
	if err != nil {
		return nil, err
	}

	return newRepositoryResolver(repo), nil
}

func newRepositoryResolver(repo *types.Repo) *RepositoryResolver {
	return &RepositoryResolver{repo: repo}
}

func (r *RepositoryResolver) ID() graphql.ID {
	return relay.MarshalID("Repository", r.repo.ID)
}

func (r *RepositoryResolver) Name() string {
	return string(r.repo.Name)
}

func (r *RepositoryResolver) Type(ctx context.Context) (*types.Repo, error) {
	return r.repo, nil
}

func (r *RepositoryResolver) CommitFromID(ctx context.Context, args *resolverstubs.RepositoryCommitArgs, commitID api.CommitID) (resolverstubs.GitCommitResolver, error) {
	resolver := newGitCommitResolver(r, commitID)
	if args.InputRevspec != nil {
		resolver.inputRev = args.InputRevspec
	} else {
		resolver.inputRev = &args.Rev
	}
	return resolver, nil
}

func (r *RepositoryResolver) URL() string {
	return r.url().String()
}

func (r *RepositoryResolver) URI(ctx context.Context) (string, error) {
	return r.repo.URI, nil
}

func (r *RepositoryResolver) url() *url.URL {
	path := "/" + string(r.repo.Name)
	return &url.URL{Path: path}
}

func (r *RepositoryResolver) RepoName() api.RepoName {
	return r.repo.Name
}

func (r *RepositoryResolver) ExternalRepository() resolverstubs.ExternalRepositoryResolver {
	return newExternalRepositoryResolver(r.repo.ExternalRepo.ServiceID, r.repo.ExternalRepo.ServiceType)
}

type externalRepositoryResolver struct {
	serviceID   string
	serviceType string
}

func newExternalRepositoryResolver(serviceID, serviceType string) *externalRepositoryResolver {
	return &externalRepositoryResolver{
		serviceID:   serviceID,
		serviceType: serviceType,
	}
}

func (r *externalRepositoryResolver) ServiceID() string   { return r.serviceID }
func (r *externalRepositoryResolver) ServiceType() string { return r.serviceType }
