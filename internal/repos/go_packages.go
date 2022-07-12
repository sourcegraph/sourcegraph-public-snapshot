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

// NewGoPackagesSource returns a new GoModulesSource from the given external service.
func NewGoPackagesSource(svc *types.ExternalService, cf *httpcli.Factory) (*PackagesSource, error) {
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
		scheme:     dependencies.GoPackagesScheme,
		src: &goPackagesSource{
			client: gomodproxy.NewClient(svc.URN(), c.Urls, cli),
		},
	}, nil
}

type goPackagesSource struct {
	client *gomodproxy.Client
}

var _ packagesSource = &goPackagesSource{}

func (s *goPackagesSource) Get(ctx context.Context, name, version string) (reposource.VersionedPackage, error) {
	mod, err := s.client.GetVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}
	return reposource.NewGoVersionedPackage(*mod), nil
}

func (goPackagesSource) ParseVersionedPackageFromConfiguration(dep string) (reposource.VersionedPackage, error) {
	return reposource.ParseGoVersionedPackage(dep)
}

func (goPackagesSource) ParsePackageFromName(name string) (reposource.Package, error) {
	return reposource.ParseGoDependencyFromName(name)
}

func (goPackagesSource) ParsePackageFromRepoName(repoName string) (reposource.Package, error) {
	return reposource.ParseGoDependencyFromRepoName(repoName)
}
