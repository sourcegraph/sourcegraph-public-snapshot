package gitserver

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
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
	ctx := context.Background()
	logger := logtest.Scoped(t)

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
				got := ga.AddrForRepo(ctx, logger, "gitserver", tc.repo)
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
		shardParentRepo := ga.AddrForRepo(ctx, logger, "gitserver", parentRepo)
		require.Equal(t, "gitserver-2", shardParentRepo)

		shardForkedRepo := ga.AddrForRepo(ctx, logger, "gitserver", forkedRepo)
		require.Equal(t, "gitserver-3", shardForkedRepo)

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

				require.Equal(t, tc.expectedShardParentRepo, ga.AddrForRepo(ctx, logger, "gitserver", parentRepo))
				require.Equal(t, tc.expectedShardForkedRepo, ga.AddrForRepo(ctx, logger, "gitserver", forkedRepo))
			})
		}
	})
}

func TestRepoAddressCache(t *testing.T) {
	repoAddrCache := repoAddressCache{}

	// Read an object that does not exist in the cache.
	item := repoAddrCache.Read("foo")
	require.Nil(t, item)

	// Now insert an item to the cache.
	repoName := api.RepoName("github.com/foo/bar")
	addr := "127.0.0.1:3080"
	repoAddrCache.Write(repoName, addr)

	cachedItem := repoAddrCache.cache[repoName]

	item = repoAddrCache.Read(repoName)
	require.NotNil(t, item)
	require.Equal(t, addr, item.address)
	require.Equal(t, cachedItem.expiresAt, item.expiresAt)

	// Now verify that the item in the cache when read again is the same, that is we did not update
	// the cached on the Read path.
	//
	// The following test may seem unnecessary looking at the current design of the cache, but the
	// first version of this cache during development was updating the timestamp of the cached item
	// on the read path. This test is to ensure that is not happening anymore.
	item2 := repoAddrCache.cache[repoName]
	require.NotNil(t, item2)
	require.Equal(t, item.address, item2.address)
	require.Equal(t, item.expiresAt, item2.expiresAt)

	// Mock now to be 17 minutes in the past such that the next time the item is read, it will be
	// expired.
	now := time.Now().Add(-17 * time.Minute)
	repoAddrCache.cache[repoName] = repoAddressCachedItem{
		address:   addr,
		expiresAt: now,
	}

	require.Nil(t, repoAddrCache.Read(repoName))
}

func TestGitserverAddresses_withUpdateCache(t *testing.T) {
	ga := GitserverAddresses{}

	// Ensures that a nil repoAddressCache will not cause a panic if consumers of GitserverAddresses
	// do not initialise a cache.
	require.Nil(t, ga.repoAddressCache)

	repo := api.RepoName("repo1")
	addr := "address1"
	gotAddress := ga.withUpdateCache(repo, addr)
	require.Equal(t, addr, gotAddress)

	item := ga.repoAddressCache.Read(repo)
	require.Equal(t, addr, item.address)
}

func TestGitserverAddresses_getCachedRepoAddress(t *testing.T) {
	db := database.NewMockDB()
	ga := &GitserverAddresses{
		db:        db,
		Addresses: []string{"gitserver-1", "gitserver-2", "gitserver-3"},
		PinnedServers: map[string]string{
			"repo2": "gitserver-1",
		},
	}

	require.Nil(t, ga.repoAddressCache)

	repo := api.RepoName("repo1")
	require.Equal(t, "", ga.getCachedRepoAddress(repo))

	require.NotNil(t, ga.repoAddressCache)
	require.Equal(t, "", ga.getCachedRepoAddress(repo))

	addr := "address1"
	ga.repoAddressCache.Write(repo, addr)
	require.Equal(t, addr, ga.getCachedRepoAddress(repo))
}
