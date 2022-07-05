package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewNpmPackagesSource returns a new PackagesSource from the given external
// service.
func NewNpmPackagesSource(svc *types.ExternalService, cf *httpcli.Factory) (*PackagesSource, error) {
	var c schema.NpmPackagesConnection
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
		scheme:     dependencies.NpmPackagesScheme,
		/* depsSvc initialized in SetDependenciesService */
		src: &npmPackagesSource{
			client: npm.NewHTTPClient(svc.URN(), c.Registry, c.Credentials, cli),
		},
	}, nil
}

var _ packagesSource = &npmPackagesSource{}

type npmPackagesSource struct {
	client npm.Client
}

func (npmPackagesSource) ParsePackageVersionFromConfiguration(dep string) (reposource.PackageVersion, error) {
	return reposource.ParseNpmPackageVersion(dep)
}

func (npmPackagesSource) ParsePackageFromRepoName(repoName string) (reposource.Package, error) {
	pkg, err := reposource.ParseNpmPackageFromRepoURL(repoName)
	if err != nil {
		return nil, err
	}
	return &reposource.NpmPackageVersion{NpmPackageName: pkg}, nil
}

func (s *npmPackagesSource) Get(ctx context.Context, name, version string) (reposource.PackageVersion, error) {
	parsedDbPackage, err := reposource.ParseNpmPackageFromPackageSyntax(name)
	if err != nil {
		return nil, err
	}

	dep := &reposource.NpmPackageVersion{NpmPackageName: parsedDbPackage, Version: version}

	info, err := s.client.GetDependencyInfo(ctx, dep)
	if err != nil {
		return nil, err
	}

	dep.PackageDescription = info.Description
	dep.TarballURL = info.Dist.TarballURL

	return dep, nil
}
