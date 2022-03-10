package npmtest

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MockClient struct {
	Packages map[string]*npm.PackageInfo
	Tarballs map[string][]byte
}

func NewMockClient(t testing.TB, deps ...string) *MockClient {
	t.Helper()

	packages := map[string]*npm.PackageInfo{}
	for _, dep := range deps {
		d, err := reposource.ParseNPMDependency(dep)
		if err != nil {
			t.Fatal(err)
		}

		name := d.PackageSyntax()
		info := packages[name]

		if info == nil {
			info = &npm.PackageInfo{Versions: map[string]*npm.DependencyInfo{}}
			packages[name] = info
		}

		info.Description = name + " description"
		version := info.Versions[d.Version]
		if version == nil {
			version = &npm.DependencyInfo{}
			info.Versions[d.Version] = version
		}
	}

	return &MockClient{Packages: packages}
}

var _ npm.Client = &MockClient{}

func (m *MockClient) GetPackageInfo(ctx context.Context, pkg *reposource.NPMPackage) (info *npm.PackageInfo, err error) {
	info = m.Packages[pkg.PackageSyntax()]
	if info == nil {
		return nil, errors.Newf("No version for package: %s", pkg)
	}
	return info, nil
}

func (m *MockClient) DoesDependencyExist(ctx context.Context, dep *reposource.NPMDependency) (exists bool, err error) {
	pkg := m.Packages[dep.PackageSyntax()]
	if pkg == nil {
		return false, nil
	}
	return pkg.Versions[dep.Version] != nil, nil
}

func (m *MockClient) FetchTarball(_ context.Context, dep *reposource.NPMDependency) (io.ReadCloser, error) {
	info, ok := m.Packages[dep.PackageSyntax()]
	if !ok {
		return nil, errors.Newf("Unknown dependency: %s", dep.PackageManagerSyntax())
	}

	version, ok := info.Versions[dep.Version]
	if !ok {
		return nil, errors.Newf("Unknown dependency: %s", dep.PackageManagerSyntax())
	}

	tgz, ok := m.Tarballs[version.Dist.TarballURL]
	if !ok {
		return nil, errors.Newf("no tarball for %s", version.Dist.TarballURL)
	}

	return io.NopCloser(bytes.NewReader(tgz)), nil
}
