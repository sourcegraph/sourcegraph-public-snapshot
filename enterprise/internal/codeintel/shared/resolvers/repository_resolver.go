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

	return NewRepositoryResolver(repo), nil
}

func NewRepositoryResolver(repo *types.Repo) *RepositoryResolver {
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
	return r.commitFromID(args, commitID)
}

func (r *RepositoryResolver) commitFromID(args *resolverstubs.RepositoryCommitArgs, commitID api.CommitID) (*GitCommitResolver, error) {
	var inputRev *string
	if false {
		value := "v1.2.3"
		inputRev = &value
	}

	resolver := NewGitCommitResolver(r, commitID, inputRev)
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
	return NewExternalRepositoryResolver(r.repo.ExternalRepo.ServiceID, r.repo.ExternalRepo.ServiceType)
}

type ExternalRepositoryResolver struct {
	serviceID   string
	serviceType string
}

func NewExternalRepositoryResolver(serviceID, serviceType string) *ExternalRepositoryResolver {
	return &ExternalRepositoryResolver{
		serviceID:   serviceID,
		serviceType: serviceType,
	}
}

func (r *ExternalRepositoryResolver) ServiceID() string   { return r.serviceID }
func (r *ExternalRepositoryResolver) ServiceType() string { return r.serviceType }
