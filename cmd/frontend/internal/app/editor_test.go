package app

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
		return &types.Repo{ID: api.RepoID(1), Name: name},

			nil
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

func TestEditorRedirect(t *testing.T) {
	cases := []struct {
		name            string
		q               url.Values
		wantRedirectURL string
		wantUserErr     error
		wantInternalErr error
	}{
		{
			name: "open file",
			q: url.Values{
				"editor":     []string{"Atom"},
				"version":    []string{"v1.2.1"},
				"remote_url": []string{"git@github.com:a/b"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
				"start_row":  []string{"123"},
				"start_col":  []string{"1"},
				"end_row":    []string{"123"},
				"end_col":    []string{"10"},
			},
			wantRedirectURL: "/github.com/a/b@0ad12f/-/blob/mux.go?utm_source=Atom-v1.2.1#L124:2-124:11",
		},
		{
			name: "open file no selection",
			q: url.Values{
				"editor":     []string{"Atom"},
				"version":    []string{"v1.2.1"},
				"remote_url": []string{"git@github.com:a/b"},
				"branch":     []string{"dev"},
				"revision":   []string{"0ad12f"},
				"file":       []string{"mux.go"},
			},
			wantRedirectURL: "/github.com/a/b@0ad12f/-/blob/mux.go?utm_source=Atom-v1.2.1#L1:1", // L1:1 is expected (but could be nicer by omitting it)
		},
		{
			name: "search",
			q: url.Values{
				"editor":  []string{"Atom"},
				"version": []string{"v1.2.1"},
				"search":  []string{"foobar"},
			},
			wantRedirectURL: "/search?q=%22foobar%22&utm_source=Atom-v1.2.1",
		},
		{
			// BUG!
			name:            "empty request",
			q:               url.Values{},
			wantRedirectURL: "/-/blob?utm_source=-#L1:1",
		},
		{
			// BUG!
			name: "unknown request",
			q: url.Values{
				"editor":  []string{"Atom"},
				"version": []string{"v1.2.1"},
			},
			wantRedirectURL: "/-/blob?utm_source=Atom-v1.2.1#L1:1",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			redirectURL, userErr, internalErr := editorRedirect(context.Background(), c.q)
			if redirectURL != c.wantRedirectURL {
				t.Fatalf("got redirectURL %q want %q", redirectURL, c.wantRedirectURL)
			}
			if userErr != c.wantUserErr {
				t.Fatalf("got userErr %q want %q", userErr, c.wantUserErr)
			}
			if internalErr != c.wantInternalErr {
				t.Fatalf("got internalErr %q want %q", internalErr, c.wantInternalErr)
			}
		})
	}
}
