package app

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGuessRepoNameFromRemoteURL(t *testing.T) {
	cases := []struct {
		url               string
		hostnameToPattern map[string]string
		expName           api.RepoName
	}{
		{"github.com:a/b", nil, "github.com/a/b"},
		{"github.com:a/b.git", nil, "github.com/a/b"},
		{"git@github.com:a/b", nil, "github.com/a/b"},
		{"git@github.com:a/b.git", nil, "github.com/a/b"},
		{"ssh://git@github.com/a/b.git", nil, "github.com/a/b"},
		{"ssh://github.com/a/b.git", nil, "github.com/a/b"},
		{"ssh://github.com:1234/a/b.git", nil, "github.com/a/b"},
		{"https://github.com:1234/a/b.git", nil, "github.com/a/b"},
		{"http://alice@foo.com:1234/a/b", nil, "foo.com/a/b"},
		{"github.com:a/b", map[string]string{"github.com": "{hostname}/{path}"}, "github.com/a/b"},
		{"github.com:a/b", map[string]string{"asdf.com": "{hostname}-----{path}"}, "github.com/a/b"},
		{"github.com:a/b", map[string]string{"github.com": "{hostname}-{path}"}, "github.com-a/b"},
		{"github.com:a/b", map[string]string{"github.com": "{path}"}, "a/b"},
		{"github.com:a/b", map[string]string{"github.com": "{hostname}"}, "github.com"},
		{"github.com:a/b", map[string]string{"github.com": "github/{path}", "asdf.com": "asdf/{path}"}, "github/a/b"},
		{"asdf.com:a/b", map[string]string{"github.com": "github/{path}", "asdf.com": "asdf/{path}"}, "asdf/a/b"},
	}
	for _, c := range cases {
		if got, want := guessRepoNameFromRemoteURL(c.url, c.hostnameToPattern), c.expName; got != want {
			t.Errorf("%+v: got %q, want %q", c, got, want)
		}
	}
}

func TestEditorRev(t *testing.T) {
	repoName := api.RepoName("myRepo")
	backend.Mocks.Repos.ResolveRev = func(v0 context.Context, repo *types.Repo, rev string) (api.CommitID, error) {
		if rev == "branch" {
			return api.CommitID(strings.Repeat("b", 40)), nil
		}
		if rev == "" || rev == "defaultBranch" {
			return api.CommitID(strings.Repeat("d", 40)), nil
		}
		if len(rev) == 40 {
			return api.CommitID(rev), nil
		}
		t.Fatalf("unexpected RepoRev request rev: %q", rev)
		return "", nil
	}
	backend.Mocks.Repos.GetByName = func(v0 context.Context, name api.RepoName) (*types.Repo, error) {
		return &types.Repo{
			ID:   api.RepoID(1),
			Name: name,
		}, nil
	}
	ctx := context.Background()

	cases := []struct {
		inputRev     string
		expEditorRev string
		beExplicit   bool
	}{
		{strings.Repeat("a", 40), "@" + strings.Repeat("a", 40), false},
		{"branch", "@branch", false},
		{"", "", false},
		{"defaultBranch", "", false},
		{strings.Repeat("d", 40), "", false},                           // default revision
		{strings.Repeat("d", 40), "@" + strings.Repeat("d", 40), true}, // default revision, explicit
	}
	for _, c := range cases {
		got, err := editorRev(ctx, repoName, c.inputRev, c.beExplicit)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.expEditorRev {
			t.Errorf("On input rev %q: got %q, want %q", c.inputRev, got, c.expEditorRev)
		}
	}
}
