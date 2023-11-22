package resolvers

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	githubapp "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/githubappauth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/batches/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var update = flag.Bool("update", false, "update testdata")

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

var testDiff = []byte(`diff README.md README.md
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
`)

// testDiffGraphQL is the parsed representation of testDiff.
var testDiffGraphQL = apitest.FileDiffs{
	TotalCount: 2,
	RawDiff:    string(testDiff),
	DiffStat:   apitest.DiffStat{Added: 2, Deleted: 2},
	PageInfo:   apitest.PageInfo{},
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
			Stat: apitest.DiffStat{Added: 1, Deleted: 1},
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
			Stat: apitest.DiffStat{Added: 1, Deleted: 1},
		},
	},
}

func marshalDateTime(t testing.TB, ts time.Time) string {
	t.Helper()

	dt := gqlutil.DateTime{Time: ts}

	bs, err := dt.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	// Unquote the date time.
	return strings.ReplaceAll(string(bs), "\"", "")
}

func parseJSONTime(t testing.TB, ts string) time.Time {
	t.Helper()

	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}

	return timestamp
}

func newSchema(db database.DB, bcr graphqlbackend.BatchChangesResolver) (*graphql.Schema, error) {
	ghar := githubapp.NewResolver(log.NoOp(), db)
	return graphqlbackend.NewSchemaWithBatchChangesResolver(db, bcr, ghar)
}

func newGitHubExternalService(t *testing.T, store database.ExternalServiceStore) *types.ExternalService {
	t.Helper()

	clock := timeutil.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	svc := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		// The authorization field is needed to enforce permissions
		Config:    extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "authorization": {}, "token": "abc", "repos": ["owner/name"]}`),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.Upsert(context.Background(), &svc); err != nil {
		t.Fatalf("failed to insert external services: %v", err)
	}

	return &svc
}

func newGitHubTestRepo(name string, externalService *types.ExternalService) *types.Repo {
	return &types.Repo{
		Name:    api.RepoName(name),
		Private: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          fmt.Sprintf("external-id-%d", externalService.ID),
			ServiceType: "github",
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			externalService.URN(): {
				ID:       externalService.URN(),
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
}

func mockRepoComparison(t *testing.T, gitserverClient *gitserver.MockClient, baseRev, headRev string, diff []byte) {
	t.Helper()

	spec := fmt.Sprintf("%s...%s", baseRev, headRev)
	gitserverClientWithExecReader := gitserver.NewMockClientWithExecReader(nil, func(_ context.Context, _ api.RepoName, args []string) (io.ReadCloser, error) {
		if len(args) < 1 && args[0] != "diff" {
			t.Fatalf("gitserver.ExecReader received wrong args: %v", args)
		}

		if have, want := args[len(args)-2], spec; have != want {
			t.Fatalf("gitserver.ExecReader received wrong spec: %q, want %q", have, want)
		}
		return io.NopCloser(bytes.NewReader(diff)), nil
	})

	gitserverClientWithExecReader.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		if spec != baseRev && spec != headRev {
			t.Fatalf("gitserver.Mocks.ResolveRevision received unknown spec: %s", spec)
		}
		return api.CommitID(spec), nil
	})

	gitserverClientWithExecReader.MergeBaseFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, a api.CommitID, b api.CommitID) (api.CommitID, error) {
		if string(a) != baseRev && string(b) != headRev {
			t.Fatalf("git.Mocks.MergeBase received unknown commit ids: %s %s", a, b)
		}
		return a, nil
	})
	*gitserverClient = *gitserverClientWithExecReader
}

func addChangeset(t *testing.T, ctx context.Context, s *store.Store, c *btypes.Changeset, batchChange int64) {
	t.Helper()

	c.BatchChanges = append(c.BatchChanges, btypes.BatchChangeAssoc{BatchChangeID: batchChange})
	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatal(err)
	}
}

func pruneUserCredentials(t *testing.T, db database.DB, key encryption.Key) {
	t.Helper()
	ctx := actor.WithInternalActor(context.Background())
	creds, _, err := db.UserCredentials(key).List(ctx, database.UserCredentialsListOpts{})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range creds {
		if err := db.UserCredentials(key).Delete(ctx, c.ID); err != nil {
			t.Fatal(err)
		}
	}
}

func pruneSiteCredentials(t *testing.T, bstore *store.Store) {
	t.Helper()
	creds, _, err := bstore.ListSiteCredentials(context.Background(), store.ListSiteCredentialsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range creds {
		if err := bstore.DeleteSiteCredential(context.Background(), c.ID); err != nil {
			t.Fatal(err)
		}
	}
}
