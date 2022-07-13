package repos

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestPackagesSource_GetRepo(t *testing.T) {
	ctx := context.Background()

	dummySrc := &dummyPackagesSource{}
	src := &PackagesSource{src: dummySrc, svc: &types.ExternalService{ID: 1, Kind: extsvc.KindGoPackages}}

	src.GetRepo(ctx, "go/github.com/sourcegraph-testing/go-repo-a")

	if !dummySrc.parsePackageFromRepoNameCalled {
		t.Fatalf("expected ParsePackageFromRepoName to be called, was not")
	}

	if !dummySrc.getPackageCalled {
		t.Fatalf("expected GetPackage to be called, was not")
	}
}

var _ packagesSource = &dummyPackagesSource{}
var _ packagesDownloadSource = &dummyPackagesSource{}

// dummyPackagesSource is a tiny shim around Go-specific methods to track when they're called.
type dummyPackagesSource struct {
	parseVersionedPackageFromConfiguration bool
	parsePackageFromRepoNameCalled         bool
	parsePackageFromNameCalled             bool
	getPackageCalled                       bool
}

// GetPackage implements packagesDownloadSource
func (d *dummyPackagesSource) GetPackage(ctx context.Context, name string) (reposource.Package, error) {
	d.getPackageCalled = true
	return reposource.ParseGoDependencyFromName(name)
}

func (d *dummyPackagesSource) ParseVersionedPackageFromConfiguration(dep string) (reposource.VersionedPackage, error) {
	d.parseVersionedPackageFromConfiguration = true
	return reposource.ParseGoVersionedPackage(dep)
}

func (d *dummyPackagesSource) ParsePackageFromName(name string) (reposource.Package, error) {
	d.parsePackageFromNameCalled = true
	return reposource.ParseGoDependencyFromName(name)
}

func (d *dummyPackagesSource) ParsePackageFromRepoName(repoName string) (reposource.Package, error) {
	d.parsePackageFromRepoNameCalled = true
	return reposource.ParseGoDependencyFromRepoName(repoName)
}
