package sharedresolvers

import (
	"context"
	"net/url"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepositoryResolver struct {
	logger    sglog.Logger
	hydration sync.Once
	err       error

	// Invariant: Name and ID of RepoMatch are always set and safe to use. They are
	// used to hydrate the inner repo, and should always be the same as the name and
	// id of the inner repo, but referring to the inner repo directly is unsafe
	// because it may cause a race during hydration.
	result.RepoMatch

	db database.DB

	// innerRepo may only contain ID and Name information.
	// To access any other repo information, use repo() instead.
	innerRepo *types.Repo
}

func NewRepositoryResolver(db database.DB, repo *types.Repo) *RepositoryResolver {
	// Protect against a nil repo
	var name api.RepoName
	var id api.RepoID
	if repo != nil {
		name = repo.Name
		id = repo.ID
	}

	return &RepositoryResolver{
		db:        db,
		innerRepo: repo,
		RepoMatch: result.RepoMatch{
			ID:   id,
			Name: name,
		},
		logger: log.Scoped("repositoryResolver", "resolve a specific repository").
			With(log.Object("repo",
				log.String("name", string(name)),
				log.Int32("id", int32(id)))),
	}
}

func (r *RepositoryResolver) ID() graphql.ID {
	return relay.MarshalID("Repository", r.RepoMatch.ID)
}

func (r *RepositoryResolver) Name() string {
	return string(r.RepoMatch.Name)
}

func (r *RepositoryResolver) Type(ctx context.Context) (*types.Repo, error) {
	return r.repo(ctx)
}

func (r *RepositoryResolver) CommitFromID(ctx context.Context, args *resolverstubs.RepositoryCommitArgs, commitID api.CommitID) (resolverstubs.GitCommitResolver, error) {
	return r.commitFromID(args, commitID)
}

func (r *RepositoryResolver) commitFromID(args *resolverstubs.RepositoryCommitArgs, commitID api.CommitID) (*GitCommitResolver, error) {
	resolver := NewGitCommitResolver(r, commitID)
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
	repo, err := r.repo(ctx)
	return repo.URI, err
}

func (r *RepositoryResolver) url() *url.URL {
	path := "/" + string(r.RepoMatch.Name)
	if r.Rev != "" {
		path += "@" + r.Rev
	}
	return &url.URL{Path: path}
}

// repo makes sure the repo is hydrated before returning it.
func (r *RepositoryResolver) repo(ctx context.Context) (*types.Repo, error) {
	err := r.hydrate(ctx)
	return r.innerRepo, err
}

func (r *RepositoryResolver) RepoName() api.RepoName {
	return r.RepoMatch.Name
}

func (r *RepositoryResolver) hydrate(ctx context.Context) error {
	r.hydration.Do(func() {
		// Repositories with an empty creation date were created using RepoName.ToRepo(),
		// they only contain ID and name information.
		if r.innerRepo != nil && !r.innerRepo.CreatedAt.IsZero() {
			return
		}

		r.logger.Debug("RepositoryResolver.hydrate", sglog.String("repo.ID", string(r.RepoMatch.ID)))

		var repo *types.Repo
		repo, r.err = r.db.Repos().Get(ctx, r.RepoMatch.ID)
		if r.err == nil {
			r.innerRepo = repo
		}
	})

	return r.err
}

func (r *RepositoryResolver) ExternalRepository() resolverstubs.ExternalRepositoryResolver {
	return NewExternalRepositoryResolver(r.innerRepo.ExternalRepo.ServiceID, r.innerRepo.ExternalRepo.ServiceType)
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
