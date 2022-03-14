package repos

import (
	"context"
	"database/sql"
	"os"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmpackages"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	dependenciesStore "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/store"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetNpmDependencyRepos(t *testing.T) {
	_, store, ctx, _ := setupDependenciesInDB(t)

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
		deps, err := store.ListDependencyRepos(ctx, dependenciesStore.ListDependencyReposOpts{
			Scheme: dependenciesStore.NpmPackagesScheme,
			Name:   testCase.pkgName,
		})
		require.Nil(t, err)
		depStrs := []string{}
		for _, dep := range deps {
			pkg, err := reposource.ParseNpmPackageFromPackageSyntax(dep.Name)
			require.Nil(t, err)
			depStrs = append(depStrs,
				(&reposource.NpmDependency{pkg, dep.Version}).PackageManagerSyntax(),
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
			deps, err := store.ListDependencyRepos(ctx, dependenciesStore.ListDependencyReposOpts{
				Scheme: dependenciesStore.NpmPackagesScheme,
				Name:   testCase.pkgName,
				After:  lastID,
				Limit:  1,
			})
			require.Nil(t, err)
			require.Equal(t, len(deps), 1)
			pkg, err := reposource.ParseNpmPackageFromPackageSyntax(deps[0].Name)
			require.Nil(t, err)
			depStrs = append(depStrs, (&reposource.NpmDependency{pkg, deps[0].Version}).PackageManagerSyntax())
			lastID = deps[0].ID
		}
		sort.Strings(depStrs)
		sort.Strings(testCase.matches)
		require.Equal(t, testCase.matches, depStrs)
	}
}

func setupDependenciesInDB(t *testing.T) (*sql.DB, *dependenciesStore.Store, context.Context, []string) {
	t.Helper()
	db := dbtest.NewDB(t)
	store := dependenciesStore.TestStore(database.NewDB(db))
	ctx := context.Background()

	dependencies := []string{
		"pkg1@1",
		"pkg1@2",
		"pkg2@1",
		"@scope/pkg1@1",
		"pkg1@3",
		"pkg2@0.1-abc",
	}
	insertDependencies(t, ctx, store, dependencies)
	return db, store, ctx, dependencies
}

func TestListRepos(t *testing.T) {
	db, _, ctx, dependencies := setupDependenciesInDB(t)
	sort.Strings(dependencies)

	dir, err := os.MkdirTemp("", "")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	svc := types.ExternalService{
		Kind:   extsvc.KindNpmPackages,
		Config: `{"registry": "https://placeholder.lol", "rateLimit": {"enabled": false}}`,
	}
	packageSource, err := NewNpmPackagesSource(&svc)
	require.Nil(t, err)
	packageSource.SetDB(db)
	packageSource.client = npmtest.NewMockClient(t, dependencies...)
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
	for _, dep := range dependencies {
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

func insertDependencies(t *testing.T, ctx context.Context, s *dependenciesStore.Store, dependencies []string) {
	for _, depStr := range dependencies {
		dep, err := reposource.ParseNpmDependency(depStr)
		require.Nil(t, err)
		// See also: enterprise/internal/codeintel/stores/dbstore/dependency_index.go:InsertCloneableDependencyRepo
		rows, err :=
			s.Store.Query(ctx, sqlf.Sprintf(
				`INSERT INTO lsif_dependency_repos (scheme, name, version) VALUES (%s, %s, %s)`,
				dependenciesStore.NpmPackagesScheme, dep.PackageSyntax(), dep.Version))
		require.Nil(t, err)
		for rows.Next() {
		}
		require.Nil(t, rows.Err())
	}
}
