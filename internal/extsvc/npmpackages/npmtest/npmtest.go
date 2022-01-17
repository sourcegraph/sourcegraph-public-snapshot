package npmtest

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npmpackages/npm"
)

type MockClient struct {
	TarballMap              map[string]string
	DoesDependencyExistFunc func(_ context.Context, dep reposource.NPMDependency) (exists bool, err error)
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
			versions[dep.PackageManagerSyntax()] = struct{}{}
		}
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("No version for package: %s", pkg.PackageSyntax())
	}
	return versions, err
}

func (m *MockClient) DoesDependencyExist(ctx context.Context, dep reposource.NPMDependency) (exists bool, err error) {
	if m.DoesDependencyExistFunc != nil {
		return m.DoesDependencyExistFunc(ctx, dep)
	}
	_, found := m.TarballMap[dep.PackageManagerSyntax()]
	return found, nil
}

func (m *MockClient) FetchTarball(_ context.Context, dep reposource.NPMDependency) (closer io.ReadSeekCloser, err error) {
	path, found := m.TarballMap[dep.PackageManagerSyntax()]
	if !found {
		return nil, fmt.Errorf("Unknown dependency: %s", dep.PackageManagerSyntax())
	}
	return os.Open(path)
}
