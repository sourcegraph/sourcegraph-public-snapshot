package server

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewNpmPackagesSyncer create a new NpmPackageSyncer. If customClient is nil,
// the client for the syncer is configured based on the connection parameter.
func NewNpmPackagesSyncer(
	connection schema.NpmPackagesConnection,
	svc *dependencies.Service,
	customClient npm.Client,
	urn string,
) VCSSyncer {
	var client = customClient
	if client == nil {
		client = npm.NewHTTPClient(urn, connection.Registry, connection.Credentials)
	}

	placeholder, err := reposource.ParseNpmDependency("@sourcegraph/placeholder@1.0.0")
	if err != nil {
		panic(fmt.Sprintf("expected placeholder package to parse but got %v", err))
	}

	return &vcsDependenciesSyncer{
		typ:         "npm_packages",
		scheme:      dependencies.NpmPackagesScheme,
		placeholder: placeholder,
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &npmPackagesSyncer{client: client},
	}
}

type npmPackagesSyncer struct {
	// The client to use for making queries against npm.
	client npm.Client
}

func (npmPackagesSyncer) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return reposource.ParseNpmDependency(dep)
}

func (npmPackagesSyncer) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	pkg, err := reposource.ParseNpmPackageFromRepoURL(repoName)
	if err != nil {
		return nil, err
	}
	return &reposource.NpmDependency{NpmPackage: pkg}, nil
}

func (s *npmPackagesSyncer) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	dep, err := reposource.ParseNpmDependency(name + "@" + version)
	if err != nil {
		return nil, err
	}

	info, err := s.client.GetDependencyInfo(ctx, dep)
	if err != nil {
		return nil, err
	}

	dep.TarballURL = info.Dist.TarballURL
	return dep, nil
}

func (s *npmPackagesSyncer) Download(ctx context.Context, dir string, dep reposource.PackageDependency) error {
	tgz, err := npm.FetchSources(ctx, s.client, dep.(*reposource.NpmDependency))
	if err != nil {
		return errors.Wrap(err, "fetch tarball")
	}
	defer tgz.Close()

	if err = unpack.DecompressTgz(tgz, dir); err != nil {
		return errors.Wrapf(err, "failed to decompress gzipped tarball for %s", dep.PackageManagerSyntax())
	}

	return nil
}
