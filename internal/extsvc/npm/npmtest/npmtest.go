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
	Packages map[string]*npm.PackageInfo
}

var _ npm.Client = &MockClient{}

func (m *MockClient) GetPackage(_ context.Context, name string) (info *npm.PackageInfo, err error) {
	info = m.Packages[name]
	if info == nil {
		return nil, errors.Newf("No version for package: %s", name)
	}
	return info, nil
}

func (m *MockClient) DoesDependencyExist(ctx context.Context, dep *reposource.NPMDependency) (exists bool, err error) {
	return m.Packages[dep.PackageManagerSyntax()] != nil, nil
}

func (m *MockClient) FetchTarball(_ context.Context, dep *reposource.NPMDependency) (closer io.ReadSeekCloser, err error) {
	info, ok := m.Packages[dep.PackageManagerSyntax()]
	if !ok {
		return nil, errors.Newf("Unknown dependency: %s", dep.PackageManagerSyntax())
	}

	version, ok := info.Versions[dep.PackageVersion()]
	if !ok {
		return nil, errors.Newf("Unknown dependency: %s", dep.PackageManagerSyntax())
	}

	return os.Open(version.Dist.TarballURL)
}
