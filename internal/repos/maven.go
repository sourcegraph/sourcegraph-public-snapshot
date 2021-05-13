package repos

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A MavenSource yields depots from a single Maven connection configured
// in Sourcegraph via the external services configuration.
type MavenSource struct {
	svc    *types.ExternalService
	config *schema.MavenConnection
}

// NewMavenSource returns a new MavenSource from the given external
// service.
func NewMavenSource(svc *types.ExternalService) (*MavenSource, error) {
	var c schema.MavenConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newMavenSource(svc, &c)
}

func newMavenSource(svc *types.ExternalService, c *schema.MavenConnection) (*MavenSource, error) {
	return &MavenSource{
		svc:    svc,
		config: c,
	}, nil
}

// ListRepos returns all Maven artifacts accessible to all connections
// configured in Sourcegraph via the external services configuration.
func (s MavenSource) ListRepos(ctx context.Context, results chan SourceResult) {
	if s.config.CloneAll {
		s.listAllRepos(ctx, results)
	} else {
		s.listDependentRepos(ctx, results)
	}
}

func (s MavenSource) listAllRepos(ctx context.Context, results chan SourceResult) {
	var orgs []string = coursier.Run("complete ''")
	for _, org := range orgs {
		var artifacts []string = coursier.Run("complete " + org)
		for _, artifact := range artifacts {
			results <- SourceResult{
				Source: s,
				Repo: s.makeRepo(org, artifact),
			}
		}
	}
}

func (s MavenSource) listDependentRepos(ctx context.Context, results chan SourceResult) {
	// select blah from blah
}

func (s MavenSource) makeRepo(org, artifact string) *types.Repo {
	artifactID := org + ":" + artifact
	artifactPath := org + "/" + artifact

	urn := s.svc.URN()
	cloneURL := url.URL{
		Scheme: "maven",
		Host: s.config.Url,
		Path: artifactPath,
	}
	return &types.Repo{
		Name: reposource.MavenRepoName(
			s.config.RepositoryPathPattern,
			artifactID,
		),
		URI: string(reposource.MavenRepoName(
			s.config.RepositoryPathPattern,
			artifactPath,
		)),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          artifactID,
			ServiceType: extsvc.TypeMaven,
			ServiceID:   s.config.Url,
		},
		// TODO
		Private: true,
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL.String(),
			},
		},
		// TODO
		Metadata: nil,
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s MavenSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}
