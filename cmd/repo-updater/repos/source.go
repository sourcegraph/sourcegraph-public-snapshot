package repos

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A Sourcer yields Sources whose Repos should be synced.
type Sourcer interface {
	ListSources(ctx context.Context, kinds ...string) ([]Source, error)
}

// ExternalServicesSourcer converts each code host connection configured via external services
// to a Source that yields Repos. Each invocation of ListSources may yield different Sources
// depending on what the user configured at a given point in time.
type ExternalServicesSourcer struct {
	st Store
	cf httpcli.Factory
}

// NewExternalServicesSourcer returns a new ExternalServicesSourcer with the given Store.
func NewExternalServicesSourcer(st Store, cf httpcli.Factory) *ExternalServicesSourcer {
	return &ExternalServicesSourcer{st: st, cf: cf}
}

// ListSources lists all configured repository yielding Sources of the given kinds,
// based on the code host connections configured via external services in the frontend API.
func (s ExternalServicesSourcer) ListSources(ctx context.Context, kinds ...string) ([]Source, error) {
	svcs, err := s.st.ListExternalServices(ctx, kinds...)
	if err != nil {
		return nil, err
	}

	srcs := make([]Source, 0, len(svcs)+1)
	errs := new(multierror.Error)
	for _, svc := range svcs {
		if src, err := NewSource(svc, s.cf); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			srcs = append(srcs, src)
		}
	}

	if !includesGitHubDotComSource(srcs) {
		// add a GitHub.com source by default, to support navigating to URL
		// paths like /github.com/foo/bar to auto-add that repository. This
		// source returns nothing for ListRepos. However, in the future we
		// intend to use it in repoLookup.
		src, err := NewGithubDotComSource(s.cf)
		srcs, errs = append(srcs, src), multierror.Append(errs, err)
	}

	return srcs, errs.ErrorOrNil()
}

// NewSource returns a repository yielding Source from the given ExternalService configuration.
func NewSource(svc *ExternalService, cf httpcli.Factory) (Source, error) {
	switch svc.Kind {
	case "GITHUB":
		return NewGithubSource(svc, cf)
	default:
		panic(fmt.Sprintf("source not implemented for external service kind %q", svc.Kind))
	}
}

func includesGitHubDotComSource(srcs []Source) bool {
	for _, src := range srcs {
		if gs, ok := src.(*GithubSource); !ok {
			continue
		} else if u, err := url.Parse(gs.conn.config.Url); err != nil {
			continue
		} else if strings.HasSuffix(u.Hostname(), "github.com") {
			return true
		}
	}
	return false
}

// A Source yields repositories to be stored and analysed by Sourcegraph.
// Successive calls to its ListRepos method may yield different results.
type Source interface {
	// TODO(keegancsmith) document contract of ListRepos + contract tests
	ListRepos(context.Context) ([]*Repo, error)
}

// A GithubSource yields repositories from a single Github connection configured
// in Sourcegraph via the external services configuration.
type GithubSource struct {
	svc  *ExternalService
	conn *githubConnection
}

// NewGithubSource returns a new GithubSource from the given external service.
func NewGithubSource(svc *ExternalService, cf httpcli.Factory) (*GithubSource, error) {
	var c schema.GitHubConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newGithubSource(svc, &c, cf)
}

// NewGithubDotComSource returns a GithubSource for github.com, meant to be added
// to the list of sources in Sourcer when one isn't already configured in order to
// support navigating to URL paths like /github.com/foo/bar to auto-add that repository.
func NewGithubDotComSource(cf httpcli.Factory) (*GithubSource, error) {
	svc := ExternalService{Kind: "GITHUB"}
	return newGithubSource(&svc, &schema.GitHubConnection{
		RepositoryQuery:             []string{"none"}, // don't try to list all repositories during syncs
		Url:                         "https://github.com",
		InitialRepositoryEnablement: true,
	}, cf)
}

func newGithubSource(svc *ExternalService, c *schema.GitHubConnection, cf httpcli.Factory) (*GithubSource, error) {
	conn, err := newGitHubConnection(c, cf)
	if err != nil {
		return nil, err
	}
	return &GithubSource{svc: svc, conn: conn}, nil
}

// ListRepos returns all Github repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s GithubSource) ListRepos(ctx context.Context) (repos []*Repo, err error) {
	rs, err := s.conn.listAllRepositories(ctx)
	for _, r := range rs {
		repos = append(repos, githubRepoToRepo(s.svc, r, s.conn))
	}
	return repos, err
}

func githubRepoToRepo(
	svc *ExternalService,
	ghrepo *github.Repository,
	conn *githubConnection,
) *Repo {
	urn := externalServiceURN(svc)
	return &Repo{
		Name:         string(githubRepositoryToRepoPath(conn, ghrepo)),
		ExternalRepo: *github.ExternalRepoSpec(ghrepo, *conn.baseURL),
		Description:  ghrepo.Description,
		Fork:         ghrepo.IsFork,
		Archived:     ghrepo.IsArchived,
		Sources: map[string]*SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: conn.authenticatedRemoteURL(ghrepo),
			},
		},
		Metadata: ghrepo,
	}
}

func externalServiceURN(svc *ExternalService) string {
	return "extsvc:" + strconv.FormatInt(svc.ID, 10)
}
