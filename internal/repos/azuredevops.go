package repos

import (
	"context"
	"github.com/goware/urlx"
	"path"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// A ADOSource yields repositories from a single ADO connection configured
// in Sourcegraph via the external services configuration.
type ADOSource struct {
	svc       *types.ExternalService
	cli       *azuredevops.Client
	serviceID string
	perPage   int
	config    azuredevops.ADOConnection
}

// NewADOSource returns a new ADOSource from the given external service.
func NewADOSource(ctx context.Context, svc *types.ExternalService, cf *httpcli.Factory) (*ADOSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c azuredevops.ADOConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	httpCli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	cli, err := azuredevops.NewClient(svc.URN(), &c, httpCli)
	if err != nil {
		return nil, err
	}

	return &ADOSource{
		svc:       svc,
		cli:       cli,
		serviceID: extsvc.NormalizeBaseURL(cli.URL).String(),
		perPage:   100,
		config:    c,
	}, nil
}

// CheckConnection at this point assumes availability and relies on errors returned
// from the subsequent calls. This is going to be expanded as part of issue #44683
// to actually only return true if the source can serve requests.
func (s *ADOSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns all ADO repositories configured with this ADOSource's config.
func (s *ADOSource) ListRepos(ctx context.Context, results chan SourceResult) {

	for _, project := range s.config.Projects {
		s.processReposFromProjectOrOrg(ctx, project, results)
	}

	for _, org := range s.config.Orgs {
		s.processReposFromProjectOrOrg(ctx, org, results)
	}
}

func (s *ADOSource) processReposFromProjectOrOrg(ctx context.Context, name string, results chan SourceResult) {
	repos, err := s.cli.ListRepositoriesByProjectOrOrg(ctx, azuredevops.ListRepositoriesByProjectOrOrgArgs{
		ProjectOrOrgName: name,
	})
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	if repos == nil {
		results <- SourceResult{Source: s, Err: errors.New("got an empty response from ADO.")}
		return
	}

	for _, repo := range repos.Value {
		repo, err := s.makeRepo(repo)
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}
		results <- SourceResult{Source: s, Repo: repo}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *ADOSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *ADOSource) makeRepo(p azuredevops.RepositoriesValue) (*types.Repo, error) {
	urn := s.svc.URN()

	fullURL, err := urlx.Parse(s.cli.URL.String() + p.Name)
	if err != nil {
		return nil, err
	}

	name := path.Join(fullURL.Host, fullURL.Path)
	return &types.Repo{
		Name: api.RepoName(name),
		URI:  name,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          p.ID,
			ServiceType: extsvc.TypeADO,
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
