package gitserver

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
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
