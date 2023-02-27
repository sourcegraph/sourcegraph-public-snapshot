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
	Packages map[reposource.PackageName]*npm.PackageInfo
	Tarballs map[string]io.Reader
}

func NewMockClient(t testing.TB, deps ...string) *MockClient {
	t.Helper()

	packages := map[reposource.PackageName]*npm.PackageInfo{}
	for _, dep := range deps {
		d, err := reposource.ParseNpmVersionedPackage(dep)
		if err != nil {
			t.Fatal(err)
		}

		name := d.PackageSyntax()
		info := packages[name]

		if info == nil {
			info = &npm.PackageInfo{Versions: map[string]*npm.DependencyInfo{}}
			packages[name] = info
		}

		info.Description = string(name) + " description"
		version := info.Versions[d.Version]
		if version == nil {
			version = &npm.DependencyInfo{}
			info.Versions[d.Version] = version
		}
	}

	return &MockClient{Packages: packages}
}

var _ npm.Client = &MockClient{}

func (m *MockClient) GetPackageInfo(ctx context.Context, pkg *reposource.NpmPackageName) (info *npm.PackageInfo, err error) {
	info = m.Packages[pkg.PackageSyntax()]
	if info == nil {
		return nil, errors.Newf("package not found: %s", pkg.PackageSyntax())
	}
	return info, nil
}

func (m *MockClient) GetDependencyInfo(ctx context.Context, dep *reposource.NpmVersionedPackage) (info *npm.DependencyInfo, err error) {
	pkg, err := m.GetPackageInfo(ctx, dep.NpmPackageName)
	if err != nil {
		return nil, err
	}

	info = pkg.Versions[dep.Version]
	if info == nil {
		return nil, errors.Newf("package version not found: %s", dep.VersionedPackageSyntax())
	}

	return info, nil
}

func (m *MockClient) FetchTarball(_ context.Context, dep *reposource.NpmVersionedPackage) (io.ReadCloser, error) {
	info, ok := m.Packages[dep.PackageSyntax()]
	if !ok {
		return nil, errors.Newf("Unknown dependency: %s", dep.VersionedPackageSyntax())
	}

	version, ok := info.Versions[dep.Version]
	if !ok {
		return nil, errors.Newf("Unknown dependency: %s", dep.VersionedPackageSyntax())
	}

	tgz, ok := m.Tarballs[version.Dist.TarballURL]
	if !ok {
		return nil, errors.Newf("no tarball for %s", version.Dist.TarballURL)
	}

	// tee to a new buffer, to avoid EOF from reading the same one multiple times
	var newTgz bytes.Buffer
	tee := io.TeeReader(tgz, &newTgz)
	m.Tarballs[version.Dist.TarballURL] = &newTgz

	return io.NopCloser(tee), nil
}
