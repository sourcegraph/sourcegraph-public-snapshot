package campaigns

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func testMigratePatchesWithoutDiffStats(db *sql.DB, userID int32) func(*testing.T) {
	const testDiff = `diff --git INSTALL.md INSTALL.md
index b9f9438..cb1ab9f 100644
--- INSTALL.md
+++ INSTALL.md
@@ -2,4 +2,4 @@

 Foobar

-barfoo
+Pfannkuchen
diff --git README.md README.md
index 437e1a8..540f2f3 100644
--- README.md
+++ README.md
@@ -1,5 +1,5 @@
 # README

 Line 1
-Line 2
+Line Foobar
 Line 3
diff --git main.c main.c
new file mode 100644
index 0000000..44e82e2
--- /dev/null
+++ main.c
@@ -0,0 +1 @@
+int main(int argc, char *argv[]) { return 0; }
`

	return func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		clock := func() time.Time { return now.UTC().Truncate(time.Microsecond) }
		ctx := context.Background()

		reposStore := repos.NewDBStore(db, sql.TxOptions{})
		repo := testRepo(1, extsvc.TypeGitHub)
		if err := reposStore.UpsertRepos(context.Background(), repo); err != nil {
			t.Fatal(err)
		}

		s := NewStoreWithClock(db, clock)

		patches := make([]*cmpgn.Patch, 0, 3)

		for i := 0; i < cap(patches); i++ {
			patchSet := &cmpgn.PatchSet{UserID: userID}
			err := s.CreatePatchSet(context.Background(), patchSet)
			if err != nil {
				t.Fatal(err)
			}

			p := &cmpgn.Patch{
				PatchSetID: patchSet.ID,
				RepoID:     repo.ID,
				Rev:        api.CommitID("deadbeef"),
				BaseRef:    "master",
				Diff:       testDiff,
			}

			err = s.CreatePatch(ctx, p)
			if err != nil {
				t.Fatal(err)
			}

			patches = append(patches, p)
		}

		withoutStats, _, err := s.ListPatches(ctx, ListPatchesOpts{OnlyWithoutDiffStats: true})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(withoutStats), len(patches); have != want {
			t.Fatalf("wrong number of patches without stats. have=%d, want=%d", have, want)
		}

		err = MigratePatchesWithoutDiffStats(ctx, s)
		if err != nil {
			t.Fatal(err)
		}

		withoutStats, _, err = s.ListPatches(ctx, ListPatchesOpts{OnlyWithoutDiffStats: true})
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(withoutStats), 0; have != want {
			t.Fatalf("wrong number of patches without stats. have=%d, want=%d", have, want)
		}
	}
}
