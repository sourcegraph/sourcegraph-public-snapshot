package repos

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A GitoliteSource yields repositories from a single Gitolite connection configured
// in Sourcegraph via the external services configuration.
type GitoliteSource struct {
	svc  *types.ExternalService
	conn *schema.GitoliteConnection
	// We ask gitserver to talk to gitolite because it holds the ssh keys
	// required for authentication.
	cli     *gitserver.ClientImplementor
	exclude excludeFunc
}

// NewGitoliteSource returns a new GitoliteSource from the given external service.
func NewGitoliteSource(db database.DB, svc *types.ExternalService, cf *httpcli.Factory) (*GitoliteSource, error) {
	var c schema.GitoliteConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}

	gitserverDoer, err := cf.Doer(
		httpcli.NewMaxIdleConnsPerHostOpt(500),
		// The provided httpcli.Factory is one used for external services - however,
		// GitoliteSource asks gitserver to communicate to gitolite instead, so we
		// have to ensure that the actor transport used for internal clients is provided.
		httpcli.ActorTransportOpt)
	if err != nil {
		return nil, err
	}

	var eb excludeBuilder
	for _, r := range c.Exclude {
		eb.Exact(r.Name)
		eb.Pattern(r.Pattern)
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}

	gitserverClient := gitserver.NewClient(db)
	gitserverClient.HTTPClient = gitserverDoer

	return &GitoliteSource{
		svc:     svc,
		conn:    &c,
		cli:     gitserverClient,
		exclude: exclude,
	}, nil
}

// ListRepos returns all Gitolite repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s *GitoliteSource) ListRepos(ctx context.Context, results chan SourceResult) {
	all, err := s.cli.ListGitolite(ctx, s.conn.Host)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	for _, r := range all {
		repo := s.makeRepo(r)
		if !s.excludes(r, repo) {
			results <- SourceResult{Source: s, Repo: repo}
		}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s GitoliteSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s GitoliteSource) excludes(gr *gitolite.Repo, r *types.Repo) bool {
	return s.exclude(gr.Name) ||
		strings.ContainsAny(string(r.Name), "\\^$|()[]*?{},")
}

func (s GitoliteSource) makeRepo(repo *gitolite.Repo) *types.Repo {
	urn := s.svc.URN()
	name := string(reposource.GitoliteRepoName(s.conn.Prefix, repo.Name))
	return &types.Repo{
		Name:         api.RepoName(name),
		URI:          name,
		ExternalRepo: gitolite.ExternalRepoSpec(repo, gitolite.ServiceID(s.conn.Host)),
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: repo.URL,
			},
		},
		Metadata: repo,
	}
}
