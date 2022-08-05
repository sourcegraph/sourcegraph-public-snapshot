package repos

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPythonPackagesSource_ListRepos(t *testing.T) {
	ctx := context.Background()
	depsSvc := testDependenciesService(ctx, t, []dependencies.Repo{
		{
			ID:      1,
			Scheme:  dependencies.PythonPackagesScheme,
			Name:    "requests",
			Version: "2.27.1", // test deduplication with version from config
		},
		{
			ID:      2,
			Scheme:  dependencies.PythonPackagesScheme,
			Name:    "requests",
			Version: "2.27.2", // test multiple versions of the same module
		},
		{
			ID:      3,
			Scheme:  dependencies.PythonPackagesScheme,
			Name:    "numpy",
			Version: "1.22.3",
		},
		{
			ID:      4,
			Scheme:  dependencies.PythonPackagesScheme,
			Name:    "lofi",
			Version: "foobar", // test that we create a repo for this package even if it's missing.
		}})

	svc := types.ExternalService{
		Kind: extsvc.KindPythonPackages,
		Config: marshalJSON(t, &schema.PythonPackagesConnection{
			Urls: []string{
				"https://pypi.org/simple",
			},
			Dependencies: []string{
				"requests==2.27.1",
				"lavaclient==0.3.7",
				"randio==0.1.1",
				"pytimeparse==1.1.8",
			},
		}),
	}

	cf, save := newClientFactory(t, t.Name())
	t.Cleanup(func() { save(t) })

	src, err := NewPythonPackagesSource(&svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	src.SetDependenciesService(depsSvc)

	repos, err := listAll(ctx, src)
	if err != nil {
		t.Fatal(err)
	}

	sort.SliceStable(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	testutil.AssertGolden(t, "testdata/sources/"+t.Name(), update(t.Name()), repos)
}
