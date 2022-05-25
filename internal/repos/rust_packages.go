package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/crates"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewRustPackagesSource returns a new RustPackagesSource from the given external service.
func NewRustPackagesSource(svc *types.ExternalService, cf *httpcli.Factory) (*DependenciesSource, error) {
	var c schema.RustPackagesConnection
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
		scheme:     dependencies.RustPackagesScheme,
		src:        &rustPackagesSource{client: crates.NewClient(svc.URN(), c.Urls, cli)},
	}, nil
}

type rustPackagesSource struct {
	client *crates.Client
}

var _ dependenciesSource = &rustPackagesSource{}

func (s *rustPackagesSource) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	_, err := s.client.Version(ctx, name, version)
	if err != nil {
		return nil, err
	}
	return reposource.NewRustDependency(name, version), nil
}

func (rustPackagesSource) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return reposource.ParseRustDependency(dep)
}

func (rustPackagesSource) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	return reposource.ParseRustDependencyFromRepoName(repoName)
}
