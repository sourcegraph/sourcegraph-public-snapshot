package resolvers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

const testDiff = `diff README.md README.md
index 671e50a..851b23a 100644
--- README.md
+++ README.md
@@ -1,2 +1,2 @@
 # README
-This file is hosted at example.com and is a test file.
+This file is hosted at sourcegraph.com and is a test file.
diff --git urls.txt urls.txt
index 6f8b5d9..17400bc 100644
--- urls.txt
+++ urls.txt
@@ -1,3 +1,3 @@
 another-url.com
-example.com
+sourcegraph.com
 never-touch-the-mouse.com
`

// testDiffGraphQL is the parsed representation of testDiff.
var testDiffGraphQL = apitest.FileDiffs{
	RawDiff:  testDiff,
	DiffStat: apitest.DiffStat{Changed: 2},
	PageInfo: struct {
		HasNextPage bool
		EndCursor   string
	}{},
	Nodes: []apitest.FileDiff{
		{
			OldPath: "README.md",
			NewPath: "README.md",
			OldFile: apitest.File{Name: "README.md"},
			Hunks: []apitest.FileDiffHunk{
				{
					Body:     " # README\n-This file is hosted at example.com and is a test file.\n+This file is hosted at sourcegraph.com and is a test file.\n",
					OldRange: apitest.DiffRange{StartLine: 1, Lines: 2},
					NewRange: apitest.DiffRange{StartLine: 1, Lines: 2},
				},
			},
			Stat: apitest.DiffStat{Changed: 1},
		},
		{
			OldPath: "urls.txt",
			NewPath: "urls.txt",
			OldFile: apitest.File{Name: "urls.txt"},
			Hunks: []apitest.FileDiffHunk{
				{
					Body:     " another-url.com\n-example.com\n+sourcegraph.com\n never-touch-the-mouse.com\n",
					OldRange: apitest.DiffRange{StartLine: 1, Lines: 3},
					NewRange: apitest.DiffRange{StartLine: 1, Lines: 3},
				},
			},
			Stat: apitest.DiffStat{Changed: 1},
		},
	},
}

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}

func marshalDateTime(t testing.TB, ts time.Time) string {
	t.Helper()

	dt := graphqlbackend.DateTime{Time: ts}

	bs, err := dt.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}

func parseJSONTime(t testing.TB, ts string) time.Time {
	t.Helper()

	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}

	return timestamp
}

func insertTestUser(t *testing.T, db *sql.DB, name string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id", name, isAdmin)

	err := db.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

func newGitHubTestRepo(name string, externalID int) *repos.Repo {
	return &repos.Repo{
		Name: name,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          fmt.Sprintf("external-id-%d", externalID),
			ServiceType: "github",
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*repos.SourceInfo{
			"extsvc:github:4": {
				ID:       "extsvc:github:4",
				CloneURL: fmt.Sprintf("https://secrettoken@%s", name),
			},
		},
	}
}

func mockBackendCommits(t *testing.T, revs ...api.CommitID) {
	t.Helper()

	byRev := map[api.CommitID]struct{}{}
	for _, r := range revs {
		byRev[r] = struct{}{}
	}

	backend.Mocks.Repos.ResolveRev = func(_ context.Context, _ *types.Repo, rev string) (api.CommitID, error) {
		if _, ok := byRev[api.CommitID(rev)]; !ok {
			t.Fatalf("ResolveRev received unexpected rev: %q", rev)
		}
		return api.CommitID(rev), nil
	}
	t.Cleanup(func() { backend.Mocks.Repos.ResolveRev = nil })

	backend.Mocks.Repos.GetCommit = func(_ context.Context, _ *types.Repo, id api.CommitID) (*git.Commit, error) {
		if _, ok := byRev[id]; !ok {
			t.Fatalf("GetCommit received unexpected ID: %s", id)
		}
		return &git.Commit{ID: id}, nil
	}
	t.Cleanup(func() { backend.Mocks.Repos.GetCommit = nil })
}

func mockRepoComparison(t *testing.T, baseRev, headRev, diff string) {
	t.Helper()

	spec := fmt.Sprintf("%s...%s", baseRev, headRev)

	git.Mocks.GetCommit = func(id api.CommitID) (*git.Commit, error) {
		if string(id) != baseRev && string(id) != headRev {
			t.Fatalf("git.Mocks.GetCommit received unknown commit id: %s", id)
		}
		return &git.Commit{ID: api.CommitID(id)}, nil
	}
	t.Cleanup(func() { git.Mocks.GetCommit = nil })

	git.Mocks.ExecReader = func(args []string) (io.ReadCloser, error) {
		if len(args) < 1 && args[0] != "diff" {
			t.Fatalf("gitserver.ExecReader received wrong args: %v", args)
		}

		if have, want := args[len(args)-2], spec; have != want {
			t.Fatalf("gitserver.ExecReader received wrong spec: %q, want %q", have, want)
		}
		return ioutil.NopCloser(strings.NewReader(testDiff)), nil
	}
	t.Cleanup(func() { git.Mocks.ExecReader = nil })

	git.Mocks.MergeBase = func(repo gitserver.Repo, a, b api.CommitID) (api.CommitID, error) {
		if string(a) != baseRev && string(b) != headRev {
			t.Fatalf("git.Mocks.MergeBase received unknown commit ids: %s %s", a, b)
		}
		return a, nil
	}
	t.Cleanup(func() { git.Mocks.MergeBase = nil })
}
