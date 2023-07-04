package gitserver

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/require"
)

func TestAddrForRepo(t *testing.T) {
	db := database.NewMockDB()
	ga := GitserverAddresses{
		db:        db,
		Addresses: []string{"gitserver-1", "gitserver-2", "gitserver-3"},
		PinnedServers: map[string]string{
			"repo2": "gitserver-1",
		},
	}

	t.Run("no deduplicated forks", func(t *testing.T) {
		testCases := []struct {
			name string
			repo api.RepoName
			want string
		}{
			{
				name: "repo1",
				repo: api.RepoName("repo1"),
				want: "gitserver-3",
			},
			{
				name: "check we normalise",
				repo: api.RepoName("repo1.git"),
				want: "gitserver-3",
			},
			{
				name: "another repo",
				repo: api.RepoName("github.com/sourcegraph/sourcegraph.git"),
				want: "gitserver-2",
			},
			{
				name: "pinned repo", // different server address that the hashing function would normally yield
				repo: api.RepoName("repo2"),
				want: "gitserver-1",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got := ga.AddrForRepo("gitserver", tc.repo)
				if got != tc.want {
					t.Fatalf("Want %q, got %q", tc.want, got)
				}
			})
		}
	})

	t.Run("deduplicated forks", func(t *testing.T) {
		parentRepo := api.RepoName("github.com/sourcegraph/sourcegraph")
		forkedRepo := api.RepoName("github.com/forked/sourcegraph")

		// At this point no additional config has been set so we expect to get the hashed names
		// directly.
		//
		// This serves both as a test and a test documentation on what shard to expect for which
		// repo.
		shardParentRepo := ga.AddrForRepo("gitserver", parentRepo)
		require.Equal(t, shardParentRepo, "gitserver-2")

		shardForkedRepo := ga.AddrForRepo("gitserver", forkedRepo)
		require.Equal(t, shardForkedRepo, "gitserver-3")

		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				Repositories: &schema.Repositories{
					DeduplicateForks: []string{
						string(parentRepo),
					},
				},
			},
		})

		testCases := []struct {
			name                         string
			getPoolRepoFuncDefaultReturn func() (*types.PoolRepo, error)
			expectedShardParentRepo      string
			expectedShardForkedRepo      string
		}{
			{
				name:                         "valid pool repo",
				getPoolRepoFuncDefaultReturn: func() (*types.PoolRepo, error) { return &types.PoolRepo{RepoName: parentRepo}, nil },
				expectedShardParentRepo:      shardParentRepo,
				expectedShardForkedRepo:      shardParentRepo,
			},
			{
				name:                         "no pool repo",
				getPoolRepoFuncDefaultReturn: func() (*types.PoolRepo, error) { return nil, nil },
				expectedShardParentRepo:      shardParentRepo,
				expectedShardForkedRepo:      shardForkedRepo,
			},
			{
				name:                         "get pool repo returns an error",
				getPoolRepoFuncDefaultReturn: func() (*types.PoolRepo, error) { return nil, errors.New("mocked error") },
				expectedShardParentRepo:      shardParentRepo,
				expectedShardForkedRepo:      shardForkedRepo,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {

				repos := database.NewMockRepoStore()
				repos.GetByNameFunc.PushReturn(
					&types.Repo{
						ID:   api.RepoID(1),
						Name: parentRepo,
						Fork: false,
					}, nil,
				)
				repos.GetByNameFunc.PushReturn(
					&types.Repo{
						ID:   api.RepoID(2),
						Name: forkedRepo,
						Fork: true,
					}, nil,
				)
				db.ReposFunc.SetDefaultReturn(repos)

				gs := database.NewMockGitserverRepoStore()
				gs.GetPoolRepoFunc.SetDefaultReturn(tc.getPoolRepoFuncDefaultReturn())
				db.GitserverReposFunc.SetDefaultReturn(gs)

				require.Equal(t, ga.AddrForRepo("gitserver", parentRepo), tc.expectedShardParentRepo)
				require.Equal(t, ga.AddrForRepo("gitserver", forkedRepo), tc.expectedShardForkedRepo)
			})
		}
	})
}

func TestRepoAddressCache(t *testing.T) {
	repoAddrCache := repoAddressCache{}

	// Read an object that does not exist in the cache.
	item := repoAddrCache.Read("foo")
	require.Nil(t, item)

	// Mock currentTime. Expect this when the value is read.
	expectedFirstWrite := time.Date(2023, 01, 01, 23, 00, 00, 0, time.UTC)
	currentTime = func() time.Time {
		return expectedFirstWrite
	}

	// Now insert an item to the cache.
	repoName := api.RepoName("github.com/foo/bar")
	addr := "127.0.0.1:3080"
	repoAddrCache.Write(repoName, addr)

	// Increment the clock for a subsequent read.
	currentTime = func() time.Time {
		return time.Date(2023, 01, 01, 23, 30, 00, 1, time.UTC)
	}

	item = repoAddrCache.Read(repoName)
	require.NotNil(t, item)
	require.Equal(t, item.address, addr)
	require.Equal(t, item.lastAccessed, expectedFirstWrite)

	// No verify that the item in the cache has the updated timestamp after we Read the item.
	item2 := repoAddrCache.cache[repoName]
	require.Greater(t, item2.lastAccessed, item.lastAccessed)
}

func TestWithUpdateCache(t *testing.T) {
	ga := GitserverAddresses{}

	// Ensures that a nil repoAddressCache will not cause a panic if consumers of GitserverAddresses
	// do not initialise a cache.
	require.Nil(t, ga.repoAddressCache)

	repo := api.RepoName("repo1")
	addr := "address1"
	gotAddress := ga.withUpdateCache(repo, addr)
	require.Equal(t, gotAddress, addr)

	item := ga.repoAddressCache.Read(repo)
	require.Equal(t, item.address, addr)
}

func TestGetCachedRepoAddress(t *testing.T) {
	db := database.NewMockDB()
	ga := GitserverAddresses{
		db:        db,
		Addresses: []string{"gitserver-1", "gitserver-2", "gitserver-3"},
		PinnedServers: map[string]string{
			"repo2": "gitserver-1",
		},
	}

	require.Nil(t, ga.repoAddressCache)

	repo := api.RepoName("repo1")
	require.Equal(t, ga.getCachedRepoAddress(repo), "")

	require.NotNil(t, ga.repoAddressCache)
	require.Equal(t, ga.getCachedRepoAddress(repo), "")

	addr := "address1"
	ga.repoAddressCache.Write(repo, addr)
	require.Equal(t, ga.getCachedRepoAddress(repo), addr)
}
