package graphqlbackend

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestGitCommitResolver(t *testing.T) {
	ctx := context.Background()

	commit := &git.Commit{
		ID:      "c1",
		Message: "subject: Changes things\nBody of changes",
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
	}

	t.Run("Lazy loading", func(t *testing.T) {
		git.Mocks.GetCommit = func(api.CommitID) (*git.Commit, error) {
			return commit, nil
		}
		t.Cleanup(func() {
			git.Mocks.GetCommit = nil
		})

		for _, tc := range []struct {
			name string
			want interface{}
			have func(*GitCommitResolver) (interface{}, error)
		}{{
			name: "author",
			want: toSignatureResolver(&commit.Author, true),
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.Author(ctx)
			},
		}, {
			name: "committer",
			want: toSignatureResolver(commit.Committer, true),
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.Committer(ctx)
			},
		}, {
			name: "message",
			want: commit.Message,
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.Message(ctx)
			},
		}, {
			name: "subject",
			want: "subject: Changes things",
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.Subject(ctx)
			},
		}, {
			name: "body",
			want: "Body of changes",
			have: func(r *GitCommitResolver) (interface{}, error) {
				s, err := r.Body(ctx)
				return *s, err
			},
		}, {
			name: "url",
			want: "/bob-repo/-/commit/c1",
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.URL()
			},
		}, {
			name: "canonical-url",
			want: "/bob-repo/-/commit/c1",
			have: func(r *GitCommitResolver) (interface{}, error) {
				return r.CanonicalURL()
			},
		}} {
			t.Run(tc.name, func(t *testing.T) {
				repo := NewRepositoryResolver(&types.Repo{Name: "bob-repo"})
				// We pass no commit here to test that it gets lazy loaded via
				// the git.GetCommit mock above.
				r := toGitCommitResolver(repo, "c1", nil)

				have, err := tc.have(r)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(have, tc.want) {
					t.Errorf("\nhave: %s\nwant: %s", spew.Sprint(have), spew.Sprint(tc.want))
				}
			})
		}
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
