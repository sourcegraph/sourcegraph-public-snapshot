package repos

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetNpmDependencyRepos(t *testing.T) {
	ctx := context.Background()
	depsSvc := testDependenciesService(ctx, t, testDependencyRepos)

	type testCase struct {
		pkgName string
		matches []string
	}

	testCases := []testCase{
		{"pkg1", []string{"pkg1@1", "pkg1@2", "pkg1@3"}},
		{"pkg2", []string{"pkg2@1", "pkg2@0.1-abc"}},
		{"@scope/pkg1", []string{"@scope/pkg1@1"}},
		{"missing", []string{}},
	}

	for _, testCase := range testCases {
		deps, _, hasMore, err := depsSvc.ListPackageRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme:        dependencies.NpmPackagesScheme,
			Name:          reposource.PackageName(testCase.pkgName),
			ExactNameOnly: true,
		})
		if err != nil {
			t.Fatalf("unexpected error listing package repos: %v", err)
		}

		if hasMore {
			t.Error("unexpected more-pages flag set, expected no more pages to follow")
		}

		depStrs := []string{}
		for _, dep := range deps {
			pkg, err := reposource.ParseNpmPackageFromPackageSyntax(dep.Name)
			if err != nil {
				t.Fatalf("unexpected error parsing package from package name: %v", err)
			}

			for _, version := range dep.Versions {
				depStrs = append(depStrs,
					(&reposource.NpmVersionedPackage{
						NpmPackageName: pkg,
						Version:        version.Version,
					}).VersionedPackageSyntax(),
				)
			}
		}
		sort.Strings(depStrs)
		sort.Strings(testCase.matches)
		require.Equal(t, testCase.matches, depStrs)
	}

	for _, testCase := range testCases {
		var depStrs []string
		deps, _, _, err := depsSvc.ListPackageRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme:        dependencies.NpmPackagesScheme,
			Name:          reposource.PackageName(testCase.pkgName),
			ExactNameOnly: true,
			Limit:         1,
		})
		require.Nil(t, err)
		if len(testCase.matches) > 0 {
			require.Equal(t, 1, len(deps))
		} else {
			require.Equal(t, 0, len(deps))
			continue
		}
		pkg, err := reposource.ParseNpmPackageFromPackageSyntax(deps[0].Name)
		require.Nil(t, err)
		for _, version := range deps[0].Versions {
			depStrs = append(depStrs, (&reposource.NpmVersionedPackage{
				NpmPackageName: pkg,
				Version:        version.Version,
			}).VersionedPackageSyntax())
		}
		sort.Strings(depStrs)
		sort.Strings(testCase.matches)
		require.Equal(t, testCase.matches, depStrs)
	}
}

func testDependenciesService(ctx context.Context, t *testing.T, dependencyRepos []dependencies.MinimalPackageRepoRef) *dependencies.Service {
	t.Helper()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	depsSvc := dependencies.TestService(db)

	_, _, err := depsSvc.InsertPackageRepoRefs(ctx, dependencyRepos)
	if err != nil {
		t.Fatalf(err.Error())
	}

	return depsSvc
}

var testDependencies = []string{
	"@scope/pkg1@1",
	"pkg1@1",
	"pkg1@2",
	"pkg1@3",
	"pkg2@0.1-abc",
	"pkg2@1",
}

var testDependencyRepos = func() []dependencies.MinimalPackageRepoRef {
	dependencyRepos := []dependencies.MinimalPackageRepoRef{}
	for _, depStr := range testDependencies {
		dep, err := reposource.ParseNpmVersionedPackage(depStr)
		if err != nil {
			panic(err.Error())
		}

		dependencyRepos = append(dependencyRepos, dependencies.MinimalPackageRepoRef{
			Scheme:   dependencies.NpmPackagesScheme,
			Name:     dep.PackageSyntax(),
			Versions: []dependencies.MinimalPackageRepoRefVersion{{Version: dep.Version}},
		})
	}

	return dependencyRepos
}()

func TestNPMPackagesSource_ListRepos(t *testing.T) {
	ctx := context.Background()
	depsSvc := testDependenciesService(ctx, t, []dependencies.MinimalPackageRepoRef{
		{
			Scheme: dependencies.NpmPackagesScheme,
			Name:   "@sourcegraph/sourcegraph.proposed",
			Versions: []dependencies.MinimalPackageRepoRefVersion{
				{Version: "12.0.0"}, // test deduplication with version from config
				{Version: "12.0.1"}, // test deduplication with version from config
			},
		},
		{
			Scheme:   dependencies.NpmPackagesScheme,
			Name:     "@sourcegraph/web-ext",
			Versions: []dependencies.MinimalPackageRepoRefVersion{{Version: "3.0.0-fork.1"}},
		},
		{
			Scheme:   dependencies.NpmPackagesScheme,
			Name:     "fastq",
			Versions: []dependencies.MinimalPackageRepoRefVersion{{Version: "0.9.9"}}, // test missing modules still create a repo.
		},
	})

	svc := typestest.MakeExternalService(t, extsvc.VariantNpmPackages, &schema.NpmPackagesConnection{
		Registry:     "https://registry.npmjs.org",
		Dependencies: []string{"@sourcegraph/prettierrc@2.2.0"},
	})

	cf, save := NewClientFactory(t, t.Name())
	t.Cleanup(func() { save(t) })

	src, err := NewNpmPackagesSource(ctx, svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	src.SetDependenciesService(depsSvc)

	repos, err := ListAll(ctx, src)
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/sources/"+t.Name(), Update(t.Name()), repos)
}
