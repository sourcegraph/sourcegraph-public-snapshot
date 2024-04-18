package connection

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAddrForRepo(t *testing.T) {
	ga := GitserverAddresses{
		Addresses: []string{"gitserver-1", "gitserver-2", "gitserver-3"},
		PinnedServers: map[string]string{
			"repo2": "gitserver-1",
		},
	}
	ctx := context.Background()

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
			name: "check we target the original instance prior to deletion",
			repo: api.RepoName("DELETED-123123.123123-repo1"),
			want: "gitserver-3",
		},
		{
			name: "deletion and pinning work together",
			repo: api.RepoName("DELETED-123123.123123-repo2"),
			want: "gitserver-1",
		},
		{
			name: "another repo",
			repo: api.RepoName("github.com/sourcegraph/sourcegraph"),
			want: "gitserver-2",
		},
		{
			name: "case sensitive repo",
			repo: api.RepoName("github.com/sourcegraph/Sourcegraph"),
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
			got := ga.AddrForRepo(ctx, tc.repo)
			if got != tc.want {
				t.Fatalf("Want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestGitserverAddresses_AddrForRepo_PinnedRepos(t *testing.T) {
	addrs := NewGitserverAddresses(newConfig(
		[]string{"gitserver1", "gitserver2"},
		map[string]string{"repo1": "gitserver2"},
	))

	ctx := context.Background()

	addr := addrs.AddrForRepo(ctx, "repo1")
	require.Equal(t, "gitserver2", addr)

	// simulate config change - site admin manually changes the pinned repo config
	addrs = NewGitserverAddresses(newConfig(
		[]string{"gitserver1", "gitserver2"},
		map[string]string{"repo1": "gitserver1"},
	))

	require.Equal(t, "gitserver1", addrs.AddrForRepo(ctx, "repo1"))
}

func newConfig(addrs []string, pinned map[string]string) *conf.Unified {
	return &conf.Unified{
		ServiceConnectionConfig: conftypes.ServiceConnections{
			GitServers: addrs,
		},
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				GitServerPinnedRepos: pinned,
			},
		},
	}
}

func BenchmarkAddrForKey(b *testing.B) {
	for _, count := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("Count-%d", count), func(b *testing.B) {
			var nodes []string
			for i := range count {
				nodes = append(nodes, fmt.Sprintf("Node%d", i))
			}
			b.ResetTimer()
			for range b.N {
				addrForKey("foo", nodes)
			}
		})
	}
}
