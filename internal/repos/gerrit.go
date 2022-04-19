package repos

import (
	"context"
	"path"
	"sort"

	"github.com/goware/urlx"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A GerritSource yields repositories from a single Gerrit connection configured
// in Sourcegraph via the external services configuration.
type GerritSource struct {
	svc       *types.ExternalService
	cli       *gerrit.Client
	serviceID string
	perPage   int
}

// NewGerritSource returns a new GerritSource from the given external service.
func NewGerritSource(svc *types.ExternalService, cf *httpcli.Factory) (*GerritSource, error) {
	var c schema.GerritConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	httpCli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	cli, err := gerrit.NewClient(svc.URN(), &c, httpCli)
	if err != nil {
		return nil, err
	}

	return &GerritSource{
		svc:       svc,
		cli:       cli,
		serviceID: extsvc.NormalizeBaseURL(cli.URL).String(),
		perPage:   100,
	}, nil
}

// ListRepos returns all Gerrit repositories configured with this GerritSource's config.
func (s *GerritSource) ListRepos(ctx context.Context, results chan SourceResult) {
	args := gerrit.ListProjectsArgs{
		Cursor: &gerrit.Pagination{PerPage: s.perPage, Page: 1},
	}

	for {
		page, nextPage, err := s.cli.ListProjects(ctx, args)
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		// Unfortunately, because Gerrit API responds with a map, we have to sort it to maintain proper ordering
		pageAsMap := map[string]*gerrit.Project(*page)
		pageKeySlice := make([]string, 0, len(pageAsMap))

		for p := range pageAsMap {
			pageKeySlice = append(pageKeySlice, p)
		}

		sort.Strings(pageKeySlice)

		for _, p := range pageKeySlice {
			repo, err := s.makeRepo(p, pageAsMap[p])
			if err != nil {
				results <- SourceResult{Source: s, Err: err}
				return
			}
			results <- SourceResult{Source: s, Repo: repo}
		}

		if !nextPage {
			break
		}

		args.Cursor.Page++
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *GerritSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *GerritSource) makeRepo(projectName string, p *gerrit.Project) (*types.Repo, error) {
	urn := s.svc.URN()

	fullURL, err := urlx.Parse(s.cli.URL.String() + projectName)
	if err != nil {
		return nil, err
	}

	name := path.Join(fullURL.Host, fullURL.Path)
	return &types.Repo{
		Name:        api.RepoName(name),
		URI:         name,
		Description: p.Description,
		Fork:        p.Parent != "",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          p.ID,
			ServiceType: extsvc.TypeGerrit,
			ServiceID:   s.serviceID,
		},
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: fullURL.String(),
			},
		},
		Metadata: p,
	}, nil
}
