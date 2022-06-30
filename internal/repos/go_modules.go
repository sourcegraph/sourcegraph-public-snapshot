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
func NewGoModulesSource(svc *types.ExternalService, cf *httpcli.Factory) (*PackagesSource, error) {
	var c schema.GoModulesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	return &PackagesSource{
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

var _ packagesSource = &goModulesSource{}

func (s *goModulesSource) Get(ctx context.Context, name, version string) (reposource.PackageVersion, error) {
	mod, err := s.client.GetVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}
	return reposource.NewGoPackageVersion(*mod), nil
}

func (goModulesSource) ParsePackageVersionFromConfiguration(dep string) (reposource.PackageVersion, error) {
	return reposource.ParseGoPackageVersion(dep)
}

func (goModulesSource) ParsePackageFromRepoName(repoName string) (reposource.Package, error) {
	return reposource.ParseGoDependencyFromRepoName(repoName)
}
