package repos

import (
	"context"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitea"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A GiteaSource yields repositories from a single Gitea connection configured
// in Sourcegraph via the external services configuration.
type GiteaSource struct {
	svc       *types.ExternalService
	client    *gitea.Client
	serviceID string

	url           *url.URL
	searchQueries []url.Values
}

// NewGiteaSource returns a new GiteaSource from the given external service.
func NewGiteaSource(ctx context.Context, svc *types.ExternalService, cf *httpcli.Factory) (*GiteaSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error: %s", svc.ID)
	}
	var c schema.GiteaConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}

	var searchQueries []url.Values
	for _, sq := range c.SearchQuery {
		v, err := parseSearchQuery(sq)
		if err != nil {
			return nil, errors.Wrapf(err, "external service id=%d config error: %s", svc.ID)
		}
		searchQueries = append(searchQueries, v)
	}

	url, err := url.Parse(c.Url)
	if err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error: %s", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	httpCli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	client, err := gitea.NewClient(svc.URN(), &c, httpCli)
	if err != nil {
		return nil, err
	}

	return &GiteaSource{
		svc:           svc,
		client:        client,
		serviceID:     extsvc.NormalizeBaseURL(client.URL).String(),
		url:           url,
		searchQueries: searchQueries,
	}, nil
}

func parseSearchQuery(s string) (url.Values, error) {
	switch s {
	case "", "all":
		s = "?"
	}

	if !strings.HasPrefix(s, "?") {
		return nil, errors.Errorf("gitea searchQuery string %q does not start with \"?\"", s)
	}

	params, err := url.ParseQuery(strings.TrimPrefix(s, "?"))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse the searchQuery %q", s)
	}

	// We allow admins to set pagination variables, but this is very much in
	// "I know what I am doing" territory. We expect these to be unset and as
	// such set good defaults to ensure we list all repos and avoid missing
	// some if the underlying collections changes as we paginate.
	defaults := map[string]string{
		"limit": "100",
		"sort":  "created",
		"order": "asc",

		// Since we don't sync permissions, default to not searching for
		// private repositories.
		"private": "false",
	}
	for k, v := range defaults {
		if !params.Has(k) {
			params.Set(k, v)
		}
	}

	return params, nil
}

// ListRepos returns all Gitea repositories configured with this GiteaSource's config.
func (s *GiteaSource) ListRepos(ctx context.Context, results chan SourceResult) {
	// TODO do we need to do de-duplication?

	var wg sync.WaitGroup
	for _, sq := range s.searchQueries {
		sq := sq
		wg.Add(1)
		go func() {
			defer wg.Done()

			it := s.client.ReposSearch(ctx, sq)
			for it.Next() {
				results <- SourceResult{Source: s, Repo: s.makeRepo(it.Current())}
			}

			if err := it.Err(); err != nil {
				results <- SourceResult{Source: s, Err: err}
			}
		}()
	}

	wg.Wait()
}

// ExternalServices returns a singleton slice containing the external service.
func (s *GiteaSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *GiteaSource) makeRepo(r *gitea.Repository) *types.Repo {
	urn := s.svc.URN()
	url := s.url.JoinPath(r.FullName)
	name := path.Join(url.Host, url.Path)

	return &types.Repo{
		Name:        api.RepoName(name),
		URI:         name,
		Description: r.Description,
		Fork:        r.Fork,
		Archived:    r.Archived,
		Stars:       r.Stars,
		// not r.Private. We intentionally do not sync privacy since we have
		// no way to enforce it yet.
		Private: false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          strconv.FormatInt(r.ID, 10),
			ServiceType: extsvc.TypeGitea,
			ServiceID:   s.serviceID,
		},
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: r.CloneURL,
			},
		},
		Metadata: r,
	}
}
