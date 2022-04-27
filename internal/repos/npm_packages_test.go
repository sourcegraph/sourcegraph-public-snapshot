package repos

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/live"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmpackages"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
			Name:   testCase.pkgName,
		})
		require.Nil(t, err)
		depStrs := []string{}
		for _, dep := range deps {
			pkg, err := reposource.ParseNpmPackageFromPackageSyntax(dep.Name)
			require.Nil(t, err)
			depStrs = append(depStrs,
				(&reposource.NpmDependency{NpmPackage: pkg, Version: dep.Version}).PackageManagerSyntax(),
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
				Name:   testCase.pkgName,
				After:  lastID,
				Limit:  1,
			})
			require.Nil(t, err)
			require.Equal(t, len(deps), 1)
			pkg, err := reposource.ParseNpmPackageFromPackageSyntax(deps[0].Name)
			require.Nil(t, err)
			depStrs = append(depStrs, (&reposource.NpmDependency{NpmPackage: pkg, Version: deps[0].Version}).PackageManagerSyntax())
			lastID = deps[0].ID
		}
		sort.Strings(depStrs)
		sort.Strings(testCase.matches)
		require.Equal(t, testCase.matches, depStrs)
	}
}

func testDependenciesService(ctx context.Context, t *testing.T, dependencyRepos []dependencies.Repo) *dependencies.Service {
	t.Helper()
	db := database.NewDB(dbtest.NewDB(t))
	depsSvc := live.TestService(db, nil)

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
		dep, err := reposource.ParseNpmDependency(depStr)
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

func TestListRepos(t *testing.T) {
	ctx := context.Background()
	depsSvc := testDependenciesService(ctx, t, testDependencyRepos)

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	svc := types.ExternalService{
		Kind:   extsvc.KindNpmPackages,
		Config: `{"registry": "https://placeholder.lol", "rateLimit": {"enabled": false}}`,
	}
	packageSource, err := NewNpmPackagesSource(&svc)
	require.Nil(t, err)
	packageSource.SetDependenciesService(depsSvc)
	packageSource.client = npmtest.NewMockClient(t, testDependencies...)
	results := make(chan SourceResult, 10)
	go func() {
		packageSource.ListRepos(ctx, results)
		close(results)
	}()

	var have []*types.Repo
	for r := range results {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		have = append(have, r.Repo)
	}

	sort.Sort(types.Repos(have))

	var want []*types.Repo
	for _, dep := range testDependencies {
		dep, err := reposource.ParseNpmDependency(dep)
		if err != nil {
			t.Fatal(err)
		}
		want = append(want, &types.Repo{
			Name:        dep.RepoName(),
			Description: dep.PackageSyntax() + " description",
			URI:         string(dep.RepoName()),
			ExternalRepo: api.ExternalRepoSpec{
				ID:          string(dep.RepoName()),
				ServiceID:   extsvc.TypeNpmPackages,
				ServiceType: extsvc.TypeNpmPackages,
			},
			Sources: map[string]*types.SourceInfo{
				packageSource.svc.URN(): {
					ID:       packageSource.svc.URN(),
					CloneURL: dep.CloneURL(),
				},
			},
			Metadata: &npmpackages.Metadata{
				Package: dep.NpmPackage,
			},
		})
	}

	sort.Sort(types.Repos(want))

	// Compare after uniquing after addressing [FIXME: deduplicate-listed-repos].
	require.Equal(t, want, have)
}
