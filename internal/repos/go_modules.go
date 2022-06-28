package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gomodproxy"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewGoModulesSource returns a new GoModulesSource from the given external service.
func NewGoModulesSource(svc *types.ExternalService, cf *httpcli.Factory) (*DependenciesSource, error) {
	var c schema.GoModulesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	return &DependenciesSource{
		svc:        svc,
		configDeps: c.Dependencies,
		scheme:     dependencies.GoModulesScheme,
		src: &goModulesSource{
			client: gomodproxy.NewClient(svc.URN(), c.Urls, cli),
		},
	}, nil
}

type goModulesSource struct {
	client *gomodproxy.Client
}

var _ dependenciesSource = &goModulesSource{}

func (s *goModulesSource) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	mod, err := s.client.GetVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}
	return reposource.NewGoDependency(*mod), nil
}

func (goModulesSource) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return reposource.ParseGoDependency(dep)
}

func (goModulesSource) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	return reposource.ParseGoDependencyFromRepoName(repoName)
}
