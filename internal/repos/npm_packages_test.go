package repos

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmpackages"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetNpmDependencyRepos(t *testing.T) {
	ctx, depsSvc := setupDependenciesService(t, parseDepStrs(t, []string{
		"pkg1@1",
		"pkg1@2",
		"pkg2@1",
		"@scope/pkg1@1",
		"pkg1@3",
		"pkg2@0.1-abc",
	}))

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
			depStrs = append(depStrs, (&reposource.NpmDependency{pkg, deps[0].Version}).PackageManagerSyntax())
			lastID = deps[0].ID
		}
		sort.Strings(depStrs)
		sort.Strings(testCase.matches)
		require.Equal(t, testCase.matches, depStrs)
	}
}

func setupDependenciesService(t *testing.T, drs []dependencies.DependencyRepo) (context.Context, DependenciesService) {
	t.Helper()
	ctx := context.Background()
	depsSvc := NewMockDependenciesService()

	depsSvc.ListDependencyReposFunc.SetDefaultHook(func(ctx context.Context, opts dependencies.ListDependencyReposOpts) (matching []dependencies.DependencyRepo, _ error) {
		sort.Slice(drs, func(i, j int) bool {
			if opts.NewestFirst {
				return drs[i].ID > drs[j].ID
			} else {
				return drs[i].ID < drs[j].ID
			}
		})

		for _, dependencyRepo := range drs {
			matches := dependencyRepo.Scheme == opts.Scheme && (opts.Name == "" || dependencyRepo.Name == opts.Name)
			if !matches {
				continue
			}

			inPage := opts.After == 0 || ((opts.NewestFirst && opts.After > dependencyRepo.ID) || (!opts.NewestFirst && opts.After < dependencyRepo.ID))
			if !inPage {
				continue
			}

			matching = append(matching, dependencyRepo)
		}

		if opts.Limit != 0 && len(matching) > opts.Limit {
			matching = matching[:opts.Limit]
		}

		return matching, nil
	})

	return ctx, depsSvc
}

func parseDepStrs(t *testing.T, depStrs []string) []dependencies.DependencyRepo {
	parsedDependencyRepos := []dependencies.DependencyRepo{}
	for i, depStr := range depStrs {
		dep, err := reposource.ParseNpmDependency(depStr)
		if err != nil {
			t.Fatal(err)
		}

		parsedDependencyRepos = append(parsedDependencyRepos, dependencies.DependencyRepo{
			ID:      i + 1,
			Scheme:  dependencies.NpmPackagesScheme,
			Name:    dep.PackageSyntax(),
			Version: dep.Version,
		})
	}

	return parsedDependencyRepos
}

func TestListRepos(t *testing.T) {
	dependencies := []string{
		"pkg1@1",
		"pkg1@2",
		"pkg2@1",
		"@scope/pkg1@1",
		"pkg1@3",
		"pkg2@0.1-abc",
	}
	sort.Strings(dependencies)
	ctx, depsSvc := setupDependenciesService(t, parseDepStrs(t, dependencies))

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
