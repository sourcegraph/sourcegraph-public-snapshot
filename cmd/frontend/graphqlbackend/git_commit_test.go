package graphqlbackend

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestGitCommitResolver(t *testing.T) {
	t.Run("All fields", func(t *testing.T) {
		git.Mocks.GetCommit = func(id api.CommitID) (*git.Commit, error) {
			return &git.Commit{
				ID:      id,
				Message: "Changes things",
				Parents: []api.CommitID{"p1", "p2"},
				Author: git.Signature{
					Name:  "Bob",
					Email: "bob@alice.com",
					Date:  time.Now(),
				},
				Committer: &git.Signature{
					Name:  "Alice",
					Email: "alice@bob.com",
					Date:  time.Now(),
				},
			}, nil
		}
		defer func() { git.Mocks.GetCommit = nil }()
	})

}

func TestGitCommitBody(t *testing.T) {
	tests := map[string]string{
		"hello":               "",
		"hello\n":             "",
		"hello\n\n":           "",
		"hello\nworld":        "world",
		"hello\n\nworld":      "world",
		"hello\n\nworld\nfoo": "world\nfoo",
	}
	for input, want := range tests {
		got := GitCommitBody(input)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
