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
	"github.com/sourcegraph/sourcegraph/internal/types"
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
		deps, err := depsSvc.ListDependencyRepos(ctx, dependencies.ListDependencyReposOpts{
			Scheme: dependencies.NpmPackagesScheme,
			Name:   reposource.PackageName(testCase.pkgName),
		})
		require.Nil(t, err)
		depStrs := []string{}
		for _, dep := range deps {
			pkg, err := reposource.ParseNpmPackageFromPackageSyntax(dep.Name)
			require.Nil(t, err)
			depStrs = append(depStrs,
				(&reposource.NpmVersionedPackage{NpmPackageName: pkg, Version: dep.Version}).VersionedPackageSyntax(),
			)
		}
		sort.Strings(depStrs)
		sort.Strings(testCase.matches)
		require.Equal(t, testCase.matches, depStrs)
	}

	for _, testCase := range testCases {
		depStrs := []string{}
		lastID := 0
		for i := 0; i < len(testCase.matches); i++ {
			deps, err := depsSvc.ListDependencyRepos(ctx, dependencies.ListDependencyReposOpts{
				Scheme: dependencies.NpmPackagesScheme,
				Name:   reposource.PackageName(testCase.pkgName),
				After:  lastID,
				Limit:  1,
			})
			require.Nil(t, err)
			require.Equal(t, len(deps), 1)
			pkg, err := reposource.ParseNpmPackageFromPackageSyntax(deps[0].Name)
			require.Nil(t, err)
			depStrs = append(depStrs, (&reposource.NpmVersionedPackage{NpmPackageName: pkg, Version: deps[0].Version}).VersionedPackageSyntax())
			lastID = deps[0].ID
		}
		sort.Strings(depStrs)
		sort.Strings(testCase.matches)
		require.Equal(t, testCase.matches, depStrs)
	}
}

func testDependenciesService(ctx context.Context, t *testing.T, dependencyRepos []dependencies.Repo) *dependencies.Service {
	t.Helper()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	depsSvc := dependencies.TestService(db, nil)

	_, err := depsSvc.UpsertDependencyRepos(ctx, dependencyRepos)
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
var testDependencyRepos = func() []dependencies.Repo {
	dependencyRepos := []dependencies.Repo{}
	for i, depStr := range testDependencies {
		dep, err := reposource.ParseNpmVersionedPackage(depStr)
		if err != nil {
			panic(err.Error())
		}

		dependencyRepos = append(dependencyRepos, dependencies.Repo{
			ID:      i + 1,
			Scheme:  dependencies.NpmPackagesScheme,
			Name:    dep.PackageSyntax(),
			Version: dep.Version,
		})
	}

	return dependencyRepos
}()

func TestNPMPackagesSource_ListRepos(t *testing.T) {
	ctx := context.Background()
	depsSvc := testDependenciesService(ctx, t, []dependencies.Repo{
		{
			ID:      1,
			Scheme:  dependencies.NpmPackagesScheme,
			Name:    "@sourcegraph/sourcegraph.proposed",
			Version: "12.0.0", // test deduplication with version from config
		},
		{
			ID:      2,
			Scheme:  dependencies.NpmPackagesScheme,
			Name:    "@sourcegraph/sourcegraph.proposed",
			Version: "12.0.1", // test deduplication with version from config
		},
		{
			ID:      3,
			Scheme:  dependencies.NpmPackagesScheme,
			Name:    "@sourcegraph/web-ext",
			Version: "3.0.0-fork.1",
		},
		{
			ID:      4,
			Scheme:  dependencies.NpmPackagesScheme,
			Name:    "fastq",
			Version: "0.9.9", // test missing modules still create a repo.
		},
	})

	svc := types.ExternalService{
		Kind: extsvc.KindNpmPackages,
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.NpmPackagesConnection{
			Registry:     "https://registry.npmjs.org",
			Dependencies: []string{"@sourcegraph/prettierrc@2.2.0"},
		})),
	}

	cf, save := newClientFactory(t, t.Name())
	t.Cleanup(func() { save(t) })

	src, err := NewNpmPackagesSource(ctx, &svc, cf)
	if err != nil {
		t.Fatal(err)
	}

	src.SetDependenciesService(depsSvc)

	repos, err := listAll(ctx, src)
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/sources/"+t.Name(), update(t.Name()), repos)
}
