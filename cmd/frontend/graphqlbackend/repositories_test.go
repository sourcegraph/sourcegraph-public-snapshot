package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestRepositoriesCloneStatusFiltering(t *testing.T) {
	mockRepos := []*types.Repo{
		{Name: "repo1"}, // not_cloned
		{Name: "repo2"}, // cloning
		{Name: "repo3"}, // cloned
	}

	repos := database.NewMockRepoStore()
	repos.ListFunc.SetDefaultHook(func(ctx context.Context, opt database.ReposListOptions) ([]*types.Repo, error) {
		if opt.NoCloned {
			return mockRepos[0:2], nil
		}
		if opt.OnlyCloned {
			return mockRepos[2:], nil
		}

		if opt.CloneStatus == types.CloneStatusNotCloned {
			return mockRepos[:1], nil
		}
		if opt.CloneStatus == types.CloneStatusCloning {
			return mockRepos[1:2], nil
		}
		if opt.CloneStatus == types.CloneStatusCloned {
			return mockRepos[2:], nil
		}

		return mockRepos, nil
	})
	repos.CountFunc.SetDefaultReturn(len(mockRepos), nil)

	users := database.NewMockUserStore()

	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)
	db.UsersFunc.SetDefaultReturn(users)

	schema := mustParseGraphQLSchema(t, db)

	t.Run("not as a site admin", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)

		RunTests(t, []*Test{
			{
				Schema: schema,
				Query: `
				{
					repositories {
						nodes { name }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo1" },
							{ "name": "repo2" },
							{ "name": "repo3" }
						],
						"totalCount": null,
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Path:          []any{"repositories", "totalCount"},
						Message:       auth.ErrMustBeSiteAdmin.Error(),
						ResolverError: auth.ErrMustBeSiteAdmin,
					},
				},
			},
		})
	})

	t.Run("as a site admin", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		RunTests(t, []*Test{
			{
				Schema: schema,
				Query: `
				{
					repositories {
						nodes { name }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo1" },
							{ "name": "repo2" },
							{ "name": "repo3" }
						],
						"totalCount": 3,
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
			},
			{
				Schema: schema,
				// cloned and notCloned are true by default
				// this test ensures the behavior is the same
				// when setting them explicitly
				Query: `
				{
					repositories(cloned: true, notCloned: true) {
						nodes { name }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo1" },
							{ "name": "repo2" },
							{ "name": "repo3" }
						],
						"totalCount": 3,
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
			},
			{
				Schema: schema,
				Query: `
				{
					repositories(first: 2) {
						nodes { name }
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo1" },
							{ "name": "repo2" }
						],
						"pageInfo": {"hasNextPage": true}
					}
				}
			`,
			},
			{
				Schema: schema,
				Query: `
				{
					repositories(cloned: false) {
						nodes { name }
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo1" },
							{ "name": "repo2" }
						],
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
			},
			{
				Schema: schema,
				Query: `
				{
					repositories(notCloned: false) {
						nodes { name }
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo3" }
						],
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
			},
			{
				Schema: schema,
				Query: `
				{
					repositories(notCloned: false, cloned: false) {
						nodes { name }
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: "null",
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Path:          []any{"repositories"},
						Message:       "excluding cloned and not cloned repos leaves an empty set",
						ResolverError: errors.New("excluding cloned and not cloned repos leaves an empty set"),
					},
				},
			},
			{
				Schema: schema,
				Query: `
				{
					repositories(cloneStatus: CLONED) {
						nodes { name }
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo3" }
						],
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
			},
			{
				Schema: schema,
				Query: `
				{
					repositories(cloneStatus: CLONING) {
						nodes { name }
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo2" }
						],
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
			},
			{
				Schema: schema,
				Query: `
				{
					repositories(cloneStatus: NOT_CLONED) {
						nodes { name }
						pageInfo { hasNextPage }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo1" }
						],
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
			},
		})
	})
}

func TestRepositoriesIndexingFiltering(t *testing.T) {
	mockRepos := map[string]bool{
		"repo-indexed-1":     true,
		"repo-indexed-2":     true,
		"repo-not-indexed-3": false,
		"repo-not-indexed-4": false,
	}

	filterRepos := func(t *testing.T, opt database.ReposListOptions) []*types.Repo {
		t.Helper()
		var repos types.Repos
		for n, idx := range mockRepos {
			if opt.NoIndexed && idx {
				continue
			}
			if opt.OnlyIndexed && !idx {
				continue
			}
			repos = append(repos, &types.Repo{Name: api.RepoName(n)})
		}
		sort.Sort(repos)
		return repos
	}
	repos := database.NewMockRepoStore()
	repos.ListFunc.SetDefaultHook(func(_ context.Context, opt database.ReposListOptions) ([]*types.Repo, error) {
		return filterRepos(t, opt), nil
	})
	repos.CountFunc.SetDefaultHook(func(_ context.Context, opt database.ReposListOptions) (int, error) {
		repos := filterRepos(t, opt)
		return len(repos), nil
	})

	users := database.NewMockUserStore()

	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)
	db.UsersFunc.SetDefaultReturn(users)

	schema := mustParseGraphQLSchema(t, db)

	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

	RunTests(t, []*Test{
		{
			Schema: schema,
			Query: `
				{
					repositories {
						nodes { name }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo-indexed-1" },
							{ "name": "repo-indexed-2" },
							{ "name": "repo-not-indexed-3" },
							{ "name": "repo-not-indexed-4" }
						],
						"totalCount": 4,
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
		},
		{
			Schema: schema,
			// indexed and notIndexed are true by default
			// this test ensures the behavior is the same
			// when setting them explicitly
			Query: `
				{
					repositories(indexed: true, notIndexed: true) {
						nodes { name }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo-indexed-1" },
							{ "name": "repo-indexed-2" },
							{ "name": "repo-not-indexed-3" },
							{ "name": "repo-not-indexed-4" }
						],
						"totalCount": 4,
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
		},
		{
			Schema: schema,
			Query: `
				{
					repositories(indexed: false) {
						nodes { name }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo-not-indexed-3" },
							{ "name": "repo-not-indexed-4" }
						],
						"totalCount": 2,
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
		},
		{
			Schema: schema,
			Query: `
				{
					repositories(notIndexed: false) {
						nodes { name }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "name": "repo-indexed-1" },
							{ "name": "repo-indexed-2" }
						],
						"totalCount": 2,
						"pageInfo": {"hasNextPage": false}
					}
				}
			`,
		},
		{
			Schema: schema,
			Query: `
				{
					repositories(notIndexed: false, indexed: false) {
						nodes { name }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []any{"repositories"},
					Message:       "excluding indexed and not indexed repos leaves an empty set",
					ResolverError: errors.New("excluding indexed and not indexed repos leaves an empty set"),
				},
			},
		},
	})
}

func TestRepositories_CursorPagination(t *testing.T) {
	mockRepos := []*types.Repo{
		{ID: 0, Name: "repo1"},
		{ID: 1, Name: "repo2"},
		{ID: 2, Name: "repo3"},
	}

	repos := database.NewMockRepoStore()
	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)

	buildQuery := func(first int, after string) string {
		var args []string
		if first != 0 {
			args = append(args, fmt.Sprintf("first: %d", first))
		}
		if after != "" {
			args = append(args, fmt.Sprintf("after: %q", after))
		}

		return fmt.Sprintf(`{ repositories(%s) { nodes { name } pageInfo { endCursor } } }`, strings.Join(args, ", "))
	}

	t.Run("Initial page without a cursor present", func(t *testing.T) {
		repos.ListFunc.SetDefaultReturn(mockRepos[0:2], nil)

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  buildQuery(1, ""),
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [{
							"name": "repo1"
						}],
						"pageInfo": {
						  "endCursor": "UmVwb3NpdG9yeUN1cnNvcjp7IkNvbHVtbiI6Im5hbWUiLCJWYWx1ZSI6InJlcG8yIiwiRGlyZWN0aW9uIjoibmV4dCJ9"
						}
					}
				}
			`,
		})
	})

	t.Run("Second page in ascending order", func(t *testing.T) {
		repos.ListFunc.SetDefaultReturn(mockRepos[1:], nil)

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  buildQuery(1, "UmVwb3NpdG9yeUN1cnNvcjp7IkNvbHVtbiI6Im5hbWUiLCJWYWx1ZSI6InJlcG8yIiwiRGlyZWN0aW9uIjoibmV4dCJ9"),
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [{
							"name": "repo2"
						}],
						"pageInfo": {
						  "endCursor": "UmVwb3NpdG9yeUN1cnNvcjp7IkNvbHVtbiI6Im5hbWUiLCJWYWx1ZSI6InJlcG8zIiwiRGlyZWN0aW9uIjoibmV4dCJ9"
						}
					}
				}
			`,
		})
	})

	t.Run("Second page in descending order", func(t *testing.T) {
		repos.ListFunc.SetDefaultReturn(mockRepos[1:], nil)

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  buildQuery(1, "UmVwb3NpdG9yeUN1cnNvcjp7IkNvbHVtbiI6Im5hbWUiLCJWYWx1ZSI6InJlcG8yIiwiRGlyZWN0aW9uIjoicHJldiJ9"),
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [{
							"name": "repo2"
						}],
						"pageInfo": {
						  "endCursor": "UmVwb3NpdG9yeUN1cnNvcjp7IkNvbHVtbiI6Im5hbWUiLCJWYWx1ZSI6InJlcG8zIiwiRGlyZWN0aW9uIjoicHJldiJ9"
						}
					}
				}
			`,
		})
	})

	t.Run("Initial page with no further rows to fetch", func(t *testing.T) {
		repos.ListFunc.SetDefaultReturn(mockRepos, nil)

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  buildQuery(3, ""),
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [{
							"name": "repo1"
						}, {
							"name": "repo2"
						}, {
							"name": "repo3"
						}],
						"pageInfo": {
						  "endCursor": null
						}
					}
				}
			`,
		})
	})

	t.Run("With no repositories present", func(t *testing.T) {
		repos.ListFunc.SetDefaultReturn(nil, nil)

		RunTest(t, &Test{
			Schema: mustParseGraphQLSchema(t, db),
			Query:  buildQuery(1, ""),
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [],
						"pageInfo": {
						  "endCursor": null
						}
					}
				}
			`,
		})
	})

	t.Run("With an invalid cursor provided", func(t *testing.T) {
		repos.ListFunc.SetDefaultReturn(nil, nil)

		RunTest(t, &Test{
			Schema:         mustParseGraphQLSchema(t, db),
			Query:          buildQuery(1, "invalid-cursor-value"),
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Path:          []any{"repositories"},
					Message:       `cannot unmarshal repository cursor type: ""`,
					ResolverError: errors.New(`cannot unmarshal repository cursor type: ""`),
				},
			},
		})
	})
}

func TestRepositories_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	schema := mustParseGraphQLSchema(t, db)

	repos := []struct {
		repo        *types.Repo
		size        int64
		cloneStatus types.CloneStatus
		indexed     bool
		lastError   string
	}{
		{repo: &types.Repo{Name: "repo1"}, size: 20, cloneStatus: types.CloneStatusNotCloned},
		{repo: &types.Repo{Name: "repo2"}, size: 30, cloneStatus: types.CloneStatusNotCloned, lastError: "repo2 error"},
		{repo: &types.Repo{Name: "repo3"}, size: 40, cloneStatus: types.CloneStatusCloning},
		{repo: &types.Repo{Name: "repo4"}, size: 50, cloneStatus: types.CloneStatusCloning, lastError: "repo4 error"},
		{repo: &types.Repo{Name: "repo5"}, size: 60, cloneStatus: types.CloneStatusCloned},
		{repo: &types.Repo{Name: "repo6"}, size: 10, cloneStatus: types.CloneStatusCloned, lastError: "repo6 error"},
		{repo: &types.Repo{Name: "repo7"}, size: 70, cloneStatus: types.CloneStatusCloned, indexed: false},
		{repo: &types.Repo{Name: "repo8"}, size: 80, cloneStatus: types.CloneStatusCloned, indexed: true},
	}

	for _, rsc := range repos {
		if err := db.Repos().Create(ctx, rsc.repo); err != nil {
			t.Fatal(err)
		}

		gitserverRepos := db.GitserverRepos()
		if err := gitserverRepos.SetRepoSize(ctx, rsc.repo.Name, rsc.size, "shard-1"); err != nil {
			t.Fatal(err)
		}
		if err := gitserverRepos.SetCloneStatus(ctx, rsc.repo.Name, rsc.cloneStatus, "shard-1"); err != nil {
			t.Fatal(err)
		}

		if rsc.indexed {
			err := db.ZoektRepos().UpdateIndexStatuses(ctx, map[uint32]*zoekt.MinimalRepoListEntry{
				uint32(rsc.repo.ID): {},
			})
			if err != nil {
				t.Fatal(err)
			}
		}

		if msg := rsc.lastError; msg != "" {
			if err := gitserverRepos.SetLastError(ctx, rsc.repo.Name, msg, "shard-1"); err != nil {
				t.Fatal(err)
			}
		}
	}

	admin, err := db.Users().Create(ctx, database.NewUser{Username: "admin", Password: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, actor.FromUser(admin.ID))

	tests := []repositoriesQueryTest{
		// no args
		{
			wantRepos:      []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8"},
			wantTotalCount: 8,
		},
		// first
		{
			args:           "first: 2",
			wantRepos:      []string{"repo1", "repo2"},
			wantTotalCount: 8,
		},
		// cloned
		{
			// cloned only says whether to "Include cloned repositories.", it doesn't exclude non-cloned.
			args:           "cloned: true",
			wantRepos:      []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8"},
			wantTotalCount: 8,
		},
		{
			args:           "cloned: false",
			wantRepos:      []string{"repo1", "repo2", "repo3", "repo4"},
			wantTotalCount: 4,
		},
		{
			args:           "cloned: false, first: 2",
			wantRepos:      []string{"repo1", "repo2"},
			wantTotalCount: 4,
		},
		// notCloned
		{
			args:           "notCloned: true",
			wantRepos:      []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8"},
			wantTotalCount: 8,
		},
		{
			args:           "notCloned: false",
			wantRepos:      []string{"repo5", "repo6", "repo7", "repo8"},
			wantTotalCount: 4,
		},
		// failedFetch
		{
			args:           "failedFetch: true",
			wantRepos:      []string{"repo2", "repo4", "repo6"},
			wantTotalCount: 3,
		},
		{
			args:           "failedFetch: true, first: 2",
			wantRepos:      []string{"repo2", "repo4"},
			wantTotalCount: 3,
		},
		{
			args:           "failedFetch: false",
			wantRepos:      []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8"},
			wantTotalCount: 8,
		},
		// cloneStatus
		{
			args:           "cloneStatus:NOT_CLONED",
			wantRepos:      []string{"repo1", "repo2"},
			wantTotalCount: 2,
		},
		{
			args:           "cloneStatus:CLONING",
			wantRepos:      []string{"repo3", "repo4"},
			wantTotalCount: 2,
		},
		{
			args:           "cloneStatus:CLONED",
			wantRepos:      []string{"repo5", "repo6", "repo7", "repo8"},
			wantTotalCount: 4,
		},
		{
			args:           "cloneStatus:NOT_CLONED, first: 1",
			wantRepos:      []string{"repo1"},
			wantTotalCount: 2,
		},
		// indexed
		{
			// indexed only says whether to "Include indexed repositories.", it doesn't exclude non-indexed.
			args:           "indexed: true",
			wantRepos:      []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8"},
			wantTotalCount: 8,
		},
		{
			args:           "indexed: false",
			wantRepos:      []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7"},
			wantTotalCount: 7,
		},
		{
			args:           "indexed: false, first: 2",
			wantRepos:      []string{"repo1", "repo2"},
			wantTotalCount: 7,
		},
		// notIndexed
		{
			args:           "notIndexed: true",
			wantRepos:      []string{"repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7", "repo8"},
			wantTotalCount: 8,
		},
		{
			args:           "notIndexed: false",
			wantRepos:      []string{"repo8"},
			wantTotalCount: 1,
		},
		{
			args:           "orderBy:SIZE, descending:false, first: 5",
			wantRepos:      []string{"repo6", "repo1", "repo2", "repo3", "repo4"},
			wantTotalCount: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			runRepositoriesQuery(t, ctx, schema, tt)
		})
	}

}

type repositoriesQueryTest struct {
	args string

	wantRepos []string

	wantNoTotalCount bool
	wantTotalCount   int
}

func runRepositoriesQuery(t *testing.T, ctx context.Context, schema *graphql.Schema, want repositoriesQueryTest) {
	t.Helper()

	type node struct {
		Name string `json:"name"`
	}

	type repositories struct {
		Nodes      []node `json:"nodes"`
		TotalCount *int   `json:"totalCount"`
	}

	type expected struct {
		Repositories repositories `json:"repositories"`
	}

	nodes := make([]node, 0, len(want.wantRepos))
	for _, name := range want.wantRepos {
		nodes = append(nodes, node{Name: name})
	}

	ex := expected{
		Repositories: repositories{
			Nodes:      nodes,
			TotalCount: &want.wantTotalCount,
		},
	}

	if want.wantNoTotalCount {
		ex.Repositories.TotalCount = nil
	}

	marshaled, err := json.Marshal(ex)
	if err != nil {
		t.Fatalf("failed to marshal expected repositories query result: %s", err)
	}

	var query string
	if want.args != "" {
		query = fmt.Sprintf(`{ repositories(%s) { nodes { name } totalCount } } `, want.args)
	} else {
		query = `{ repositories { nodes { name } totalCount } }`
	}

	RunTest(t, &Test{
		Context:        ctx,
		Schema:         schema,
		Query:          query,
		ExpectedResult: string(marshaled),
	})
}
