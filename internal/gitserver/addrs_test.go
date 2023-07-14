package gitserver

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/stretchr/testify/require"
)

func TestAddrForRepo(t *testing.T) {
	ga := GitserverAddresses{
		Addresses: []string{"gitserver-1", "gitserver-2", "gitserver-3"},
		PinnedServers: map[string]string{
			"repo2": "gitserver-1",
		},
	}

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
