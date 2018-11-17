package app

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGuessRepoNameFromRemoteURL(t *testing.T) {
	tests := map[string]api.RepoName{
		"github.com:a/b":                  "github.com/a/b",
		"github.com:a/b.git":              "github.com/a/b",
		"git@github.com:a/b":              "github.com/a/b",
		"git@github.com:a/b.git":          "github.com/a/b",
		"ssh://git@github.com/a/b.git":    "github.com/a/b",
		"ssh://github.com/a/b.git":        "github.com/a/b",
		"ssh://github.com:1234/a/b.git":   "github.com/a/b",
		"https://github.com:1234/a/b.git": "github.com/a/b",
		"http://alice@foo.com:1234/a/b":   "foo.com/a/b",
	}
	for input, want := range tests {
		got := guessRepoNameFromRemoteURL(input)
		if got != want {
			t.Errorf("%s: got %q, want %q", input, got, want)
		}
	}
}

func TestEditorRef(t *testing.T) {
	ctx := testContext()
	repoName := api.RepoName("myRepo")

	db.Mocks.Repos.MockGetByName(t, repoName, 1)

	type BranchAndRevision struct {
		branchName string
		revision   string
	}
	tests := map[BranchAndRevision]string{
		BranchAndRevision{"", "sha1"}:       "@sha1",
		BranchAndRevision{"branch", ""}:     "@branch",
		BranchAndRevision{"branch", "sha2"}: "@sha2",
	}
	for input, want := range tests {
		got, _ := editorRef(ctx, repoName, input.branchName, input.revision)

		if got != want {
			t.Errorf("%s: got %q, want %q", input, got, want)
		}
	}
}
