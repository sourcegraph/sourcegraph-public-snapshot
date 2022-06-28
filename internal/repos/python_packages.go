package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/pypi"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewPythonPackagesSource returns a new PythonPackagesSource from the given external service.
func NewPythonPackagesSource(svc *types.ExternalService, cf *httpcli.Factory) (*DependenciesSource, error) {
	var c schema.PythonPackagesConnection
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
		scheme:     dependencies.PythonPackagesScheme,
		src:        &pythonPackagesSource{client: pypi.NewClient(svc.URN(), c.Urls, cli)},
	}, nil
}

type pythonPackagesSource struct {
	client *pypi.Client
}

var _ dependenciesSource = &pythonPackagesSource{}

func (s *pythonPackagesSource) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	_, err := s.client.Version(ctx, name, version)
	if err != nil {
		return nil, err
	}
	return reposource.NewPythonDependency(name, version), nil
}

func (pythonPackagesSource) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return reposource.ParsePythonDependency(dep)
}

func (pythonPackagesSource) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	return reposource.ParsePythonDependencyFromRepoName(repoName)
}
