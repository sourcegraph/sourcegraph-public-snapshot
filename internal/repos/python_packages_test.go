package repos

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPythonPackagesSource_ListRepos(t *testing.T) {
	ctx := context.Background()
	depsSvc := testDependenciesService(ctx, t, []dependencies.MinimalPackageRepoRef{
		{
			Scheme: dependencies.PythonPackagesScheme,
			Name:   "requests",
			Versions: []dependencies.MinimalPackageRepoRefVersion{
				{Version: "2.27.1"}, // test deduplication with version from config
				{Version: "2.27.2"}, // test multiple versions of the same module
			},
		},
		{
			Scheme:   dependencies.PythonPackagesScheme,
			Name:     "numpy",
			Versions: []dependencies.MinimalPackageRepoRefVersion{{Version: "1.22.3"}},
		},
		{
			Scheme:   dependencies.PythonPackagesScheme,
			Name:     "lofi",
			Versions: []dependencies.MinimalPackageRepoRefVersion{{Version: "foobar"}}, // test that we create a repo for this package even if it's missing.
		},
	})

	svc := typestest.MakeExternalService(t,
		extsvc.VariantPythonPackages,
		&schema.PythonPackagesConnection{
			Urls: []string{
				"https://pypi.org/simple",
			},
			Dependencies: []string{
				"requests==2.27.1",
				"lavaclient==0.3.7",
				"randio==0.1.1",
				"pytimeparse==1.1.8",
			},
		})

	cf, save := NewClientFactory(t, t.Name())
	t.Cleanup(func() { save(t) })

	src, err := NewPythonPackagesSource(ctx, svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	src.SetDependenciesService(depsSvc)

	repos, err := ListAll(ctx, src)
	if err != nil {
		t.Fatal(err)
	}

	sort.SliceStable(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	testutil.AssertGolden(t, "testdata/sources/"+t.Name(), Update(t.Name()), repos)
}
