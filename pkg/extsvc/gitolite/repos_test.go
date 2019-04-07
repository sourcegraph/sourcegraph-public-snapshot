package gitolite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_decodeRepos(t *testing.T) {
	tests := []struct {
		host         string
		gitoliteInfo string
		expRepos     []*Repo
	}{
		{
			host: "git@gitolite.example.com",
			gitoliteInfo: `hello admin, this is git@gitolite-799486b5db-ghrxg running gitolite3 v3.6.6-0-g908f8c6 on git 2.7.4

		 R W    gitolite-admin
		 R W    repowith@sign
		 R W    testing
		`,
			expRepos: []*Repo{
				{Name: "gitolite-admin", URL: "git@gitolite.example.com:gitolite-admin"},
				{Name: "repowith@sign", URL: "git@gitolite.example.com:repowith@sign"},
				{Name: "testing", URL: "git@gitolite.example.com:testing"},
			},
		},
		{
			host:         "git@gitolite.example.com",
			gitoliteInfo: "",
			expRepos:     nil,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			repos := decodeRepos(test.host, test.gitoliteInfo)
			if diff := cmp.Diff(test.expRepos, repos); diff != "" {
				t.Error(diff)
			}
		})
	}
}
