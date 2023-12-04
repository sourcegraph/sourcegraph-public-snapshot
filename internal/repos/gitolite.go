package repos

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
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
	svc      *types.ExternalService
	conn     *schema.GitoliteConnection
	excluder repoExcluder

	// gitoliteLister allows us to list Gitlolite repos. In practice, we ask
	// gitserver to talk to gitolite because it holds the ssh keys required for
	// authentication.
	lister *gitserver.GitoliteLister
}

// NewGitoliteSource returns a new GitoliteSource from the given external service.
func NewGitoliteSource(ctx context.Context, svc *types.ExternalService, cf *httpcli.Factory) (*GitoliteSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.GitoliteConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}

	gitoliteDoer, err := cf.Doer(
		httpcli.NewMaxIdleConnsPerHostOpt(500),
		// The provided httpcli.Factory is one used for external services - however,
		// GitoliteSource asks gitserver to communicate to gitolite instead, so we
		// have to ensure that the actor transport used for internal clients is provided.
		httpcli.ActorTransportOpt,
	)
	if err != nil {
		return nil, err
	}

	var ex repoExcluder
	for _, r := range c.Exclude {
		ex.AddRule().
			Exact(r.Name).
			Pattern(r.Pattern)
	}
	if err := ex.RuleErrors(); err != nil {
		return nil, err
	}

	lister := gitserver.NewGitoliteLister(gitoliteDoer)

	return &GitoliteSource{
		svc:      svc,
		conn:     &c,
		lister:   lister,
		excluder: ex,
	}, nil
}

// CheckConnection at this point assumes availability and relies on errors returned
// from the subsequent calls. This is going to be expanded as part of issue #44683
// to actually only return true if the source can serve requests.
func (s *GitoliteSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns all Gitolite repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s *GitoliteSource) ListRepos(ctx context.Context, results chan SourceResult) {
	all, err := s.lister.ListRepos(ctx, s.conn.Host)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	for _, r := range all {
		repo := s.makeRepo(r)
		if !s.excludes(r, repo) {
			select {
			case <-ctx.Done():
				results <- SourceResult{Err: ctx.Err()}
				return
			case results <- SourceResult{Source: s, Repo: repo}:
			}
		}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s GitoliteSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s GitoliteSource) excludes(gr *gitolite.Repo, r *types.Repo) bool {
	return s.excluder.ShouldExclude(gr.Name) ||
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
		Private:  !s.svc.Unrestricted,
	}
}
