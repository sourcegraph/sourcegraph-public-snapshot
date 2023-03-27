package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewJVMPackagesSource returns a new MavenSource from the given external
// service.
func NewJVMPackagesSource(ctx context.Context, svc *types.ExternalService) (*PackagesSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.JVMPackagesConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	return &PackagesSource{
		svc:        svc,
		configDeps: c.Maven.Dependencies,
		scheme:     dependencies.JVMPackagesScheme,
		src:        &jvmPackagesSource{config: &c},
	}, nil
}

// A jvmPackagesSource creates git repositories from `*-sources.jar` files of
// published Maven dependencies from the JVM ecosystem.
type jvmPackagesSource struct {
	config *schema.JVMPackagesConnection
}

var _ packagesSource = &jvmPackagesSource{}

// Commented out as importing 'internal/extsvc/jvmpackages/coursier' here includes it in the frontend and repo-updater binaries.
// We don't want that due to the side-effects of importing that package.
/* func (s *jvmPackagesSource) Get(ctx context.Context, name, version string) (reposource.VersionedPackage, error) {
	mavenDependency, err := reposource.ParseMavenVersionedPackage(name + ":" + version)
	if err != nil {
		return nil, err
	}

	err = coursier.Exists(ctx, s.config, mavenDependency)
	if err != nil {
		return nil, err
	}
	return mavenDependency, nil
} */

func (jvmPackagesSource) ParseVersionedPackageFromConfiguration(dep string) (reposource.VersionedPackage, error) {
	return reposource.ParseMavenVersionedPackage(dep)
}

func (jvmPackagesSource) ParsePackageFromName(name reposource.PackageName) (reposource.Package, error) {
	return reposource.ParseMavenPackageFromName(name)
}

func (jvmPackagesSource) ParsePackageFromRepoName(repoName api.RepoName) (reposource.Package, error) {
	return reposource.ParseMavenPackageFromRepoName(repoName)
}
