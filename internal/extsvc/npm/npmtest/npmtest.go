package npmtest

import (
	"context"
	"io"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MockClient struct {
	// TarballMap is a map from dependency (in package manager syntax)
	// to optional tarball paths.
	TarballMap map[string]string
}

var _ npm.Client = &MockClient{}

func (m *MockClient) AvailablePackageVersions(_ context.Context, pkg reposource.NPMPackage) (versions map[string]struct{}, err error) {
	versions = map[string]struct{}{}
	for dep := range m.TarballMap {
		dep, err := reposource.ParseNPMDependency(dep)
		if err != nil {
			return versions, err
		}
		if pkg == dep.Package {
			versions[dep.Version] = struct{}{}
		}
	}
	if len(versions) == 0 {
		return nil, errors.Newf("No version for package: %s", pkg.PackageSyntax())
	}
	return versions, err
}

func (m *MockClient) DoesDependencyExist(ctx context.Context, dep reposource.NPMDependency) (exists bool, err error) {
	_, found := m.TarballMap[dep.PackageManagerSyntax()]
	return found, nil
}

func (m *MockClient) FetchTarball(_ context.Context, dep reposource.NPMDependency) (closer io.ReadSeekCloser, err error) {
	path, found := m.TarballMap[dep.PackageManagerSyntax()]
	if !found {
		return nil, errors.Newf("Unknown dependency: %s", dep.PackageManagerSyntax())
	}
	return os.Open(path)
}
