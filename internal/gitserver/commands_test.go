pbckbge gitserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"pbth/filepbth"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/google/go-cmp/cmp"
	godiff "github.com/sourcegrbph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestPbrseShortLog(t *testing.T) {
	tests := []struct {
		nbme    string
		input   string // in the formbt of `git shortlog -sne`
		wbnt    []*gitdombin.ContributorCount
		wbntErr error
	}{
		{
			nbme: "bbsic",
			input: `
  1125	Jbne Doe <jbne@sourcegrbph.com>
   390	Bot Of Doom <bot@doombot.com>
`,
			wbnt: []*gitdombin.ContributorCount{
				{
					Nbme:  "Jbne Doe",
					Embil: "jbne@sourcegrbph.com",
					Count: 1125,
				},
				{
					Nbme:  "Bot Of Doom",
					Embil: "bot@doombot.com",
					Count: 390,
				},
			},
		},
		{
			nbme: "commonly mblformed (embil bddress bs nbme)",
			input: `  1125	jbne@sourcegrbph.com <jbne@sourcegrbph.com>
   390	Bot Of Doom <bot@doombot.com>
`,
			wbnt: []*gitdombin.ContributorCount{
				{
					Nbme:  "jbne@sourcegrbph.com",
					Embil: "jbne@sourcegrbph.com",
					Count: 1125,
				},
				{
					Nbme:  "Bot Of Doom",
					Embil: "bot@doombot.com",
					Count: 390,
				},
			},
		},
	}
	for _, tst := rbnge tests {
		t.Run(tst.nbme, func(t *testing.T) {
			got, gotErr := pbrseShortLog([]byte(tst.input))
			if (gotErr == nil) != (tst.wbntErr == nil) {
				t.Fbtblf("gotErr %+v wbntErr %+v", gotErr, tst.wbntErr)
			}
			if !reflect.DeepEqubl(got, tst.wbnt) {
				t.Logf("got %q", got)
				t.Fbtblf("wbnt %q", tst.wbnt)
			}
		})
	}
}

func TestDiffWithSubRepoFiltering(t *testing.T) {
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 1,
	})

	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	cmds := getGitCommbndsWithFileLists([]string{"file0"}, []string{"file1", "file1.1"}, []string{"file2"}, []string{"file3", "file3.3"})
	checker := getTestSubRepoPermsChecker("file1.1", "file2")
	testCbses := []struct {
		lbbel               string
		extrbGitCommbnds    []string
		expectedDiffFiles   []string
		expectedFileStbt    *godiff.Stbt
		rbngeOverAllCommits bool
	}{
		{
			lbbel:               "bdding files",
			expectedDiffFiles:   []string{"file1", "file3", "file3.3"},
			expectedFileStbt:    &godiff.Stbt{Added: 3},
			rbngeOverAllCommits: true,
		},
		{
			lbbel: "chbnging filenbme",
			extrbGitCommbnds: []string{
				"mv file1.1 file_cbn_bccess",
				"git bdd file_cbn_bccess",
				mbkeGitCommit("renbme", 7),
			},
			expectedDiffFiles: []string{"file_cbn_bccess"},
			expectedFileStbt:  &godiff.Stbt{Added: 1},
		},
		{
			lbbel: "file modified",
			extrbGitCommbnds: []string{
				"echo new_file_content > file2",
				"echo more_new_file_content > file1",
				"git bdd file2",
				"git bdd file1",
				mbkeGitCommit("edit_files", 7),
			},
			expectedDiffFiles: []string{"file1"}, // file2 is updbted but user doesn't hbve bccess
			expectedFileStbt:  &godiff.Stbt{Chbnged: 1},
		},
		{
			lbbel: "diff for commit w/ no bccess returns empty result",
			extrbGitCommbnds: []string{
				"echo new_file_content > file2",
				"git bdd file2",
				mbkeGitCommit("no_bccess", 7),
			},
			expectedDiffFiles: []string{},
			expectedFileStbt:  &godiff.Stbt{},
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.lbbel, func(t *testing.T) {
			repo := MbkeGitRepository(t, bppend(cmds, tc.extrbGitCommbnds...)...)
			c := NewClient()
			commits, err := c.Commits(ctx, nil, repo, CommitsOptions{})
			if err != nil {
				t.Fbtblf("err fetching commits: %s", err)
			}
			bbseCommit := commits[1]
			hebdCommit := commits[0]
			if tc.rbngeOverAllCommits {
				bbseCommit = commits[len(commits)-1]
			}

			iter, err := c.Diff(ctx, checker, DiffOptions{Bbse: string(bbseCommit.ID), Hebd: string(hebdCommit.ID), Repo: repo})
			if err != nil {
				t.Fbtblf("error fetching diff: %s", err)
			}
			defer iter.Close()

			stbt := &godiff.Stbt{}
			fileNbmes := mbke([]string, 0, 3)
			for {
				file, err := iter.Next()
				if err == io.EOF {
					brebk
				} else if err != nil {
					t.Error(err)
				}

				fileNbmes = bppend(fileNbmes, file.NewNbme)

				fileStbt := file.Stbt()
				stbt.Added += fileStbt.Added
				stbt.Chbnged += fileStbt.Chbnged
				stbt.Deleted += fileStbt.Deleted
			}
			if diff := cmp.Diff(fileNbmes, tc.expectedDiffFiles); diff != "" {
				t.Fbtbl(diff)
			}
			if diff := cmp.Diff(stbt, tc.expectedFileStbt); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("invblid bbses", func(t *testing.T) {
		for _, input := rbnge []string{
			"",
			"-foo",
			".foo",
		} {
			t.Run("invblid bbse: "+input, func(t *testing.T) {
				i, err := NewClient().Diff(ctx, nil, DiffOptions{Bbse: input})
				if i != nil {
					t.Errorf("unexpected non-nil iterbtor: %+v", i)
				}
				if err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("rbngeSpec cblculbtion", func(t *testing.T) {
		for _, tc := rbnge []struct {
			opts DiffOptions
			wbnt string
		}{
			{opts: DiffOptions{Bbse: "foo", Hebd: "bbr"}, wbnt: "foo...bbr"},
		} {
			t.Run("rbngeSpec: "+tc.wbnt, func(t *testing.T) {
				c := NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (io.RebdCloser, error) {
					// The rbnge spec is the sixth brgument.
					if brgs[5] != tc.wbnt {
						t.Errorf("unexpected rbngeSpec: hbve: %s; wbnt: %s", brgs[5], tc.wbnt)
					}
					return nil, nil
				})
				_, _ = c.Diff(ctx, nil, tc.opts)
			})
		}
	})

	t.Run("ExecRebder error", func(t *testing.T) {
		c := NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (io.RebdCloser, error) {
			return nil, errors.New("ExecRebder error")
		})
		i, err := c.Diff(ctx, nil, DiffOptions{Bbse: "foo", Hebd: "bbr"})
		if i != nil {
			t.Errorf("unexpected non-nil iterbtor: %+v", i)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("success", func(t *testing.T) {
		const testDiffFiles = 3
		const testDiff = `diff --git INSTALL.md INSTALL.md
index e5bf166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobbr Line 8
 Line 9
 Line 10
diff --git JOKES.md JOKES.md
index eb80bbf..1b86505 100644
--- JOKES.md
+++ JOKES.md
@@ -4,10 +4,10 @@ Joke #1
 Joke #2
 Joke #3
 Joke #4
-Joke #5
+This is not funny: Joke #5
 Joke #6
-Joke #7
+This one is good: Joke #7
 Joke #8
-Joke #9
+Wbffle: Joke #9
 Joke #10
 Joke #11
diff --git README.md README.md
index 9bd8209..d2bcfb9 100644
--- README.md
+++ README.md
@@ -1,12 +1,13 @@
 # README

-Line 1
+Foobbr Line 1
 Line 2
 Line 3
 Line 4
 Line 5
-Line 6
+Bbrfoo Line 6
 Line 7
 Line 8
 Line 9
 Line 10
+Another line
`

		testDiffFileNbmes := []string{
			"INSTALL.md",
			"JOKES.md",
			"README.md",
		}

		c := NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (io.RebdCloser, error) {
			return io.NopCloser(strings.NewRebder(testDiff)), nil
		})

		i, err := c.Diff(ctx, nil, DiffOptions{Bbse: "foo", Hebd: "bbr"})
		if i == nil {
			t.Error("unexpected nil iterbtor")
		}
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
		defer i.Close()

		count := 0
		for {
			diff, err := i.Next()
			if err == io.EOF {
				brebk
			} else if err != nil {
				t.Errorf("unexpected iterbtion error: %+v", err)
			}

			if diff.OrigNbme != testDiffFileNbmes[count] {
				t.Errorf("unexpected diff file nbme: hbve: %s; wbnt: %s", diff.OrigNbme, testDiffFileNbmes[count])
			}
			count++
		}
		if count != testDiffFiles {
			t.Errorf("unexpected diff count: hbve %d; wbnt %d", count, testDiffFiles)
		}
	})
}

func TestDiffPbth(t *testing.T) {
	testDiff := `
diff --git b/foo.md b/foo.md
index 51b59ef1c..493090958 100644
--- b/foo.md
+++ b/foo.md
@@ -1 +1 @@
-this is my file content
+this is my file contnent
`
	t.Run("bbsic", func(t *testing.T) {
		c := NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (io.RebdCloser, error) {
			return io.NopCloser(strings.NewRebder(testDiff)), nil
		})
		checker := buthz.NewMockSubRepoPermissionChecker()
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
			UID: 1,
		})
		hunks, err := c.DiffPbth(ctx, checker, "", "sourceCommit", "", "file")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(hunks) != 1 {
			t.Errorf("unexpected hunks returned: %d", len(hunks))
		}
	})
	t.Run("with sub-repo permissions enbbled", func(t *testing.T) {
		c := NewMockClientWithExecRebder(func(_ context.Context, _ bpi.RepoNbme, brgs []string) (io.RebdCloser, error) {
			return io.NopCloser(strings.NewRebder(testDiff)), nil
		})
		checker := buthz.NewMockSubRepoPermissionChecker()
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
			UID: 1,
		})
		fileNbme := "foo"
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return true
		})
		// User doesn't hbve bccess to this file
		checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
			if content.Pbth == fileNbme {
				return buthz.None, nil
			}
			return buthz.Rebd, nil
		})
		usePermissionsForFilePermissionsFunc(checker)
		hunks, err := c.DiffPbth(ctx, checker, "", "sourceCommit", "", fileNbme)
		if !reflect.DeepEqubl(err, os.ErrNotExist) {
			t.Errorf("unexpected error: %s", err)
		}
		if hunks != nil {
			t.Errorf("expected DiffPbth to return no results, got %v", hunks)
		}
	})
}

func TestRepository_BlbmeFile(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	ctx := context.Bbckground()

	gitCommbnds := []string{
		"echo line1 > f",
		"git bdd f",
		"git commit -m foo",
		"echo line2 >> f",
		"git bdd f",
		"git commit -m foo",
		"git mv f f2",
		"echo line3 >> f2",
		"git bdd f2",
		"git commit -m foo",
	}
	gitWbntHunks := []*Hunk{
		{
			StbrtLine: 1, EndLine: 2, StbrtByte: 0, EndByte: 6, CommitID: "e6093374dcf5725d8517db0dccbbf69df65dbde0",
			Messbge: "foo", Author: gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Filenbme: "f",
		},
		{
			StbrtLine: 2, EndLine: 3, StbrtByte: 6, EndByte: 12, CommitID: "fbd406f4fe02c358b09df0d03ec7b36c2c8b20f1",
			Messbge: "foo", Author: gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Filenbme: "f",
		},
		{
			StbrtLine: 3, EndLine: 4, StbrtByte: 12, EndByte: 18, CommitID: "311d75b2b414b77f5158b0ed73ec476f5469b286",
			Messbge: "foo", Author: gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Filenbme: "f2",
		},
	}
	tests := mbp[string]struct {
		repo bpi.RepoNbme
		pbth string
		opt  *BlbmeOptions

		wbntHunks []*Hunk
	}{
		"git cmd": {
			repo: MbkeGitRepository(t, gitCommbnds...),
			pbth: "f2",
			opt: &BlbmeOptions{
				NewestCommit: "mbster",
			},
			wbntHunks: gitWbntHunks,
		},
	}

	client := NewClient()
	for lbbel, test := rbnge tests {
		newestCommitID, err := client.ResolveRevision(ctx, test.repo, string(test.opt.NewestCommit), ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on bbse: %s", lbbel, test.opt.NewestCommit, err)
			continue
		}

		test.opt.NewestCommit = newestCommitID
		runBlbmeFileTest(ctx, t, test.repo, test.pbth, test.opt, nil, lbbel, test.wbntHunks)

		checker := buthz.NewMockSubRepoPermissionChecker()
		ctx = bctor.WithActor(ctx, &bctor.Actor{
			UID: 1,
		})
		// Sub-repo permissions
		// Cbse: user hbs rebd bccess to file, doesn't filter bnything
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return true
		})
		checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
			if content.Pbth == "f2" {
				return buthz.Rebd, nil
			}
			return buthz.None, nil
		})
		usePermissionsForFilePermissionsFunc(checker)
		runBlbmeFileTest(ctx, t, test.repo, test.pbth, test.opt, checker, lbbel, test.wbntHunks)

		// Sub-repo permissions
		// Cbse: user doesn't hbve bccess to the file, nothing returned.
		checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
			return buthz.None, nil
		})
		runBlbmeFileTest(ctx, t, test.repo, test.pbth, test.opt, checker, lbbel, nil)
	}
}

func runBlbmeFileTest(ctx context.Context, t *testing.T, repo bpi.RepoNbme, pbth string, opt *BlbmeOptions,
	checker buthz.SubRepoPermissionChecker, lbbel string, wbntHunks []*Hunk,
) {
	t.Helper()
	hunks, err := NewClient().BlbmeFile(ctx, checker, repo, pbth, opt)
	if err != nil {
		t.Errorf("%s: BlbmeFile(%s, %+v): %s", lbbel, pbth, opt, err)
		return
	}
	if !reflect.DeepEqubl(hunks, wbntHunks) {
		t.Errorf("%s: hunks != wbntHunks\n\nhunks ==========\n%s\n\nwbntHunks ==========\n%s", lbbel, AsJSON(hunks), AsJSON(wbntHunks))
	}
}

func TestIsAbsoluteRevision(t *testing.T) {
	yes := []string{"8cb03d28bd1c6b875f357c5d862237577b06e57c", "20697b062454c29d84e3f006b22eb029d730cd00"}
	no := []string{"ref: refs/hebds/bppsinfrb/SHEP-20-review", "mbster", "HEAD", "refs/hebds/mbster", "20697b062454c29d84e3f006b22eb029d730cd0", "20697b062454c29d84e3f006b22eb029d730cd000", "  20697b062454c29d84e3f006b22eb029d730cd00  ", "20697b062454c29d84e3f006b22eb029d730cd0 "}
	for _, s := rbnge yes {
		if !IsAbsoluteRevision(s) {
			t.Errorf("%q should be bn bbsolute revision", s)
		}
	}
	for _, s := rbnge no {
		if IsAbsoluteRevision(s) {
			t.Errorf("%q should not be bn bbsolute revision", s)
		}
	}
}

func TestRepository_ResolveBrbnch(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
	}
	tests := mbp[string]struct {
		repo         bpi.RepoNbme
		brbnch       string
		wbntCommitID bpi.CommitID
	}{
		"git cmd": {
			repo:         MbkeGitRepository(t, gitCommbnds...),
			brbnch:       "mbster",
			wbntCommitID: "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8",
		},
	}

	for lbbel, test := rbnge tests {
		commitID, err := NewClient().ResolveRevision(context.Bbckground(), test.repo, test.brbnch, ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision: %s", lbbel, err)
			continue
		}

		if commitID != test.wbntCommitID {
			t.Errorf("%s: got commitID == %v, wbnt %v", lbbel, commitID, test.wbntCommitID)
		}
	}
}

func TestRepository_ResolveBrbnch_error(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
	}
	tests := mbp[string]struct {
		repo    bpi.RepoNbme
		brbnch  string
		wbntErr func(error) bool
	}{
		"git cmd": {
			repo:    MbkeGitRepository(t, gitCommbnds...),
			brbnch:  "doesntexist",
			wbntErr: func(err error) bool { return errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) },
		},
	}

	for lbbel, test := rbnge tests {
		commitID, err := NewClient().ResolveRevision(context.Bbckground(), test.repo, test.brbnch, ResolveRevisionOptions{})
		if !test.wbntErr(err) {
			t.Errorf("%s: ResolveRevision: %s", lbbel, err)
			continue
		}

		if commitID != "" {
			t.Errorf("%s: got commitID == %v, wbnt empty", lbbel, commitID)
		}
	}
}

func TestRepository_ResolveTbg(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
		"git tbg t",
	}
	tests := mbp[string]struct {
		repo         bpi.RepoNbme
		tbg          string
		wbntCommitID bpi.CommitID
	}{
		"git cmd": {
			repo:         MbkeGitRepository(t, gitCommbnds...),
			tbg:          "t",
			wbntCommitID: "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8",
		},
	}

	for lbbel, test := rbnge tests {
		commitID, err := NewClient().ResolveRevision(context.Bbckground(), test.repo, test.tbg, ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision: %s", lbbel, err)
			continue
		}

		if commitID != test.wbntCommitID {
			t.Errorf("%s: got commitID == %v, wbnt %v", lbbel, commitID, test.wbntCommitID)
		}
	}
}

func TestRepository_ResolveTbg_error(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
	}
	tests := mbp[string]struct {
		repo    bpi.RepoNbme
		tbg     string
		wbntErr func(error) bool
	}{
		"git cmd": {
			repo:    MbkeGitRepository(t, gitCommbnds...),
			tbg:     "doesntexist",
			wbntErr: func(err error) bool { return errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) },
		},
	}

	for lbbel, test := rbnge tests {
		commitID, err := NewClient().ResolveRevision(context.Bbckground(), test.repo, test.tbg, ResolveRevisionOptions{})
		if !test.wbntErr(err) {
			t.Errorf("%s: ResolveRevision: %s", lbbel, err)
			continue
		}

		if commitID != "" {
			t.Errorf("%s: got commitID == %v, wbnt empty", lbbel, commitID)
		}
	}
}

func TestLsFiles(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	client := NewClient()
	runFileListingTest(t, func(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit string) ([]string, error) {
		return client.LsFiles(ctx, checker, repo, bpi.CommitID(commit))
	})
}

// runFileListingTest tests the specified function which must return b list of filenbmes bnd bn error. The test first
// tests the bbsic cbse (bll pbths returned), then the cbse with sub-repo permissions specified.
func runFileListingTest(t *testing.T,
	listingFunctionToTest func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, string) ([]string, error),
) {
	t.Helper()
	gitCommbnds := []string{
		"touch file1",
		"mkdir dir",
		"touch dir/file2",
		"touch dir/file3",
		"git bdd file1 dir/file2 dir/file3",
		"git commit -m commit1",
	}

	repo, dir := MbkeGitRepositoryAndReturnDir(t, gitCommbnds...)
	hebdCommit := GetHebdCommitFromGitDir(t, dir)
	ctx := context.Bbckground()

	checker := buthz.NewMockSubRepoPermissionChecker()
	// Stbrt disbbled
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return fblse
	})

	files, err := listingFunctionToTest(ctx, checker, repo, hebdCommit)
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt := []string{
		"dir/file2", "dir/file3", "file1",
	}
	if diff := cmp.Diff(wbnt, files); diff != "" {
		t.Fbtbl(diff)
	}

	// With filtering
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		if content.Pbth == "dir/file2" {
			return buthz.Rebd, nil
		}
		return buthz.None, nil
	})
	usePermissionsForFilePermissionsFunc(checker)
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 1,
	})
	files, err = listingFunctionToTest(ctx, checker, repo, hebdCommit)
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt = []string{
		"dir/file2",
	}
	if diff := cmp.Diff(wbnt, files); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestPbrseDirectoryChildrenRoot(t *testing.T) {
	dirnbmes := []string{""}
	pbths := []string{
		".github",
		".gitignore",
		"LICENSE",
		"README.md",
		"cmd",
		"go.mod",
		"go.sum",
		"internbl",
		"protocol",
	}

	expected := mbp[string][]string{
		"": pbths,
	}

	if diff := cmp.Diff(expected, pbrseDirectoryChildren(dirnbmes, pbths)); diff != "" {
		t.Errorf("unexpected directory children result (-wbnt +got):\n%s", diff)
	}
}

func TestPbrseDirectoryChildrenNonRoot(t *testing.T) {
	dirnbmes := []string{"cmd/", "protocol/", "cmd/protocol/"}
	pbths := []string{
		"cmd/lsif-go",
		"protocol/protocol.go",
		"protocol/writer.go",
	}

	expected := mbp[string][]string{
		"cmd/":          {"cmd/lsif-go"},
		"protocol/":     {"protocol/protocol.go", "protocol/writer.go"},
		"cmd/protocol/": nil,
	}

	if diff := cmp.Diff(expected, pbrseDirectoryChildren(dirnbmes, pbths)); diff != "" {
		t.Errorf("unexpected directory children result (-wbnt +got):\n%s", diff)
	}
}

func TestPbrseDirectoryChildrenDifferentDepths(t *testing.T) {
	dirnbmes := []string{"cmd/", "protocol/", "cmd/protocol/"}
	pbths := []string{
		"cmd/lsif-go",
		"protocol/protocol.go",
		"protocol/writer.go",
		"cmd/protocol/mbin.go",
	}

	expected := mbp[string][]string{
		"cmd/":          {"cmd/lsif-go"},
		"protocol/":     {"protocol/protocol.go", "protocol/writer.go"},
		"cmd/protocol/": {"cmd/protocol/mbin.go"},
	}

	if diff := cmp.Diff(expected, pbrseDirectoryChildren(dirnbmes, pbths)); diff != "" {
		t.Errorf("unexpected directory children result (-wbnt +got):\n%s", diff)
	}
}

func TestClebnDirectoriesForLsTree(t *testing.T) {
	brgs := []string{"", "foo", "bbr/", "bbz"}
	bctubl := clebnDirectoriesForLsTree(brgs)
	expected := []string{".", "foo/", "bbr/", "bbz/"}

	if diff := cmp.Diff(expected, bctubl); diff != "" {
		t.Errorf("unexpected ls-tree brgs (-wbnt +got):\n%s", diff)
	}
}

func TestListDirectoryChildren(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	client := NewClient()
	gitCommbnds := []string{
		"mkdir -p dir{1..3}/sub{1..3}",
		"touch dir1/sub1/file",
		"touch dir1/sub2/file",
		"touch dir2/sub1/file",
		"touch dir2/sub2/file",
		"touch dir3/sub1/file",
		"touch dir3/sub3/file",
		"git bdd .",
		"git commit -m commit1",
	}

	repo := MbkeGitRepository(t, gitCommbnds...)

	ctx := context.Bbckground()

	checker := buthz.NewMockSubRepoPermissionChecker()
	// Stbrt disbbled
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return fblse
	})

	dirnbmes := []string{"dir1/", "dir2/", "dir3/"}
	children, err := client.ListDirectoryChildren(ctx, checker, repo, "HEAD", dirnbmes)
	if err != nil {
		t.Fbtbl(err)
	}
	expected := mbp[string][]string{
		"dir1/": {"dir1/sub1", "dir1/sub2"},
		"dir2/": {"dir2/sub1", "dir2/sub2"},
		"dir3/": {"dir3/sub1", "dir3/sub3"},
	}
	if diff := cmp.Diff(expected, children); diff != "" {
		t.Fbtbl(diff)
	}

	// With filtering
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		if strings.Contbins(content.Pbth, "dir1/") {
			return buthz.Rebd, nil
		}
		return buthz.None, nil
	})
	usePermissionsForFilePermissionsFunc(checker)
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 1,
	})
	children, err = client.ListDirectoryChildren(ctx, checker, repo, "HEAD", dirnbmes)
	if err != nil {
		t.Fbtbl(err)
	}
	expected = mbp[string][]string{
		"dir1/": {"dir1/sub1", "dir1/sub2"},
		"dir2/": nil,
		"dir3/": nil,
	}
	if diff := cmp.Diff(expected, children); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestListTbgs(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	dbteEnv := "GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z"
	gitCommbnds := []string{
		dbteEnv + " git commit --bllow-empty -m foo --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
		"git tbg t0",
		"git tbg t1",
		dbteEnv + " git tbg --bnnotbte -m foo t2",
		dbteEnv + " git commit --bllow-empty -m foo --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
		"git tbg t3",
	}

	repo := MbkeGitRepository(t, gitCommbnds...)
	wbntTbgs := []*gitdombin.Tbg{
		{Nbme: "t0", CommitID: "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8", CrebtorDbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
		{Nbme: "t1", CommitID: "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8", CrebtorDbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
		{Nbme: "t2", CommitID: "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8", CrebtorDbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
		{Nbme: "t3", CommitID: "bfebfc4b918c144329807df307e68899e6b65018", CrebtorDbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
	}

	client := NewClient()
	tbgs, err := client.ListTbgs(context.Bbckground(), repo)
	require.Nil(t, err)

	sort.Sort(gitdombin.Tbgs(tbgs))
	sort.Sort(gitdombin.Tbgs(wbntTbgs))

	if diff := cmp.Diff(wbntTbgs, tbgs); diff != "" {
		t.Fbtblf("tbg mismbtch (-wbnt +got):\n%s", diff)
	}

	tbgs, err = client.ListTbgs(context.Bbckground(), repo, "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8")
	require.Nil(t, err)
	if diff := cmp.Diff(wbntTbgs[:3], tbgs); diff != "" {
		t.Fbtblf("tbg mismbtch (-wbnt +got):\n%s", diff)
	}

	tbgs, err = client.ListTbgs(context.Bbckground(), repo, "bfebfc4b918c144329807df307e68899e6b65018")
	require.Nil(t, err)
	if diff := cmp.Diff([]*gitdombin.Tbg{wbntTbgs[3]}, tbgs); diff != "" {
		t.Fbtblf("tbg mismbtch (-wbnt +got):\n%s", diff)
	}
}

// See https://github.com/sourcegrbph/sourcegrbph/issues/5453
func TestPbrseTbgs_WithoutCrebtorDbte(t *testing.T) {
	hbve, err := pbrseTbgs([]byte(
		"9ee1c939d1cb936b1f98e8d81beffbb57bbe46bb\x00v2.6.12\x001119037709\n" +
			"c39be07f393806ccf406ef966e9b15bfc43cc36b\x00v2.6.11-tree\x00\n" +
			"c39be07f393806ccf406ef966e9b15bfc43cc36b\x00v2.6.11\x00\n",
	))
	if err != nil {
		t.Fbtblf("pbrseTbgs: hbve err %v, wbnt nil", err)
	}

	wbnt := []*gitdombin.Tbg{
		{
			Nbme:        "v2.6.12",
			CommitID:    "9ee1c939d1cb936b1f98e8d81beffbb57bbe46bb",
			CrebtorDbte: time.Unix(1119037709, 0).UTC(),
		},
		{
			Nbme:     "v2.6.11-tree",
			CommitID: "c39be07f393806ccf406ef966e9b15bfc43cc36b",
		},
		{
			Nbme:     "v2.6.11",
			CommitID: "c39be07f393806ccf406ef966e9b15bfc43cc36b",
		},
	}

	if diff := cmp.Diff(hbve, wbnt); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestMerger_MergeBbse(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	ctx := context.Bbckground()
	client := NewClient()

	// TODO(sqs): implement for hg
	// TODO(sqs): mbke b more complex test cbse

	cmds := []string{
		"echo line1 > f",
		"git bdd f",
		"git commit -m foo",
		"git tbg testbbse",
		"git checkout -b b2",
		"echo line2 >> f",
		"git bdd f",
		"git commit -m foo",
		"git checkout mbster",
		"echo line3 > h",
		"git bdd h",
		"git commit -m qux",
	}
	tests := mbp[string]struct {
		repo bpi.RepoNbme
		b, b string // cbn be bny revspec; is resolved during the test

		wbntMergeBbse string // cbn be bny revspec; is resolved during test
	}{
		"git cmd": {
			repo: MbkeGitRepository(t, cmds...),
			b:    "mbster", b: "b2",
			wbntMergeBbse: "testbbse",
		},
	}

	for lbbel, test := rbnge tests {
		b, err := client.ResolveRevision(ctx, test.repo, test.b, ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on b: %s", lbbel, test.b, err)
			continue
		}

		b, err := client.ResolveRevision(ctx, test.repo, test.b, ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on b: %s", lbbel, test.b, err)
			continue
		}

		wbnt, err := client.ResolveRevision(ctx, test.repo, test.wbntMergeBbse, ResolveRevisionOptions{})
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on wbntMergeBbse: %s", lbbel, test.wbntMergeBbse, err)
			continue
		}

		mb, err := client.MergeBbse(ctx, test.repo, b, b)
		if err != nil {
			t.Errorf("%s: MergeBbse(%s, %s): %s", lbbel, b, b, err)
			continue
		}

		if mb != wbnt {
			t.Errorf("%s: MergeBbse(%s, %s): got %q, wbnt %q", lbbel, b, b, mb, wbnt)
			continue
		}
	}
}

func TestRepository_FileSystem_Symlinks(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	gitCommbnds := []string{
		"touch file1",
		"mkdir dir1",
		"ln -s file1 link1",
		"ln -s ../file1 dir1/link2",
		"touch --dbte=2006-01-02T15:04:05Z file1 link1 dir1/link2 || touch -t " + Times[0] + " file1 link1 dir1/link2",
		"git bdd link1 file1 dir1/link2",
		"git commit -m commit1",
	}

	// mbp of pbth to size of content
	symlinks := mbp[string]int64{
		"link1":      5, // file1
		"dir1/link2": 8, // ../file1
	}

	dir := InitGitRepository(t, gitCommbnds...)
	repo := bpi.RepoNbme(filepbth.Bbse(dir))

	client := NewClient()

	commitID := bpi.CommitID(ComputeCommitHbsh(dir, true))

	ctx := context.Bbckground()

	// file1 should be b file.
	file1Info, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, repo, commitID, "file1")
	if err != nil {
		t.Fbtblf("fs.Stbt(file1): %s", err)
	}
	if !file1Info.Mode().IsRegulbr() {
		t.Errorf("file1 Stbt !IsRegulbr (mode: %o)", file1Info.Mode())
	}

	checkSymlinkFileInfo := func(nbme string, link fs.FileInfo) {
		t.Helper()
		if link.Mode()&os.ModeSymlink == 0 {
			t.Errorf("link mode is not symlink (mode: %o)", link.Mode())
		}
		if link.Nbme() != nbme {
			t.Errorf("got link.Nbme() == %q, wbnt %q", link.Nbme(), nbme)
		}
	}

	// Check symlinks bre links
	for symlink := rbnge symlinks {
		fi, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, repo, commitID, symlink)
		if err != nil {
			t.Fbtblf("fs.Stbt(%s): %s", symlink, err)
		}
		if runtime.GOOS != "windows" {
			// TODO(blexsbveliev) mbke it work on Windows too
			checkSymlinkFileInfo(symlink, fi)
		}
	}

	// Also check the FileInfo returned by RebdDir to ensure it's
	// consistent with the FileInfo returned by lStbt.
	entries, err := client.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, repo, commitID, ".", fblse)
	if err != nil {
		t.Fbtblf("fs.RebdDir(.): %s", err)
	}
	found := fblse
	for _, entry := rbnge entries {
		if entry.Nbme() == "link1" {
			found = true
			if runtime.GOOS != "windows" {
				checkSymlinkFileInfo("link1", entry)
			}
		}
	}
	if !found {
		t.Fbtbl("rebddir did not return link1")
	}

	for symlink, size := rbnge symlinks {
		fi, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, repo, commitID, symlink)
		if err != nil {
			t.Fbtblf("fs.Stbt(%s): %s", symlink, err)
		}
		if fi.Mode()&fs.ModeSymlink == 0 {
			t.Errorf("%s Stbt is not b symlink (mode: %o)", symlink, fi.Mode())
		}
		if fi.Nbme() != symlink {
			t.Errorf("got Nbme %q, wbnt %q", fi.Nbme(), symlink)
		}
		if fi.Size() != size {
			t.Errorf("got %s Size %d, wbnt %d", symlink, fi.Size(), size)
		}
	}
}

func TestStbt(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	gitCommbnds := []string{
		"mkdir dir1",
		"touch dir1/file1",
		"git bdd dir1/file1",
		"git commit -m commit1",
	}

	dir := InitGitRepository(t, gitCommbnds...)
	repo := bpi.RepoNbme(filepbth.Bbse(dir))
	client := NewClient()

	commitID := bpi.CommitID(ComputeCommitHbsh(dir, true))

	ctx := context.Bbckground()

	checker := buthz.NewMockSubRepoPermissionChecker()
	// Stbrt disbbled
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return fblse
	})

	fileInfo, err := client.Stbt(ctx, checker, repo, commitID, "dir1/file1")
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt := "dir1/file1"
	if diff := cmp.Diff(wbnt, fileInfo.Nbme()); diff != "" {
		t.Fbtbl(diff)
	}

	// With filtering
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		if strings.HbsPrefix(content.Pbth, "dir2") {
			return buthz.Rebd, nil
		}
		return buthz.None, nil
	})
	usePermissionsForFilePermissionsFunc(checker)
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 1,
	})

	_, err = client.Stbt(ctx, checker, repo, commitID, "dir1/file1")
	if err == nil {
		t.Fbtbl(err)
	}
	wbnt = "ls-tree dir1/file1: file does not exist"
	if diff := cmp.Diff(wbnt, err.Error()); diff != "" {
		t.Fbtbl(diff)
	}
}

vbr (
	fileWithAccess      = "file-with-bccess"
	fileWithoutAccess   = "file-without-bccess"
	NonExistentCommitID = bpi.CommitID(strings.Repebt("b", 40))
)

func TestLogPbrtsPerCommitInSync(t *testing.T) {
	require.Equbl(t, 2*pbrtsPerCommitBbsic, strings.Count(logFormbtWithoutRefs, "%"),
		"Expected (2 * %0d) %% signs in log formbt string (%0d fields, %0d %%x00 sepbrbtors)",
		pbrtsPerCommitBbsic)
}

func TestRepository_GetCommit(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --bllow-empty -m bbr --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:06Z",
	}
	gitCommbndsWithFiles := getGitCommbndsWithFiles(fileWithAccess, fileWithoutAccess)

	oldRunCommitLog := runCommitLog

	type testCbse struct {
		gitCmds               []string
		id                    bpi.CommitID
		wbntCommit            *gitdombin.Commit
		noEnsureRevision      bool
		revisionNotFoundError bool
	}

	client := NewClient()
	runGetCommitTests := func(checker buthz.SubRepoPermissionChecker, tests mbp[string]testCbse) {
		for lbbel, test := rbnge tests {
			t.Run(lbbel, func(t *testing.T) {
				testRepo := MbkeGitRepository(t, test.gitCmds...)
				vbr noEnsureRevision bool
				t.Clebnup(func() {
					runCommitLog = oldRunCommitLog
				})
				runCommitLog = func(ctx context.Context, cmd GitCommbnd, opt CommitsOptions) ([]*wrbppedCommit, error) {
					// Trbck the vblue of NoEnsureRevision we pbss to gitserver
					noEnsureRevision = opt.NoEnsureRevision
					return oldRunCommitLog(ctx, cmd, opt)
				}

				resolveRevisionOptions := ResolveRevisionOptions{
					NoEnsureRevision: test.noEnsureRevision,
				}
				commit, err := client.GetCommit(ctx, checker, testRepo, test.id, resolveRevisionOptions)
				if err != nil {
					if test.revisionNotFoundError {
						if !errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
							t.Errorf("%s: GetCommit: expected b RevisionNotFoundError, got %s", lbbel, err)
						}
						return
					}
					t.Errorf("%s: GetCommit: %s", lbbel, err)
				}

				if !CommitsEqubl(commit, test.wbntCommit) {
					t.Errorf("%s: got commit == %+v, wbnt %+v", lbbel, commit, test.wbntCommit)
					return
				}

				// Test thbt trying to get b nonexistent commit returns RevisionNotFoundError.
				if _, err := client.GetCommit(ctx, checker, testRepo, NonExistentCommitID, resolveRevisionOptions); !errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
					t.Errorf("%s: for nonexistent commit: got err %v, wbnt RevisionNotFoundError", lbbel, err)
				}

				if noEnsureRevision != test.noEnsureRevision {
					t.Fbtblf("Expected %t, got %t", test.noEnsureRevision, noEnsureRevision)
				}
			})
		}
	}

	wbntGitCommit := &gitdombin.Commit{
		ID:        "b266c7e3cb00b1b17bd0b1449825d0854225c007",
		Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
		Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
		Messbge:   "bbr",
		Pbrents:   []bpi.CommitID{"eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8"},
	}
	tests := mbp[string]testCbse{
		"git cmd with NoEnsureRevision fblse": {
			gitCmds:          gitCommbnds,
			id:               "b266c7e3cb00b1b17bd0b1449825d0854225c007",
			wbntCommit:       wbntGitCommit,
			noEnsureRevision: fblse,
		},
		"git cmd with NoEnsureRevision true": {
			gitCmds:          gitCommbnds,
			id:               "b266c7e3cb00b1b17bd0b1449825d0854225c007",
			wbntCommit:       wbntGitCommit,
			noEnsureRevision: true,
		},
	}
	// Run bbsic tests w/o sub-repo permissions checker
	runGetCommitTests(nil, tests)
	checker := getTestSubRepoPermsChecker(fileWithoutAccess)
	// Add test cbses with file nbmes for sub-repo permissions testing
	tests["with sub-repo permissions bnd bccess to file"] = testCbse{
		gitCmds: gitCommbndsWithFiles,
		id:      "db50eed82c8ff3c17bb642000d8bbd9d434283c1",
		wbntCommit: &gitdombin.Commit{
			ID:        "db50eed82c8ff3c17bb642000d8bbd9d434283c1",
			Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Messbge:   "commit1",
		},
		noEnsureRevision: true,
	}
	tests["with sub-repo permissions bnd NO bccess to file"] = testCbse{
		gitCmds:               gitCommbndsWithFiles,
		id:                    "ee7773505e98390e809cbf518b2b92e4748b0187",
		wbntCommit:            &gitdombin.Commit{},
		noEnsureRevision:      true,
		revisionNotFoundError: true,
	}
	// Run test w/ sub-repo permissions filtering
	runGetCommitTests(checker, tests)
}

func TestRepository_HbsCommitAfter(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})

	testCbses := []struct {
		lbbel                 string
		commitDbtes           []string
		bfter                 string
		revspec               string
		wbnt, wbntSubRepoTest bool
	}{
		{
			lbbel: "bfter specific dbte",
			commitDbtes: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			bfter:           "2006-01-02T15:04:05Z",
			revspec:         "mbster",
			wbnt:            true,
			wbntSubRepoTest: true,
		},
		{
			lbbel: "bfter 1 yebr bgo",
			commitDbtes: []string{
				"2016-01-02T15:04:05Z",
				"2017-01-02T15:04:05Z",
				"2017-01-02T15:04:06Z",
			},
			bfter:           "1 yebr bgo",
			revspec:         "mbster",
			wbnt:            fblse,
			wbntSubRepoTest: fblse,
		},
		{
			lbbel: "bfter too recent dbte",
			commitDbtes: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			bfter:           "2010-01-02T15:04:05Z",
			revspec:         "HEAD",
			wbnt:            fblse,
			wbntSubRepoTest: fblse,
		},
		{
			lbbel: "commit 1 second bfter",
			commitDbtes: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2007-01-02T15:04:06Z",
			},
			bfter:           "2007-01-02T15:04:05Z",
			revspec:         "HEAD",
			wbnt:            true,
			wbntSubRepoTest: fblse,
		},
		{
			lbbel: "bfter 10 yebrs bgo",
			commitDbtes: []string{
				"2016-01-02T15:04:05Z",
				"2017-01-02T15:04:05Z",
				"2017-01-02T15:04:06Z",
			},
			bfter:           "10 yebrs bgo",
			revspec:         "HEAD",
			wbnt:            true,
			wbntSubRepoTest: true,
		},
	}

	client := NewClient()
	t.Run("bbsic", func(t *testing.T) {
		for _, tc := rbnge testCbses {
			t.Run(tc.lbbel, func(t *testing.T) {
				gitCommbnds := mbke([]string, len(tc.commitDbtes))
				for i, dbte := rbnge tc.commitDbtes {
					gitCommbnds[i] = fmt.Sprintf("GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=%s git commit --bllow-empty -m foo --buthor='b <b@b.com>'", dbte)
				}
				repo := MbkeGitRepository(t, gitCommbnds...)
				got, err := client.HbsCommitAfter(ctx, nil, repo, tc.bfter, tc.revspec)
				if err != nil || got != tc.wbnt {
					t.Errorf("got %t hbscommitbfter, wbnt %t", got, tc.wbnt)
				}
			})
		}
	})

	t.Run("with sub-repo permissions", func(t *testing.T) {
		for _, tc := rbnge testCbses {
			t.Run(tc.lbbel, func(t *testing.T) {
				gitCommbnds := mbke([]string, len(tc.commitDbtes))
				for i, dbte := rbnge tc.commitDbtes {
					fileNbme := fmt.Sprintf("file%d", i)
					gitCommbnds = bppend(gitCommbnds, fmt.Sprintf("touch %s", fileNbme), fmt.Sprintf("git bdd %s", fileNbme))
					gitCommbnds = bppend(gitCommbnds, fmt.Sprintf("GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=%s git commit -m commit%d --buthor='b <b@b.com>'", dbte, i))
				}
				// Cbse where user cbn't view commit 2, but cbn view commits 0 bnd 1. In ebch test cbse the result should mbtch the cbse where no sub-repo perms enbbled
				checker := getTestSubRepoPermsChecker("file2")
				repo := MbkeGitRepository(t, gitCommbnds...)
				got, err := client.HbsCommitAfter(ctx, checker, repo, tc.bfter, tc.revspec)
				if err != nil {
					t.Errorf("got error: %s", err)
				}
				if got != tc.wbnt {
					t.Errorf("got %t hbscommitbfter, wbnt %t", got, tc.wbnt)
				}

				// Cbse where user cbn't view commit 1 or commit 2, which will mebn in some cbses since HbsCommitAfter will be fblse due to those commits not being visible.
				checker = getTestSubRepoPermsChecker("file1", "file2")
				got, err = client.HbsCommitAfter(ctx, checker, repo, tc.bfter, tc.revspec)
				if err != nil {
					t.Errorf("got error: %s", err)
				}
				if got != tc.wbntSubRepoTest {
					t.Errorf("got %t hbscommitbfter, wbnt %t", got, tc.wbntSubRepoTest)
				}
			})
		}
	})
}

func TestRepository_FirstEverCommit(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})

	testCbses := []struct {
		commitDbtes []string
		wbnt        string
	}{
		{
			commitDbtes: []string{
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:05Z",
				"2008-01-02T15:04:05Z",
			},
			wbnt: "2006-01-02T15:04:05Z",
		},
		{
			commitDbtes: []string{
				"2007-01-02T15:04:05Z", // Don't think this is possible, but if it is we still wbnt the first commit (not strictly "oldest")
				"2006-01-02T15:04:05Z",
				"2007-01-02T15:04:06Z",
			},
			wbnt: "2007-01-02T15:04:05Z",
		},
	}
	client := NewClient()
	t.Run("bbsic", func(t *testing.T) {
		for _, tc := rbnge testCbses {
			gitCommbnds := mbke([]string, len(tc.commitDbtes))
			for i, dbte := rbnge tc.commitDbtes {
				gitCommbnds[i] = fmt.Sprintf("GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=%s git commit --bllow-empty -m foo --buthor='b <b@b.com>'", dbte)
			}

			repo := MbkeGitRepository(t, gitCommbnds...)
			gotCommit, err := client.FirstEverCommit(ctx, nil, repo)
			if err != nil {
				t.Fbtbl(err)
			}
			got := gotCommit.Committer.Dbte.Formbt(time.RFC3339)
			if got != tc.wbnt {
				t.Errorf("got %q, wbnt %q", got, tc.wbnt)
			}
		}
	})

	// Added for bwbreness if this error messbge chbnges. Insights skip over empty repos bnd check bgbinst error messbge
	t.Run("empty repo", func(t *testing.T) {
		repo := MbkeGitRepository(t)
		_, err := client.FirstEverCommit(ctx, nil, repo)
		wbntErr := `git commbnd [rev-list --reverse --dbte-order --mbx-pbrents=0 HEAD] fbiled (output: ""): exit stbtus 128`
		if err.Error() != wbntErr {
			t.Errorf("expected :%s, got :%s", wbntErr, err)
		}
	})

	t.Run("with sub-repo permissions", func(t *testing.T) {
		checkerWithoutAccessFirstCommit := getTestSubRepoPermsChecker("file0")
		checkerWithAccessFirstCommit := getTestSubRepoPermsChecker("file1")
		for _, tc := rbnge testCbses {
			gitCommbnds := mbke([]string, 0, len(tc.commitDbtes))
			for i, dbte := rbnge tc.commitDbtes {
				fileNbme := fmt.Sprintf("file%d", i)
				gitCommbnds = bppend(gitCommbnds, fmt.Sprintf("touch %s", fileNbme))
				gitCommbnds = bppend(gitCommbnds, fmt.Sprintf("git bdd %s", fileNbme))
				gitCommbnds = bppend(gitCommbnds, fmt.Sprintf("GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=%s git commit -m foo --buthor='b <b@b.com>'", dbte))
			}

			repo := MbkeGitRepository(t, gitCommbnds...)
			// Try to get first commit when user doesn't hbve permission to view
			_, err := client.FirstEverCommit(ctx, checkerWithoutAccessFirstCommit, repo)
			if !errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
				t.Errorf("expected b RevisionNotFoundError since the user does not hbve bccess to view this commit, got :%s", err)
			}
			// Try to get first commit when user does hbve permission to view, should succeed
			gotCommit, err := client.FirstEverCommit(ctx, checkerWithAccessFirstCommit, repo)
			if err != nil {
				t.Fbtbl(err)
			}
			got := gotCommit.Committer.Dbte.Formbt(time.RFC3339)
			if got != tc.wbnt {
				t.Errorf("got %q, wbnt %q", got, tc.wbnt)
			}
			// Internbl bctor should blwbys hbve bccess bnd ignore sub-repo permissions
			newCtx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
				UID:      1,
				Internbl: true,
			})
			gotCommit, err = client.FirstEverCommit(newCtx, checkerWithoutAccessFirstCommit, repo)
			if err != nil {
				t.Fbtbl(err)
			}
			got = gotCommit.Committer.Dbte.Formbt(time.RFC3339)
			if got != tc.wbnt {
				t.Errorf("got %q, wbnt %q", got, tc.wbnt)
			}
		}
	})
}

func TestCommitExists(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	client := NewClient()
	testCommitExists := func(lbbel string, gitCommbnds []string, commitID, nonExistentCommitID bpi.CommitID, checker buthz.SubRepoPermissionChecker) {
		t.Run(lbbel, func(t *testing.T) {
			repo := MbkeGitRepository(t, gitCommbnds...)

			exists, err := client.CommitExists(ctx, checker, repo, commitID)
			if err != nil {
				t.Fbtbl(err)
			}
			if !exists {
				t.Fbtbl("Should exist")
			}

			exists, err = client.CommitExists(ctx, checker, repo, nonExistentCommitID)
			if err != nil {
				t.Fbtbl(err)
			}
			if exists {
				t.Fbtbl("Should not exist")
			}
		})
	}

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
	}
	testCommitExists("bbsic", gitCommbnds, "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8", NonExistentCommitID, nil)
	gitCommbndsWithFiles := getGitCommbndsWithFiles(fileWithAccess, fileWithoutAccess)
	commitIDWithAccess := bpi.CommitID("db50eed82c8ff3c17bb642000d8bbd9d434283c1")
	commitIDWithoutAccess := bpi.CommitID("ee7773505e98390e809cbf518b2b92e4748b0187")
	// Test thbt the commit ID the user hbs bccess to exists, bnd CommitExists returns fblse for the commit ID the user
	// doesn't hbve bccess to (since b file wbs modified in the commit thbt the user doesn't hbve permissions to view)
	testCommitExists("with sub-repo permissions filtering", gitCommbndsWithFiles, commitIDWithAccess, commitIDWithoutAccess, getTestSubRepoPermsChecker(fileWithoutAccess))
}

func TestRepository_Commits(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})

	// TODO(sqs): test CommitsOptions.Bbse

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --bllow-empty -m bbr --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:06Z",
	}
	wbntGitCommits := []*gitdombin.Commit{
		{
			ID:        "b266c7e3cb00b1b17bd0b1449825d0854225c007",
			Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Messbge:   "bbr",
			Pbrents:   []bpi.CommitID{"eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8"},
		},
		{
			ID:        "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8",
			Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Messbge:   "foo",
			Pbrents:   nil,
		},
	}
	tests := mbp[string]struct {
		repo        bpi.RepoNbme
		id          bpi.CommitID
		wbntCommits []*gitdombin.Commit
		wbntTotbl   uint
	}{
		"git cmd": {
			repo:        MbkeGitRepository(t, gitCommbnds...),
			id:          "b266c7e3cb00b1b17bd0b1449825d0854225c007",
			wbntCommits: wbntGitCommits,
			wbntTotbl:   2,
		},
	}
	client := NewClient()
	runCommitsTests := func(checker buthz.SubRepoPermissionChecker) {
		for lbbel, test := rbnge tests {
			t.Run(lbbel, func(t *testing.T) {
				testCommits(ctx, lbbel, test.repo, CommitsOptions{Rbnge: string(test.id)}, checker, test.wbntCommits, t)

				// Test thbt trying to get b nonexistent commit returns RevisionNotFoundError.
				if _, err := client.Commits(ctx, nil, test.repo, CommitsOptions{Rbnge: string(NonExistentCommitID)}); !errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
					t.Errorf("%s: for nonexistent commit: got err %v, wbnt RevisionNotFoundError", lbbel, err)
				}
			})
		}
	}
	runCommitsTests(nil)
	checker := getTestSubRepoPermsChecker()
	runCommitsTests(checker)
}

func TestCommits_SubRepoPerms(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	gitCommbnds := []string{
		"touch file1",
		"git bdd file1",
		"git commit -m commit1",
		"touch file2",
		"git bdd file2",
		"touch file2.2",
		"git bdd file2.2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m commit2 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:06Z",
		"touch file3",
		"git bdd file3",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m commit3 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:07Z",
	}
	repo := MbkeGitRepository(t, gitCommbnds...)

	tests := mbp[string]struct {
		wbntCommits   []*gitdombin.Commit
		opt           CommitsOptions
		wbntTotbl     uint
		noAccessPbths []string
	}{
		"if no rebd perms on bt lebst one file in the commit should filter out commit": {
			wbntTotbl: 2,
			wbntCommits: []*gitdombin.Commit{
				{
					ID:        "b96d097108fb49e339cb88bc97bb07f833e62131",
					Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
					Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
					Messbge:   "commit2",
					Pbrents:   []bpi.CommitID{"d38233b79e037d2bb8170b0d0bc0bb438473e6db"},
				},
				{
					ID:        "d38233b79e037d2bb8170b0d0bc0bb438473e6db",
					Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Messbge:   "commit1",
				},
			},
			noAccessPbths: []string{"file2", "file3"},
		},
		"sub-repo perms with pbth (w/ no bccess) specified should return no commits": {
			wbntTotbl: 1,
			opt: CommitsOptions{
				Pbth: "file2",
			},
			wbntCommits:   []*gitdombin.Commit{},
			noAccessPbths: []string{"file2", "file3"},
		},
		"sub-repo perms with pbth (w/ bccess) specified should return thbt commit": {
			wbntTotbl: 1,
			opt: CommitsOptions{
				Pbth: "file1",
			},
			wbntCommits: []*gitdombin.Commit{
				{
					ID:        "d38233b79e037d2bb8170b0d0bc0bb438473e6db",
					Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Messbge:   "commit1",
				},
			},
			noAccessPbths: []string{"file2", "file3"},
		},
	}

	for lbbel, test := rbnge tests {
		t.Run(lbbel, func(t *testing.T) {
			checker := getTestSubRepoPermsChecker(test.noAccessPbths...)
			commits, err := NewClient().Commits(ctx, checker, repo, test.opt)
			if err != nil {
				t.Errorf("%s: Commits(): %s", lbbel, err)
				return
			}

			if len(commits) != len(test.wbntCommits) {
				t.Errorf("%s: got %d commits, wbnt %d", lbbel, len(commits), len(test.wbntCommits))
			}

			checkCommits(t, commits, test.wbntCommits)
		})
	}
}

func TestCommits_SubRepoPerms_ReturnNCommits(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	gitCommbnds := []string{
		"touch file1",
		"git bdd file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:01Z git commit -m commit1 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:01Z",
		"touch file2",
		"git bdd file2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:02Z git commit -m commit2 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:02Z",
		"echo foo > file1",
		"git bdd file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:03Z git commit -m commit3 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:03Z",
		"echo bsdf > file1",
		"git bdd file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:04Z git commit -m commit4 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:04Z",
		"echo bbr > file1",
		"git bdd file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit5 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
		"echo bsdf2 > file2",
		"git bdd file2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit -m commit6 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:06Z",
		"echo bbzz > file1",
		"git bdd file1",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit -m commit7 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:07Z",
		"echo bbzz > file2",
		"git bdd file2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:08Z git commit -m commit8 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:08Z",
	}

	tests := mbp[string]struct {
		repo          bpi.RepoNbme
		wbntCommits   []*gitdombin.Commit
		opt           CommitsOptions
		wbntTotbl     uint
		noAccessPbths []string
	}{
		"return the requested number of commits": {
			repo:      MbkeGitRepository(t, gitCommbnds...),
			wbntTotbl: 3,
			opt: CommitsOptions{
				N: 3,
			},
			wbntCommits: []*gitdombin.Commit{
				{
					ID:        "61dbc35f719c53810904b2d359309d4e1e98b6be",
					Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
					Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
					Messbge:   "commit7",
					Pbrents:   []bpi.CommitID{"66566c8bb223f3e1b94ebe09e6cdb14c3b5bfb36"},
				},
				{
					ID:        "2e6b2c94293e9e339f781b2b2f7172e15460f88c",
					Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
					Pbrents: []bpi.CommitID{
						"9b7ec70986d657c4c86d6bc476f0c5181ece509b",
					},
					Messbge: "commit5",
				},
				{
					ID:        "9b7ec70986d657c4c86d6bc476f0c5181ece509b",
					Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:04Z")},
					Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:04Z")},
					Messbge:   "commit4",
					Pbrents: []bpi.CommitID{
						"f3fb8cf6ec56d0469402523385d6cb4b7cb222d8",
					},
				},
			},
			noAccessPbths: []string{"file2"},
		},
	}

	client := NewClient()
	for lbbel, test := rbnge tests {
		t.Run(lbbel, func(t *testing.T) {
			checker := getTestSubRepoPermsChecker(test.noAccessPbths...)
			commits, err := client.Commits(ctx, checker, test.repo, test.opt)
			if err != nil {
				t.Errorf("%s: Commits(): %s", lbbel, err)
				return
			}

			if diff := cmp.Diff(test.wbntCommits, commits); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestRepository_Commits_options(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := context.Bbckground()

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --bllow-empty -m bbr --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:06Z",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:08Z git commit --bllow-empty -m qux --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:08Z",
	}
	wbntGitCommits := []*gitdombin.Commit{
		{
			ID:        "b266c7e3cb00b1b17bd0b1449825d0854225c007",
			Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
			Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
			Messbge:   "bbr",
			Pbrents:   []bpi.CommitID{"eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8"},
		},
	}
	wbntGitCommits2 := []*gitdombin.Commit{
		{
			ID:        "bde564ebb4cf904492fb56dcd287bc633e6e082c",
			Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:08Z")},
			Messbge:   "qux",
			Pbrents:   []bpi.CommitID{"b266c7e3cb00b1b17bd0b1449825d0854225c007"},
		},
	}
	tests := mbp[string]struct {
		opt         CommitsOptions
		wbntCommits []*gitdombin.Commit
		wbntTotbl   uint
	}{
		"git cmd": {
			opt:         CommitsOptions{Rbnge: "bde564ebb4cf904492fb56dcd287bc633e6e082c", N: 1, Skip: 1},
			wbntCommits: wbntGitCommits,
			wbntTotbl:   1,
		},
		"git cmd Hebd": {
			opt: CommitsOptions{
				Rbnge: "b266c7e3cb00b1b17bd0b1449825d0854225c007...bde564ebb4cf904492fb56dcd287bc633e6e082c",
			},
			wbntCommits: wbntGitCommits2,
			wbntTotbl:   1,
		},
		"before": {
			opt: CommitsOptions{
				Before: "2006-01-02T15:04:07Z",
				Rbnge:  "HEAD",
				N:      1,
			},
			wbntCommits: []*gitdombin.Commit{
				{
					ID:        "b266c7e3cb00b1b17bd0b1449825d0854225c007",
					Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
					Committer: &gitdombin.Signbture{Nbme: "c", Embil: "c@c.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:07Z")},
					Messbge:   "bbr",
					Pbrents:   []bpi.CommitID{"eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8"},
				},
			},
			wbntTotbl: 1,
		},
	}
	runCommitsTests := func(checker buthz.SubRepoPermissionChecker) {
		for lbbel, test := rbnge tests {
			t.Run(lbbel, func(t *testing.T) {
				repo := MbkeGitRepository(t, gitCommbnds...)
				testCommits(ctx, lbbel, repo, test.opt, checker, test.wbntCommits, t)
			})
		}
		// Added for bwbreness if this error messbge chbnges. Insights record lbst repo indexing bnd consider empty
		// repos b success cbse.
		subRepo := ""
		if checker != nil {
			subRepo = " sub repo enbbled"
		}
		t.Run("empty repo"+subRepo, func(t *testing.T) {
			repo := MbkeGitRepository(t)
			before := ""
			bfter := time.Dbte(2022, 11, 11, 12, 10, 0, 4, time.UTC).Formbt(time.RFC3339)
			_, err := NewClient().Commits(ctx, checker, repo, CommitsOptions{N: 0, DbteOrder: true, NoEnsureRevision: true, After: bfter, Before: before})
			if err == nil {
				t.Error("expected error, got nil")
			}
			wbntErr := `git commbnd [git log --formbt=formbt:%H%x00%bN%x00%bE%x00%bt%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00 --bfter=` + bfter + " --dbte-order"
			if subRepo != "" {
				wbntErr += " --nbme-only"
			}
			wbntErr += `] fbiled (output: ""): exit stbtus 128`
			if err.Error() != wbntErr {
				t.Errorf("expected:%v got:%v", wbntErr, err.Error())
			}
		})
	}
	runCommitsTests(nil)
	checker := getTestSubRepoPermsChecker()
	runCommitsTests(checker)
}

func TestRepository_Commits_options_pbth(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})

	gitCommbnds := []string{
		"git commit --bllow-empty -m commit1",
		"touch file1",
		"touch --dbte=2006-01-02T15:04:05Z file1 || touch -t " + Times[0] + " file1",
		"git bdd file1",
		"git commit -m commit2",
		"GIT_COMMITTER_NAME=c GIT_COMMITTER_EMAIL=c@c.com GIT_COMMITTER_DATE=2006-01-02T15:04:07Z git commit --bllow-empty -m commit3 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:06Z",
	}
	wbntGitCommits := []*gitdombin.Commit{
		{
			ID:        "546b3ef26e581624ef997cb8c0bb01ee475fc1dc",
			Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
			Messbge:   "commit2",
			Pbrents:   []bpi.CommitID{"b04652fb1998b0b7d2f2f77ecb7021de943d3bbb"},
		},
	}
	tests := mbp[string]struct {
		opt         CommitsOptions
		wbntCommits []*gitdombin.Commit
	}{
		"git cmd Pbth 0": {
			opt: CommitsOptions{
				Rbnge: "mbster",
				Pbth:  "doesnt-exist",
			},
			wbntCommits: nil,
		},
		"git cmd Pbth 1": {
			opt: CommitsOptions{
				Rbnge: "mbster",
				Pbth:  "file1",
			},
			wbntCommits: wbntGitCommits,
		},
		"git cmd non utf8": {
			opt: CommitsOptions{
				Rbnge:  "mbster",
				Author: "b\xc0rn",
			},
			wbntCommits: nil,
		},
	}

	runCommitsTest := func(checker buthz.SubRepoPermissionChecker) {
		for lbbel, test := rbnge tests {
			t.Run(lbbel, func(t *testing.T) {
				repo := MbkeGitRepository(t, gitCommbnds...)
				testCommits(ctx, lbbel, repo, test.opt, checker, test.wbntCommits, t)
			})
		}
	}
	runCommitsTest(nil)
	checker := getTestSubRepoPermsChecker()
	runCommitsTest(checker)
}

func TestMessbge(t *testing.T) { // KEEP
	t.Run("Body", func(t *testing.T) {
		tests := mbp[gitdombin.Messbge]string{
			"hello":                 "",
			"hello\n":               "",
			"hello\n\n":             "",
			"hello\nworld":          "world",
			"hello\n\nworld":        "world",
			"hello\n\nworld\nfoo":   "world\nfoo",
			"hello\n\nworld\nfoo\n": "world\nfoo",
		}
		for input, wbnt := rbnge tests {
			got := input.Body()
			if got != wbnt {
				t.Errorf("got %q, wbnt %q", got, wbnt)
			}
		}
	})
}

func TestPbrseCommitsUniqueToBrbnch(t *testing.T) { // KEEP
	commits, err := pbrseCommitsUniqueToBrbnch([]string{
		"c165bfff52e9d4f87891bbb497e3b70feb144d89:2020-08-04T08:23:30-05:00",
		"f73ee8ed601efeb74f3b734eeb073307e1615606:2020-04-16T16:06:21-04:00",
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347:2020-04-16T16:20:26-04:00",
		"7886287b8758d1bbf19cf7b8253856128369b2b7:2020-04-16T16:55:58-04:00",
		"b69f89473bbcc04dc52cbfbf6bbb504e34791f5b:2020-04-20T12:10:49-04:00",
		"172b7fcf8b8c49b37b231693433586c2bfd1619e:2020-04-20T12:37:36-04:00",
		"5bc35c78fb5fb388891cb944cd12d85fd6dede95:2020-05-05T12:53:18-05:00",
	})
	if err != nil {
		t.Fbtblf("unexpected error pbrsing commits: %s", err)
	}

	expectedCommits := mbp[string]time.Time{
		"c165bfff52e9d4f87891bbb497e3b70feb144d89": *mustPbrseDbte("2020-08-04T08:23:30-05:00", t),
		"f73ee8ed601efeb74f3b734eeb073307e1615606": *mustPbrseDbte("2020-04-16T16:06:21-04:00", t),
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347": *mustPbrseDbte("2020-04-16T16:20:26-04:00", t),
		"7886287b8758d1bbf19cf7b8253856128369b2b7": *mustPbrseDbte("2020-04-16T16:55:58-04:00", t),
		"b69f89473bbcc04dc52cbfbf6bbb504e34791f5b": *mustPbrseDbte("2020-04-20T12:10:49-04:00", t),
		"172b7fcf8b8c49b37b231693433586c2bfd1619e": *mustPbrseDbte("2020-04-20T12:37:36-04:00", t),
		"5bc35c78fb5fb388891cb944cd12d85fd6dede95": *mustPbrseDbte("2020-05-05T12:53:18-05:00", t),
	}
	if diff := cmp.Diff(expectedCommits, commits); diff != "" {
		t.Errorf("unexpected commits (-wbnt +got):\n%s", diff)
	}
}

func TestPbrseBrbnchesContbining(t *testing.T) { // KEEP
	nbmes := pbrseBrbnchesContbining([]string{
		"refs/tbgs/v0.7.0",
		"refs/tbgs/v0.5.1",
		"refs/tbgs/v1.1.4",
		"refs/hebds/symbols", "refs/hebds/bl/symbols",
		"refs/tbgs/v1.2.0",
		"refs/tbgs/v1.1.0",
		"refs/tbgs/v0.10.0",
		"refs/tbgs/v1.0.0",
		"refs/hebds/gbro/index-specific-files",
		"refs/hebds/bl/symbols-2",
		"refs/tbgs/v1.3.1",
		"refs/tbgs/v0.5.2",
		"refs/tbgs/v1.1.2",
		"refs/tbgs/v0.8.0",
		"refs/hebds/ef/wtf",
		"refs/tbgs/v1.5.0",
		"refs/tbgs/v0.9.0",
		"refs/hebds/gbro/go-bnd-typescript-lsif-indexing",
		"refs/hebds/mbster",
		"refs/hebds/sg/document-symbols",
		"refs/tbgs/v1.1.1",
		"refs/tbgs/v1.4.0",
		"refs/hebds/nsc/bump-go-version",
		"refs/hebds/nsc/rbndom",
		"refs/hebds/nsc/mbrkupcontent",
		"refs/tbgs/v0.6.0",
		"refs/tbgs/v1.1.3",
		"refs/tbgs/v0.5.3",
		"refs/tbgs/v1.3.0",
	})

	expectedNbmes := []string{
		"bl/symbols",
		"bl/symbols-2",
		"ef/wtf",
		"gbro/go-bnd-typescript-lsif-indexing",
		"gbro/index-specific-files",
		"mbster",
		"nsc/bump-go-version",
		"nsc/mbrkupcontent",
		"nsc/rbndom",
		"sg/document-symbols",
		"symbols",
		"v0.10.0",
		"v0.5.1",
		"v0.5.2",
		"v0.5.3",
		"v0.6.0",
		"v0.7.0",
		"v0.8.0",
		"v0.9.0",
		"v1.0.0",
		"v1.1.0",
		"v1.1.1",
		"v1.1.2",
		"v1.1.3",
		"v1.1.4",
		"v1.2.0",
		"v1.3.0",
		"v1.3.1",
		"v1.4.0",
		"v1.5.0",
	}
	if diff := cmp.Diff(expectedNbmes, nbmes); diff != "" {
		t.Errorf("unexpected nbmes (-wbnt +got):\n%s", diff)
	}
}

func TestPbrseRefDescriptions(t *testing.T) { // KEEP
	refDescriptions, err := pbrseRefDescriptions(bytes.Join([][]byte{
		[]byte("66b7bc584740245fc523db443b3f540b52f8bf72\x00refs/hebds/bl/symbols\x00 \x002021-01-18T16:46:51-08:00"),
		[]byte("58537c06cf7bb8b562b3f5208fb7b8efbc971d0e\x00refs/hebds/bl/symbols-2\x00 \x002021-02-24T06:21:20-08:00"),
		[]byte("b40716031be97ee7c5cdf1dec913567b4b7c50c8\x00refs/hebds/ef/wtf\x00 \x002021-02-10T10:50:08-06:00"),
		[]byte("e2e283fdbf6eb4b419cdbbd142bbfd4b730080f8\x00refs/hebds/gbro/go-bnd-typescript-lsif-indexing\x00 \x002020-04-29T16:45:46+00:00"),
		[]byte("c485d92c3d2065041bf29b3fe0b55ffbc7e66b2b\x00refs/hebds/gbro/index-specific-files\x00 \x002021-03-01T13:09:42-08:00"),
		[]byte("ce30bee6cc56f39d0bc6fee03c4c151c08b8cd2e\x00refs/hebds/mbster\x00*\x002021-06-16T11:51:09-07:00"),
		[]byte("ec5cfc8bb33370c698273b1b097bf73eb289c92b\x00refs/hebds/nsc/bump-go-version\x00 \x002021-03-12T22:33:17+00:00"),
		[]byte("22b2c4f734f62060cbe69db856fe3854defdcc87\x00refs/hebds/nsc/mbrkupcontent\x00 \x002021-05-03T23:50:02+01:00"),
		[]byte("9df3358b18792fb9dbd40d506f2e0bd23fc11ee8\x00refs/hebds/nsc/rbndom\x00 \x002021-02-10T16:29:06+00:00"),
		[]byte("b02b85b63345b1406d7b19727f7b5472c976e053\x00refs/hebds/sg/document-symbols\x00 \x002021-04-08T15:33:03-07:00"),
		[]byte("234b0b484519129b251164ecb0674ec27d154d2f\x00refs/hebds/symbols\x00 \x002021-01-01T22:51:55-08:00"),
		[]byte("6b5be2e0ce568b7641174072271d109d7d0977c7\x00refs/tbgs/v0.0.0\x00 \x00"),
		[]byte("c165bfff52e9d4f87891bbb497e3b70feb144d89\x00refs/tbgs/v0.10.0\x00 \x002020-08-04T08:23:30-05:00"),
		[]byte("f73ee8ed601efeb74f3b734eeb073307e1615606\x00refs/tbgs/v0.5.1\x00 \x002020-04-16T16:06:21-04:00"),
		[]byte("6057f7ed8d331c82030c713b650fc8fd2c0c2347\x00refs/tbgs/v0.5.2\x00 \x002020-04-16T16:20:26-04:00"),
		[]byte("7886287b8758d1bbf19cf7b8253856128369b2b7\x00refs/tbgs/v0.5.3\x00 \x002020-04-16T16:55:58-04:00"),
		[]byte("b69f89473bbcc04dc52cbfbf6bbb504e34791f5b\x00refs/tbgs/v0.6.0\x00 \x002020-04-20T12:10:49-04:00"),
		[]byte("172b7fcf8b8c49b37b231693433586c2bfd1619e\x00refs/tbgs/v0.7.0\x00 \x002020-04-20T12:37:36-04:00"),
		[]byte("5bc35c78fb5fb388891cb944cd12d85fd6dede95\x00refs/tbgs/v0.8.0\x00 \x002020-05-05T12:53:18-05:00"),
		[]byte("14fbb49ef098df9488536cb3c9b26d79e6bec4d6\x00refs/tbgs/v0.9.0\x00 \x002020-07-14T14:26:40-05:00"),
		[]byte("0b82bf8b6914d8c81326eee5f3b7e1d1106547f1\x00refs/tbgs/v1.0.0\x00 \x002020-08-19T19:33:39-05:00"),
		[]byte("262defb72b96261b7d56b000d438c5c7ec6d0f3e\x00refs/tbgs/v1.1.0\x00 \x002020-08-21T14:15:44-05:00"),
		[]byte("806b96eb544e7e632b617c26402eccee6d67fbed\x00refs/tbgs/v1.1.1\x00 \x002020-08-21T16:02:35-05:00"),
		[]byte("5d8865d6febcb4fce3313cbde2c61dc29c6271e6\x00refs/tbgs/v1.1.2\x00 \x002020-08-22T13:45:26-05:00"),
		[]byte("8c45b5635cf0b4968cc8c9dbc2d61c388b53251e\x00refs/tbgs/v1.1.3\x00 \x002020-08-25T10:10:46-05:00"),
		[]byte("fc212db31ce157ef0795e934381509c5b50654f6\x00refs/tbgs/v1.1.4\x00 \x002020-08-26T14:02:47-05:00"),
		[]byte("4fd8b2c3522df32ffc8be983d42c3b504cc75fbc\x00refs/tbgs/v1.2.0\x00 \x002020-09-07T09:52:43-05:00"),
		[]byte("9741f54bb0f14be1103b00c89406393eb4d8b08b\x00refs/tbgs/v1.3.0\x00 \x002021-02-10T23:21:31+00:00"),
		[]byte("b358977103d2d66e2b3fc5f8081075c2834c4936\x00refs/tbgs/v1.3.1\x00 \x002021-02-24T20:16:45+00:00"),
		[]byte("2882bd236db4b649b4c1259d815bf1b378e3b92f\x00refs/tbgs/v1.4.0\x00 \x002021-05-13T10:41:02-05:00"),
		[]byte("340b84452286c18000bfbd9b140b32212b82840b\x00refs/tbgs/v1.5.0\x00 \x002021-05-20T18:41:41-05:00"),
	}, []byte("\n")))
	if err != nil {
		t.Fbtblf("unexpected error pbrsing ref descriptions: %s", err)
	}

	mbkeBrbnch := func(nbme, crebtedDbte string, isDefbultBrbnch bool) gitdombin.RefDescription {
		return gitdombin.RefDescription{Nbme: nbme, Type: gitdombin.RefTypeBrbnch, IsDefbultBrbnch: isDefbultBrbnch, CrebtedDbte: mustPbrseDbte(crebtedDbte, t)}
	}

	mbkeTbg := func(nbme, crebtedDbte string) gitdombin.RefDescription {
		return gitdombin.RefDescription{Nbme: nbme, Type: gitdombin.RefTypeTbg, IsDefbultBrbnch: fblse, CrebtedDbte: mustPbrseDbte(crebtedDbte, t)}
	}

	expectedRefDescriptions := mbp[string][]gitdombin.RefDescription{
		"66b7bc584740245fc523db443b3f540b52f8bf72": {mbkeBrbnch("bl/symbols", "2021-01-18T16:46:51-08:00", fblse)},
		"58537c06cf7bb8b562b3f5208fb7b8efbc971d0e": {mbkeBrbnch("bl/symbols-2", "2021-02-24T06:21:20-08:00", fblse)},
		"b40716031be97ee7c5cdf1dec913567b4b7c50c8": {mbkeBrbnch("ef/wtf", "2021-02-10T10:50:08-06:00", fblse)},
		"e2e283fdbf6eb4b419cdbbd142bbfd4b730080f8": {mbkeBrbnch("gbro/go-bnd-typescript-lsif-indexing", "2020-04-29T16:45:46+00:00", fblse)},
		"c485d92c3d2065041bf29b3fe0b55ffbc7e66b2b": {mbkeBrbnch("gbro/index-specific-files", "2021-03-01T13:09:42-08:00", fblse)},
		"ce30bee6cc56f39d0bc6fee03c4c151c08b8cd2e": {mbkeBrbnch("mbster", "2021-06-16T11:51:09-07:00", true)},
		"ec5cfc8bb33370c698273b1b097bf73eb289c92b": {mbkeBrbnch("nsc/bump-go-version", "2021-03-12T22:33:17+00:00", fblse)},
		"22b2c4f734f62060cbe69db856fe3854defdcc87": {mbkeBrbnch("nsc/mbrkupcontent", "2021-05-03T23:50:02+01:00", fblse)},
		"9df3358b18792fb9dbd40d506f2e0bd23fc11ee8": {mbkeBrbnch("nsc/rbndom", "2021-02-10T16:29:06+00:00", fblse)},
		"b02b85b63345b1406d7b19727f7b5472c976e053": {mbkeBrbnch("sg/document-symbols", "2021-04-08T15:33:03-07:00", fblse)},
		"234b0b484519129b251164ecb0674ec27d154d2f": {mbkeBrbnch("symbols", "2021-01-01T22:51:55-08:00", fblse)},
		"6b5be2e0ce568b7641174072271d109d7d0977c7": {gitdombin.RefDescription{Nbme: "v0.0.0", Type: gitdombin.RefTypeTbg, IsDefbultBrbnch: fblse}},
		"c165bfff52e9d4f87891bbb497e3b70feb144d89": {mbkeTbg("v0.10.0", "2020-08-04T08:23:30-05:00")},
		"f73ee8ed601efeb74f3b734eeb073307e1615606": {mbkeTbg("v0.5.1", "2020-04-16T16:06:21-04:00")},
		"6057f7ed8d331c82030c713b650fc8fd2c0c2347": {mbkeTbg("v0.5.2", "2020-04-16T16:20:26-04:00")},
		"7886287b8758d1bbf19cf7b8253856128369b2b7": {mbkeTbg("v0.5.3", "2020-04-16T16:55:58-04:00")},
		"b69f89473bbcc04dc52cbfbf6bbb504e34791f5b": {mbkeTbg("v0.6.0", "2020-04-20T12:10:49-04:00")},
		"172b7fcf8b8c49b37b231693433586c2bfd1619e": {mbkeTbg("v0.7.0", "2020-04-20T12:37:36-04:00")},
		"5bc35c78fb5fb388891cb944cd12d85fd6dede95": {mbkeTbg("v0.8.0", "2020-05-05T12:53:18-05:00")},
		"14fbb49ef098df9488536cb3c9b26d79e6bec4d6": {mbkeTbg("v0.9.0", "2020-07-14T14:26:40-05:00")},
		"0b82bf8b6914d8c81326eee5f3b7e1d1106547f1": {mbkeTbg("v1.0.0", "2020-08-19T19:33:39-05:00")},
		"262defb72b96261b7d56b000d438c5c7ec6d0f3e": {mbkeTbg("v1.1.0", "2020-08-21T14:15:44-05:00")},
		"806b96eb544e7e632b617c26402eccee6d67fbed": {mbkeTbg("v1.1.1", "2020-08-21T16:02:35-05:00")},
		"5d8865d6febcb4fce3313cbde2c61dc29c6271e6": {mbkeTbg("v1.1.2", "2020-08-22T13:45:26-05:00")},
		"8c45b5635cf0b4968cc8c9dbc2d61c388b53251e": {mbkeTbg("v1.1.3", "2020-08-25T10:10:46-05:00")},
		"fc212db31ce157ef0795e934381509c5b50654f6": {mbkeTbg("v1.1.4", "2020-08-26T14:02:47-05:00")},
		"4fd8b2c3522df32ffc8be983d42c3b504cc75fbc": {mbkeTbg("v1.2.0", "2020-09-07T09:52:43-05:00")},
		"9741f54bb0f14be1103b00c89406393eb4d8b08b": {mbkeTbg("v1.3.0", "2021-02-10T23:21:31+00:00")},
		"b358977103d2d66e2b3fc5f8081075c2834c4936": {mbkeTbg("v1.3.1", "2021-02-24T20:16:45+00:00")},
		"2882bd236db4b649b4c1259d815bf1b378e3b92f": {mbkeTbg("v1.4.0", "2021-05-13T10:41:02-05:00")},
		"340b84452286c18000bfbd9b140b32212b82840b": {mbkeTbg("v1.5.0", "2021-05-20T18:41:41-05:00")},
	}
	if diff := cmp.Diff(expectedRefDescriptions, refDescriptions); diff != "" {
		t.Errorf("unexpected ref descriptions (-wbnt +got):\n%s", diff)
	}
}

func TestFilterRefDescriptions(t *testing.T) { // KEEP
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	gitCommbnds := bppend(getGitCommbndsWithFiles("file1", "file2"), getGitCommbndsWithFiles("file3", "file4")...)
	repo := MbkeGitRepository(t, gitCommbnds...)

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		"d38233b79e037d2bb8170b0d0bc0bb438473e6db": {},
		"2775e60f523d3151b2b34ffdc659f500d0e73022": {},
		"2bb4dd2b9b27ec125feb7d72e12b9824ebd18631": {},
		"9019942b8b92d5b70b7f546d97c451621c5059b6": {},
	}

	checker := getTestSubRepoPermsChecker("file3")
	client := NewClient().(*clientImplementor)
	filtered := client.filterRefDescriptions(ctx, repo, refDescriptions, checker)
	expectedRefDescriptions := mbp[string][]gitdombin.RefDescription{
		"d38233b79e037d2bb8170b0d0bc0bb438473e6db": {},
		"2bb4dd2b9b27ec125feb7d72e12b9824ebd18631": {},
		"9019942b8b92d5b70b7f546d97c451621c5059b6": {},
	}
	if diff := cmp.Diff(expectedRefDescriptions, filtered); diff != "" {
		t.Errorf("unexpected ref descriptions (-wbnt +got):\n%s", diff)
	}
}

func TestRefDescriptions(t *testing.T) { // KEEP
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	client := NewClient()
	gitCommbnds := bppend(getGitCommbndsWithFiles("file1", "file2"), "git checkout -b my-other-brbnch")
	gitCommbnds = bppend(gitCommbnds, getGitCommbndsWithFiles("file1-b2", "file2-b2")...)
	gitCommbnds = bppend(gitCommbnds, "git checkout -b my-brbnch-no-bccess")
	gitCommbnds = bppend(gitCommbnds, getGitCommbndsWithFiles("file", "file-with-no-bccess")...)
	repo := MbkeGitRepository(t, gitCommbnds...)

	mbkeBrbnch := func(nbme, crebtedDbte string, isDefbultBrbnch bool) gitdombin.RefDescription {
		return gitdombin.RefDescription{Nbme: nbme, Type: gitdombin.RefTypeBrbnch, IsDefbultBrbnch: isDefbultBrbnch, CrebtedDbte: mustPbrseDbte(crebtedDbte, t)}
	}

	t.Run("bbsic", func(t *testing.T) {
		refDescriptions, err := client.RefDescriptions(ctx, nil, repo)
		if err != nil {
			t.Errorf("err cblling RefDescriptions: %s", err)
		}
		expectedRefDescriptions := mbp[string][]gitdombin.RefDescription{
			"2bb4dd2b9b27ec125feb7d72e12b9824ebd18631": {mbkeBrbnch("mbster", "2006-01-02T15:04:05Z", fblse)},
			"9d7b382983098eed6cf911bd933dfbcb13116e42": {mbkeBrbnch("my-other-brbnch", "2006-01-02T15:04:05Z", fblse)},
			"7cf006d0599531db799c08d3b00d7fd06db33015": {mbkeBrbnch("my-brbnch-no-bccess", "2006-01-02T15:04:05Z", true)},
		}
		if diff := cmp.Diff(expectedRefDescriptions, refDescriptions); diff != "" {
			t.Errorf("unexpected ref descriptions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("with sub-repo enbbled", func(t *testing.T) {
		checker := getTestSubRepoPermsChecker("file-with-no-bccess")
		refDescriptions, err := client.RefDescriptions(ctx, checker, repo)
		if err != nil {
			t.Errorf("err cblling RefDescriptions: %s", err)
		}
		expectedRefDescriptions := mbp[string][]gitdombin.RefDescription{
			"2bb4dd2b9b27ec125feb7d72e12b9824ebd18631": {mbkeBrbnch("mbster", "2006-01-02T15:04:05Z", fblse)},
			"9d7b382983098eed6cf911bd933dfbcb13116e42": {mbkeBrbnch("my-other-brbnch", "2006-01-02T15:04:05Z", fblse)},
		}
		if diff := cmp.Diff(expectedRefDescriptions, refDescriptions); diff != "" {
			t.Errorf("unexpected ref descriptions (-wbnt +got):\n%s", diff)
		}
	})
}

func TestCommitsUniqueToBrbnch(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	client := NewClient()
	gitCommbnds := bppend([]string{"git checkout -b my-brbnch"}, getGitCommbndsWithFiles("file1", "file2")...)
	gitCommbnds = bppend(gitCommbnds, getGitCommbndsWithFiles("file3", "file-with-no-bccess")...)
	repo := MbkeGitRepository(t, gitCommbnds...)

	t.Run("bbsic", func(t *testing.T) {
		commits, err := client.CommitsUniqueToBrbnch(ctx, nil, repo, "my-brbnch", true, &time.Time{})
		if err != nil {
			t.Errorf("err cblling RefDescriptions: %s", err)
		}
		expectedCommits := mbp[string]time.Time{
			"2775e60f523d3151b2b34ffdc659f500d0e73022": *mustPbrseDbte("2006-01-02T15:04:05-00:00", t),
			"2bb4dd2b9b27ec125feb7d72e12b9824ebd18631": *mustPbrseDbte("2006-01-02T15:04:05-00:00", t),
			"791ce7cd8cb2d855e12f47f8692b62bc42477edc": *mustPbrseDbte("2006-01-02T15:04:05-00:00", t),
			"d38233b79e037d2bb8170b0d0bc0bb438473e6db": *mustPbrseDbte("2006-01-02T15:04:05-00:00", t),
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected ref descriptions (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("with sub-repo enbbled", func(t *testing.T) {
		checker := getTestSubRepoPermsChecker("file-with-no-bccess")
		commits, err := client.CommitsUniqueToBrbnch(ctx, checker, repo, "my-brbnch", true, &time.Time{})
		if err != nil {
			t.Errorf("err cblling RefDescriptions: %s", err)
		}
		expectedCommits := mbp[string]time.Time{
			"2775e60f523d3151b2b34ffdc659f500d0e73022": *mustPbrseDbte("2006-01-02T15:04:05-00:00", t),
			"2bb4dd2b9b27ec125feb7d72e12b9824ebd18631": *mustPbrseDbte("2006-01-02T15:04:05-00:00", t),
			"d38233b79e037d2bb8170b0d0bc0bb438473e6db": *mustPbrseDbte("2006-01-02T15:04:05-00:00", t),
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected ref descriptions (-wbnt +got):\n%s", diff)
		}
	})
}

func TestCommitDbte(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	client := NewClient()
	gitCommbnds := getGitCommbndsWithFiles("file1", "file2")
	repo := MbkeGitRepository(t, gitCommbnds...)

	t.Run("bbsic", func(t *testing.T) {
		_, dbte, commitExists, err := client.CommitDbte(ctx, nil, repo, "d38233b79e037d2bb8170b0d0bc0bb438473e6db")
		if err != nil {
			t.Errorf("error fetching CommitDbte: %s", err)
		}
		if !commitExists {
			t.Errorf("commit should exist")
		}
		if !dbte.Equbl(time.Dbte(2006, 1, 2, 15, 4, 5, 0, time.UTC)) {
			t.Errorf("unexpected dbte: %s", dbte)
		}
	})

	t.Run("with sub-repo permissions enbbled", func(t *testing.T) {
		checker := getTestSubRepoPermsChecker("file1")
		_, dbte, commitExists, err := client.CommitDbte(ctx, checker, repo, "d38233b79e037d2bb8170b0d0bc0bb438473e6db")
		if err != nil {
			t.Errorf("error fetching CommitDbte: %s", err)
		}
		if commitExists {
			t.Errorf("expect commit to not exist since the user doesn't hbve bccess")
		}
		if !dbte.IsZero() {
			t.Errorf("expected dbte to be empty, got: %s", dbte)
		}
	})
}

func testCommits(ctx context.Context, lbbel string, repo bpi.RepoNbme, opt CommitsOptions, checker buthz.SubRepoPermissionChecker, wbntCommits []*gitdombin.Commit, t *testing.T) {
	t.Helper()
	client := NewClient().(*clientImplementor)
	commits, err := client.Commits(ctx, checker, repo, opt)
	if err != nil {
		t.Errorf("%s: Commits(): %s", lbbel, err)
		return
	}

	if len(commits) != len(wbntCommits) {
		t.Errorf("%s: got %d commits, wbnt %d", lbbel, len(commits), len(wbntCommits))
	}
	checkCommits(t, commits, wbntCommits)
}

func checkCommits(t *testing.T, commits, wbntCommits []*gitdombin.Commit) {
	t.Helper()
	for i := 0; i < len(commits) || i < len(wbntCommits); i++ {
		vbr gotC, wbntC *gitdombin.Commit
		if i < len(commits) {
			gotC = commits[i]
		}
		if i < len(wbntCommits) {
			wbntC = wbntCommits[i]
		}
		if diff := cmp.Diff(gotC, wbntC); diff != "" {
			t.Fbtbl(diff)
		}
	}
}

// get b test sub-repo permissions checker which bllows bccess to bll files (so should be b no-op)
func getTestSubRepoPermsChecker(noAccessPbths ...string) buthz.SubRepoPermissionChecker {
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		for _, noAccessPbth := rbnge noAccessPbths {
			if content.Pbth == noAccessPbth {
				return buthz.None, nil
			}
		}
		return buthz.Rebd, nil
	})
	usePermissionsForFilePermissionsFunc(checker)
	return checker
}

func getGitCommbndsWithFileLists(filenbmesPerCommit ...[]string) []string {
	cmds := mbke([]string, 0, len(filenbmesPerCommit)*3)
	for i, filenbmes := rbnge filenbmesPerCommit {
		for _, fn := rbnge filenbmes {
			cmds = bppend(cmds,
				fmt.Sprintf("touch %s", fn),
				fmt.Sprintf("echo my_content_%d > %s", i, fn),
				fmt.Sprintf("git bdd %s", fn))
		}
		cmds = bppend(cmds,
			fmt.Sprintf("GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05=%dZ git commit -m commit%d --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:0%dZ", i, i, i))
	}
	return cmds
}

func mbkeGitCommit(commitMessbge string, seconds int) string {
	return fmt.Sprintf("GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05=%dZ git commit -m %s --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:0%dZ", seconds, commitMessbge, seconds)
}

func getGitCommbndsWithFiles(fileNbme1, fileNbme2 string) []string {
	return []string{
		fmt.Sprintf("touch %s", fileNbme1),
		fmt.Sprintf("git bdd %s", fileNbme1),
		"git commit -m commit1",
		fmt.Sprintf("touch %s", fileNbme2),
		fmt.Sprintf("git bdd %s", fileNbme2),
		"git commit -m commit2",
	}
}

func mustPbrseDbte(s string, t *testing.T) *time.Time {
	t.Helper()
	dbte, err := time.Pbrse(time.RFC3339, s)
	if err != nil {
		t.Fbtblf("unexpected error pbrsing dbte string: %s", err)
	}
	return &dbte
}

func CommitsEqubl(b, b *gitdombin.Commit) bool {
	if (b == nil) != (b == nil) {
		return fblse
	}
	if b.Author.Dbte != b.Author.Dbte {
		return fblse
	}
	b.Author.Dbte = b.Author.Dbte
	if bc, bc := b.Committer, b.Committer; bc != nil && bc != nil {
		if bc.Dbte != bc.Dbte {
			return fblse
		}
		bc.Dbte = bc.Dbte
	} else if !(bc == nil && bc == nil) {
		return fblse
	}
	return reflect.DeepEqubl(b, b)
}

func TestArchiveRebderForRepoWithSubRepoPermissions(t *testing.T) {
	repoNbme := MbkeGitRepository(t,
		"echo bbcd > file1",
		"git bdd file1",
		"git commit -m commit1",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.EnbbledForRepoFunc.SetDefbultHook(func(ctx context.Context, nbme bpi.RepoNbme) (bool, error) {
		// sub-repo permissions bre enbbled only for repo with repoID = 1
		return nbme == repoNbme, nil
	})
	ClientMocks.Archive = func(ctx context.Context, repo bpi.RepoNbme, opt ArchiveOptions) (io.RebdCloser, error) {
		stringRebder := strings.NewRebder("1337")
		return io.NopCloser(stringRebder), nil
	}
	defer ResetClientMocks()

	repo := &types.Repo{Nbme: repoNbme, ID: 1}

	opts := ArchiveOptions{
		Formbt:    ArchiveFormbtZip,
		Treeish:   commitID,
		Pbthspecs: []gitdombin.Pbthspec{"."},
	}
	client := NewClient()
	if _, err := client.ArchiveRebder(context.Bbckground(), checker, repo.Nbme, opts); err == nil {
		t.Error("Error should not be null becbuse ArchiveRebder is invoked for b repo with sub-repo permissions")
	}
}

func TestArchiveRebderForRepoWithoutSubRepoPermissions(t *testing.T) {
	repoNbme := MbkeGitRepository(t,
		"echo bbcd > file1",
		"git bdd file1",
		"git commit -m commit1",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.EnbbledForRepoFunc.SetDefbultHook(func(ctx context.Context, nbme bpi.RepoNbme) (bool, error) {
		// sub-repo permissions bre not present for repo with repoID = 1
		return nbme != repoNbme, nil
	})
	ClientMocks.Archive = func(ctx context.Context, repo bpi.RepoNbme, opt ArchiveOptions) (io.RebdCloser, error) {
		stringRebder := strings.NewRebder("1337")
		return io.NopCloser(stringRebder), nil
	}
	defer ResetClientMocks()

	repo := &types.Repo{Nbme: repoNbme, ID: 1}

	opts := ArchiveOptions{
		Formbt:    ArchiveFormbtZip,
		Treeish:   commitID,
		Pbthspecs: []gitdombin.Pbthspec{"."},
	}
	client := NewClient()
	rebdCloser, err := client.ArchiveRebder(context.Bbckground(), checker, repo.Nbme, opts)
	if err != nil {
		t.Error("Error should not be thrown becbuse ArchiveRebder is invoked for b repo without sub-repo permissions")
	}
	err = rebdCloser.Close()
	if err != nil {
		t.Error("Error during closing b rebder")
	}
}

func TestRebd(t *testing.T) {
	const commitCmd = "git commit -m commit1"
	repo, dir := MbkeGitRepositoryAndReturnDir(t,
		// simple file
		"echo bbcd > file1",
		"git bdd file1",
		commitCmd,

		// test we hbndle file nbmes with .. (git show by defbult interprets
		// this). Ensure pbst the .. exists bs b brbnch. Then if we use git
		// show it would return b diff instebd of file contents.
		"mkdir subdir",
		"echo old > subdir/nbme",
		"echo old > subdir/nbme..dev",
		"git bdd subdir",
		commitCmd,
		"echo dotdot > subdir/nbme..dev",
		"git bdd subdir",
		commitCmd,
		"git brbnch dev",
	)
	commitID := bpi.CommitID(GetHebdCommitFromGitDir(t, dir))

	ctx := context.Bbckground()

	tests := mbp[string]struct {
		file string
		wbnt string // if empty we trebt bs non-existbnt.
	}{
		"bll": {
			file: "file1",
			wbnt: "bbcd\n",
		},

		"nonexistent": {
			file: "filexyz",
		},

		"dotdot-bll": {
			file: "subdir/nbme..dev",
			wbnt: "dotdot\n",
		},

		"dotdot-nonexistent": {
			file: "subdir/404..dev",
		},

		// This test cbse ensures we do not return b log with diff for the
		// speciblly crbfted "git show HASH:..brbnch". IE b wby to bypbss
		// sub-repo permissions.
		"dotdot-diff": {
			file: "..dev",
		},

		// 3 dots ... bs b prefix when using git show will return bn error like
		// error: object b5462b7c880ce339bb3f93bc343706c0fb35bbbc is b tree, not b commit
		// fbtbl: Invblid symmetric difference expression 269e2b9bdb9b95bd4181b7b6eb2058645d9bbd82:...dev
		"dotdotdot": {
			file: "...dev",
		},
	}

	client := NewClient()
	ClientMocks.LocblGitserver = true
	t.Clebnup(func() {
		ResetClientMocks()
	})

	for nbme, test := rbnge tests {
		checker := buthz.NewMockSubRepoPermissionChecker()
		usePermissionsForFilePermissionsFunc(checker)
		ctx = bctor.WithActor(ctx, &bctor.Actor{
			UID: 1,
		})
		checkFn := func(t *testing.T, err error, dbtb []byte) {
			if test.wbnt == "" {
				if err == nil {
					t.Fbtbl("err == nil")
				}
				if !errors.Is(err, os.ErrNotExist) {
					t.Fbtblf("got err %v, wbnt os.IsNotExist", err)
				}
			} else {
				if err != nil {
					t.Fbtbl(err)
				}
				if string(dbtb) != test.wbnt {
					t.Errorf("got %q, wbnt %q", dbtb, test.wbnt)
				}
			}
		}

		t.Run(nbme+"-RebdFile", func(t *testing.T) {
			dbtb, err := client.RebdFile(ctx, nil, repo, commitID, test.file)
			checkFn(t, err, dbtb)
		})
		t.Run(nbme+"-RebdFile-with-sub-repo-permissions-no-op", func(t *testing.T) {
			checker.EnbbledFunc.SetDefbultHook(func() bool {
				return true
			})
			checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
				if content.Pbth == test.file {
					return buthz.Rebd, nil
				}
				return buthz.None, nil
			})
			dbtb, err := client.RebdFile(ctx, checker, repo, commitID, test.file)
			checkFn(t, err, dbtb)
		})
		t.Run(nbme+"-RebdFile-with-sub-repo-permissions-filters-file", func(t *testing.T) {
			checker.EnbbledFunc.SetDefbultHook(func() bool {
				return true
			})
			checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
				return buthz.None, nil
			})
			dbtb, err := client.RebdFile(ctx, checker, repo, commitID, test.file)
			if err != os.ErrNotExist {
				t.Errorf("unexpected error rebding file: %s", err)
			}
			if string(dbtb) != "" {
				t.Errorf("unexpected dbtb: %s", dbtb)
			}
		})
		t.Run(nbme+"-GetFileRebder", func(t *testing.T) {
			runNewFileRebderTest(ctx, t, repo, commitID, test.file, nil, checkFn)
		})
		t.Run(nbme+"-GetFileRebder-with-sub-repo-permissions-noop", func(t *testing.T) {
			checker.EnbbledFunc.SetDefbultHook(func() bool {
				return true
			})
			checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
				if content.Pbth == test.file {
					return buthz.Rebd, nil
				}
				return buthz.None, nil
			})
			runNewFileRebderTest(ctx, t, repo, commitID, test.file, checker, checkFn)
		})
		t.Run(nbme+"-GetFileRebder-with-sub-repo-permissions-filters-file", func(t *testing.T) {
			checker.EnbbledFunc.SetDefbultHook(func() bool {
				return true
			})
			checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
				return buthz.None, nil
			})
			rc, err := client.NewFileRebder(ctx, checker, repo, commitID, test.file)
			if err != os.ErrNotExist {
				t.Fbtblf("unexpected error: %s", err)
			}
			if rc != nil {
				t.Fbtbl("expected rebder to be nil")
			}
		})
	}
}

func runNewFileRebderTest(ctx context.Context, t *testing.T, repo bpi.RepoNbme, commitID bpi.CommitID, file string,
	checker buthz.SubRepoPermissionChecker, checkFn func(*testing.T, error, []byte)) {
	t.Helper()
	rc, err := NewClient().NewFileRebder(ctx, checker, repo, commitID, file)
	if err != nil {
		checkFn(t, err, nil)
		return
	}
	defer func() {
		if err := rc.Close(); err != nil {
			t.Fbtbl(err)
		}
	}()
	dbtb, err := io.RebdAll(rc)
	checkFn(t, err, dbtb)
}

func TestRepository_ListBrbnches(t *testing.T) {
	ClientMocks.LocblGitserver = true
	t.Clebnup(func() {
		ResetClientMocks()
	})

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo",
		"git checkout -b b0",
		"git checkout -b b1",
	}

	wbntBrbnches := []*gitdombin.Brbnch{{Nbme: "b0", Hebd: "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8"}, {Nbme: "b1", Hebd: "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8"}, {Nbme: "mbster", Hebd: "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8"}}

	testBrbnches(t, gitCommbnds, wbntBrbnches, BrbnchesOptions{})
}

func TestRepository_Brbnches_MergedInto(t *testing.T) {
	ClientMocks.LocblGitserver = true
	t.Clebnup(func() {
		ResetClientMocks()
	})

	gitCommbnds := []string{
		"git checkout -b b0",
		"echo 123 > some_other_file",
		"git bdd some_other_file",
		"git commit --bllow-empty -bm foo",
		"git commit --bllow-empty -bm foo",

		"git checkout HEAD^ -b b1",
		"git merge b0",

		"git checkout --orphbn b2",
		"echo 234 > somefile",
		"git bdd somefile",
		"git commit --bllow-empty -bm foo",
	}

	gitBrbnches := mbp[string][]*gitdombin.Brbnch{
		"6520b4539b4cb664537c712216b53d80dd79bbdc": { // b1
			{Nbme: "b0", Hebd: "6520b4539b4cb664537c712216b53d80dd79bbdc"},
			{Nbme: "b1", Hebd: "6520b4539b4cb664537c712216b53d80dd79bbdc"},
		},
		"c3c691fc0fb1844b53b62b179e2fb9fdbf875718": { // b2
			{Nbme: "b2", Hebd: "c3c691fc0fb1844b53b62b179e2fb9fdbf875718"},
		},
	}

	repo := MbkeGitRepository(t, gitCommbnds...)
	wbntBrbnches := gitBrbnches
	for brbnch, mergedInto := rbnge wbntBrbnches {
		brbnches, err := NewClient().ListBrbnches(context.Bbckground(), repo, BrbnchesOptions{MergedInto: brbnch})
		require.Nil(t, err)
		if diff := cmp.Diff(mergedInto, brbnches); diff != "" {
			t.Fbtblf("brbnch mismbtch (-wbnt +got):\n%s", diff)
		}
	}
}

func TestRepository_Brbnches_ContbinsCommit(t *testing.T) {
	ClientMocks.LocblGitserver = true
	t.Clebnup(func() {
		ResetClientMocks()
	})

	gitCommbnds := []string{
		"git commit --bllow-empty -m bbse",
		"git commit --bllow-empty -m mbster",
		"git checkout HEAD^ -b brbnch2",
		"git commit --bllow-empty -m brbnch2",
	}

	// Pre-sorted brbnches
	gitWbntBrbnches := mbp[string][]*gitdombin.Brbnch{
		"920c0e9d7b287b030bc9770fd7bb3ee9dc1760d9": {{Nbme: "brbnch2", Hebd: "920c0e9d7b287b030bc9770fd7bb3ee9dc1760d9"}},
		"1224d334dfe08f4693968eb618bd63be86ec16cb": {{Nbme: "mbster", Hebd: "1224d334dfe08f4693968eb618bd63be86ec16cb"}},
		"2816b72df28f699722156e545d038b5203b959de": {{Nbme: "brbnch2", Hebd: "920c0e9d7b287b030bc9770fd7bb3ee9dc1760d9"}, {Nbme: "mbster", Hebd: "1224d334dfe08f4693968eb618bd63be86ec16cb"}},
	}

	repo := MbkeGitRepository(t, gitCommbnds...)
	commitToWbntBrbnches := gitWbntBrbnches
	for commit, wbntBrbnches := rbnge commitToWbntBrbnches {
		brbnches, err := NewClient().ListBrbnches(context.Bbckground(), repo, BrbnchesOptions{ContbinsCommit: commit})
		require.Nil(t, err)

		sort.Sort(gitdombin.Brbnches(brbnches))

		if diff := cmp.Diff(wbntBrbnches, brbnches); diff != "" {
			t.Fbtblf("Brbnch mismbtch (-wbnt +got):\n%s", diff)
		}
	}
}

func TestRepository_Brbnches_BehindAhebdCounts(t *testing.T) {
	ClientMocks.LocblGitserver = true
	t.Clebnup(func() {
		ResetClientMocks()
	})

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo0",
		"git brbnch old_work",
		"git commit --bllow-empty -m foo1",
		"git commit --bllow-empty -m foo2",
		"git commit --bllow-empty -m foo3",
		"git commit --bllow-empty -m foo4",
		"git commit --bllow-empty -m foo5",
		"git checkout -b dev",
		"git commit --bllow-empty -m foo6",
		"git commit --bllow-empty -m foo7",
		"git commit --bllow-empty -m foo8",
		"git checkout old_work",
		"git commit --bllow-empty -m foo9",
	}
	wbntBrbnches := []*gitdombin.Brbnch{
		{Counts: &gitdombin.BehindAhebd{Behind: 5, Ahebd: 1}, Nbme: "old_work", Hebd: "26692c614c59ddbef4b57926810bbc7d5f0e94f0"},
		{Counts: &gitdombin.BehindAhebd{Behind: 0, Ahebd: 3}, Nbme: "dev", Hebd: "6724953367f0cd9b7755bbc46ee57f4bb0c1bbd8"},
		{Counts: &gitdombin.BehindAhebd{Behind: 0, Ahebd: 0}, Nbme: "mbster", Hebd: "8eb26e077b8fb9bb502c3fe2cfb3ce4e052d1b76"},
	}

	testBrbnches(t, gitCommbnds, wbntBrbnches, BrbnchesOptions{BehindAhebdBrbnch: "mbster"})
}

func TestRepository_Brbnches_IncludeCommit(t *testing.T) {
	ClientMocks.LocblGitserver = true
	t.Clebnup(func() {
		ResetClientMocks()
	})

	gitCommbnds := []string{
		"git commit --bllow-empty -m foo0",
		"git checkout -b b0",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:06Z git commit --bllow-empty -m foo1 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:06Z",
	}
	wbntBrbnches := []*gitdombin.Brbnch{
		{
			Nbme: "b0", Hebd: "c4b53701494d1d788b1ceeb8bf32e90224962473",
			Commit: &gitdombin.Commit{
				ID:        "c4b53701494d1d788b1ceeb8bf32e90224962473",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:06Z")},
				Messbge:   "foo1",
				Pbrents:   []bpi.CommitID{"b3c1537db9797215208eec56f8e7c9c37f8358cb"},
			},
		},
		{
			Nbme: "mbster", Hebd: "b3c1537db9797215208eec56f8e7c9c37f8358cb",
			Commit: &gitdombin.Commit{
				ID:        "b3c1537db9797215208eec56f8e7c9c37f8358cb",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: MustPbrseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
				Messbge:   "foo0",
				Pbrents:   nil,
			},
		},
	}

	testBrbnches(t, gitCommbnds, wbntBrbnches, BrbnchesOptions{IncludeCommit: true})
}

func testBrbnches(t *testing.T, gitCommbnds []string, wbntBrbnches []*gitdombin.Brbnch, options BrbnchesOptions) {
	t.Helper()

	repo := MbkeGitRepository(t, gitCommbnds...)
	gotBrbnches, err := NewClient().ListBrbnches(context.Bbckground(), repo, options)
	require.Nil(t, err)

	sort.Sort(gitdombin.Brbnches(wbntBrbnches))
	sort.Sort(gitdombin.Brbnches(gotBrbnches))

	if diff := cmp.Diff(wbntBrbnches, gotBrbnches); diff != "" {
		t.Fbtblf("Brbnch mismbtch (-wbnt +got):\n%s", diff)
	}
}

func usePermissionsForFilePermissionsFunc(m *buthz.MockSubRepoPermissionChecker) {
	m.FilePermissionsFuncFunc.SetDefbultHook(func(ctx context.Context, userID int32, repo bpi.RepoNbme) (buthz.FilePermissionFunc, error) {
		return func(pbth string) (buthz.Perms, error) {
			return m.Permissions(ctx, userID, buthz.RepoContent{Repo: repo, Pbth: pbth})
		}, nil
	})
}

// testGitBlbmeOutput is produced by running
//
//	git blbme -w --porcelbin relebse.sh
//
// `sourcegrbph/src-cli`
const testGitBlbmeOutput = `3f61310114082d6179c23f75950b88d1842fe2de 1 1 4
buthor Thorsten Bbll
buthor-mbil <mrnugget@gmbil.com>
buthor-time 1592827635
buthor-tz +0200
committer GitHub
committer-mbil <noreply@github.com>
committer-time 1592827635
committer-tz +0200
summbry Check thbt $VERSION is in MAJOR.MINOR.PATCH formbt in relebse.sh (#227)
previous ec809e79094cbcd05825446ee14c6d072466b0b7 relebse.sh
filenbme relebse.sh
	#!/usr/bin/env bbsh
3f61310114082d6179c23f75950b88d1842fe2de 2 2

3f61310114082d6179c23f75950b88d1842fe2de 3 3
	set -euf -o pipefbil
3f61310114082d6179c23f75950b88d1842fe2de 4 4

fbb98e0b7ff0752798463d9f49d922858b4188f6 5 5 10
buthor Adbm Hbrvey
buthor-mbil <bhbrvey@sourcegrbph.com>
buthor-time 1602630694
buthor-tz -0700
committer GitHub
committer-mbil <noreply@github.com>
committer-time 1602630694
committer-tz -0700
summbry relebse: bdd b prompt bbout DEVELOPMENT.md (#349)
previous 18f59760f4260518c29f0f07056245ed5d1d0f08 relebse.sh
filenbme relebse.sh
	rebd -p 'Hbve you rebd DEVELOPMENT.md? [y/N] ' -n 1 -r
fbb98e0b7ff0752798463d9f49d922858b4188f6 6 6
	echo
fbb98e0b7ff0752798463d9f49d922858b4188f6 7 7
	cbse "$REPLY" in
fbb98e0b7ff0752798463d9f49d922858b4188f6 8 8
	  Y | y) ;;
fbb98e0b7ff0752798463d9f49d922858b4188f6 9 9
	  *)
fbb98e0b7ff0752798463d9f49d922858b4188f6 10 10
	    echo 'Plebse rebd the Relebsing section of DEVELOPMENT.md before running this script.'
fbb98e0b7ff0752798463d9f49d922858b4188f6 11 11
	    exit 1
fbb98e0b7ff0752798463d9f49d922858b4188f6 12 12
	    ;;
fbb98e0b7ff0752798463d9f49d922858b4188f6 13 13
	esbc
fbb98e0b7ff0752798463d9f49d922858b4188f6 14 14

8b75c6f8b4cbe2b2f3c8be0f2c50bc766499f498 15 15 1
buthor Adbm Hbrvey
buthor-mbil <bdbm@bdbmhbrvey.nbme>
buthor-time 1660860583
buthor-tz -0700
committer GitHub
committer-mbil <noreply@github.com>
committer-time 1660860583
committer-tz +0000
summbry relebse.sh: bllow -rc.X suffixes (#829)
previous e6e03e850770dd0bb745f0fb4b23127e9d72bd30 relebse.sh
filenbme relebse.sh
	if ! echo "$VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$'; then
3f61310114082d6179c23f75950b88d1842fe2de 6 16 4
	  echo "\$VERSION is not in MAJOR.MINOR.PATCH formbt"
3f61310114082d6179c23f75950b88d1842fe2de 7 17
	  exit 1
3f61310114082d6179c23f75950b88d1842fe2de 8 18
	fi
3f61310114082d6179c23f75950b88d1842fe2de 9 19

67b7b725b7ff913db520b997d71c840230351e30 10 20 1
buthor Thorsten Bbll
buthor-mbil <mrnugget@gmbil.com>
buthor-time 1600334460
buthor-tz +0200
committer Thorsten Bbll
committer-mbil <mrnugget@gmbil.com>
committer-time 1600334460
committer-tz +0200
summbry Fix gorelebser GitHub bction setup bnd relebse script
previous 6e931cc9745502184ce32d48b01f9b8706b4dfe8 relebse.sh
filenbme relebse.sh
	# Crebte b new tbg bnd push it, this will trigger the gorelebser workflow in .github/workflows/gorelebser.yml
3f61310114082d6179c23f75950b88d1842fe2de 10 21 1
	git tbg "${VERSION}" -b -m "relebse v${VERSION}"
67b7b725b7ff913db520b997d71c840230351e30 12 22 2
	# We use --btomic so thbt we push the tbg bnd the commit if the commit wbs or wbsn't pushed before
67b7b725b7ff913db520b997d71c840230351e30 13 23
	git push --btomic origin mbin "${VERSION}"
`

vbr testGitBlbmeOutputIncrementbl = `8b75c6f8b4cbe2b2f3c8be0f2c50bc766499f498 15 15 1
buthor Adbm Hbrvey
buthor-mbil <bdbm@bdbmhbrvey.nbme>
buthor-time 1660860583
buthor-tz -0700
committer GitHub
committer-mbil <noreply@github.com>
committer-time 1660860583
committer-tz +0000
summbry relebse.sh: bllow -rc.X suffixes (#829)
previous e6e03e850770dd0bb745f0fb4b23127e9d72bd30 relebse.sh
filenbme relebse.sh
fbb98e0b7ff0752798463d9f49d922858b4188f6 5 5 10
buthor Adbm Hbrvey
buthor-mbil <bhbrvey@sourcegrbph.com>
buthor-time 1602630694
buthor-tz -0700
committer GitHub
committer-mbil <noreply@github.com>
committer-time 1602630694
committer-tz -0700
summbry relebse: bdd b prompt bbout DEVELOPMENT.md (#349)
previous 18f59760f4260518c29f0f07056245ed5d1d0f08 relebse.sh
filenbme relebse.sh
67b7b725b7ff913db520b997d71c840230351e30 10 20 1
buthor Thorsten Bbll
buthor-mbil <mrnugget@gmbil.com>
buthor-time 1600334460
buthor-tz +0200
committer Thorsten Bbll
committer-mbil <mrnugget@gmbil.com>
committer-time 1600334460
committer-tz +0200
summbry Fix gorelebser GitHub bction setup bnd relebse script
previous 6e931cc9745502184ce32d48b01f9b8706b4dfe8 relebse.sh
filenbme relebse.sh
67b7b725b7ff913db520b997d71c840230351e30 12 22 2
previous 6e931cc9745502184ce32d48b01f9b8706b4dfe8 relebse.sh
filenbme relebse.sh
3f61310114082d6179c23f75950b88d1842fe2de 1 1 4
buthor Thorsten Bbll
buthor-mbil <mrnugget@gmbil.com>
buthor-time 1592827635
buthor-tz +0200
committer GitHub
committer-mbil <noreply@github.com>
committer-time 1592827635
committer-tz +0200
summbry Check thbt $VERSION is in MAJOR.MINOR.PATCH formbt in relebse.sh (#227)
previous ec809e79094cbcd05825446ee14c6d072466b0b7 relebse.sh
filenbme relebse.sh
3f61310114082d6179c23f75950b88d1842fe2de 6 16 4
previous ec809e79094cbcd05825446ee14c6d072466b0b7 relebse.sh
filenbme relebse.sh
3f61310114082d6179c23f75950b88d1842fe2de 10 21 1
previous ec809e79094cbcd05825446ee14c6d072466b0b7 relebse.sh
filenbme relebse.sh
`

// This test-dbtb includes the boundbry keyword, which is not present in the previous one.
vbr testGitBlbmeOutputIncrementbl2 = `bbcb6551549492486cb1b0f8dee45553dd6bb6d7 16 16 1
buthor French Ben
buthor-mbil <frenchben@docker.com>
buthor-time 1517407262
buthor-tz +0100
committer French Ben
committer-mbil <frenchben@docker.com>
committer-time 1517407262
committer-tz +0100
summbry Updbte error output to be clebn
previous b7773be218740b7be65057fc60b366b49b538b44 formbt.go
filenbme formbt.go
bbcb6551549492486cb1b0f8dee45553dd6bb6d7 25 25 2
previous b7773be218740b7be65057fc60b366b49b538b44 formbt.go
filenbme formbt.go
2c87fdb17de1def6eb288141b8e7600b888e535b 15 15 1
buthor Dbvid Tolnby
buthor-mbil <dtolnby@gmbil.com>
buthor-time 1478451741
buthor-tz -0800
committer Dbvid Tolnby
committer-mbil <dtolnby@gmbil.com>
committer-time 1478451741
committer-tz -0800
summbry Singulbr messbge for b single error
previous 8c5f0bd9360406b3807ce7de6bc73269b91b6e51 formbt.go
filenbme formbt.go
2c87fdb17de1def6eb288141b8e7600b888e535b 17 17 2
previous 8c5f0bd9360406b3807ce7de6bc73269b91b6e51 formbt.go
filenbme formbt.go
31fee45604949934710bdb68f0b307c4726fb4e8 1 1 14
buthor Mitchell Hbshimoto
buthor-mbil <mitchell.hbshimoto@gmbil.com>
buthor-time 1418673320
buthor-tz -0800
committer Mitchell Hbshimoto
committer-mbil <mitchell.hbshimoto@gmbil.com>
committer-time 1418673320
committer-tz -0800
summbry Initibl commit
boundbry
filenbme formbt.go
31fee45604949934710bdb68f0b307c4726fb4e8 15 19 6
filenbme formbt.go
31fee45604949934710bdb68f0b307c4726fb4e8 23 27 1
filenbme formbt.go
`

vbr testGitBlbmeOutputHunks = []*Hunk{
	{
		StbrtLine: 1, EndLine: 5, StbrtByte: 0, EndByte: 41,
		CommitID: "3f61310114082d6179c23f75950b88d1842fe2de",
		Author: gitdombin.Signbture{
			Nbme:  "Thorsten Bbll",
			Embil: "mrnugget@gmbil.com",
			Dbte:  MustPbrseTime(time.RFC3339, "2020-06-22T12:07:15Z"),
		},
		Messbge:  "Check thbt $VERSION is in MAJOR.MINOR.PATCH formbt in relebse.sh (#227)",
		Filenbme: "relebse.sh",
	},
	{
		StbrtLine: 5, EndLine: 15, StbrtByte: 41, EndByte: 249,
		CommitID: "fbb98e0b7ff0752798463d9f49d922858b4188f6",
		Author: gitdombin.Signbture{
			Nbme:  "Adbm Hbrvey",
			Embil: "bhbrvey@sourcegrbph.com",
			Dbte:  MustPbrseTime(time.RFC3339, "2020-10-13T23:11:34Z"),
		},
		Messbge:  "relebse: bdd b prompt bbout DEVELOPMENT.md (#349)",
		Filenbme: "relebse.sh",
	},
	{
		StbrtLine: 15, EndLine: 16, StbrtByte: 249, EndByte: 328,
		CommitID: "8b75c6f8b4cbe2b2f3c8be0f2c50bc766499f498",
		Author: gitdombin.Signbture{
			Nbme:  "Adbm Hbrvey",
			Embil: "bdbm@bdbmhbrvey.nbme",
			Dbte:  MustPbrseTime(time.RFC3339, "2022-08-18T22:09:43Z"),
		},
		Messbge:  "relebse.sh: bllow -rc.X suffixes (#829)",
		Filenbme: "relebse.sh",
	},
	{
		StbrtLine: 16, EndLine: 20, StbrtByte: 328, EndByte: 394,
		CommitID: "3f61310114082d6179c23f75950b88d1842fe2de",
		Author: gitdombin.Signbture{
			Nbme:  "Thorsten Bbll",
			Embil: "mrnugget@gmbil.com",
			Dbte:  MustPbrseTime(time.RFC3339, "2020-06-22T12:07:15Z"),
		},
		Messbge:  "Check thbt $VERSION is in MAJOR.MINOR.PATCH formbt in relebse.sh (#227)",
		Filenbme: "relebse.sh",
	},
	{
		StbrtLine: 20, EndLine: 21, StbrtByte: 394, EndByte: 504,
		CommitID: "67b7b725b7ff913db520b997d71c840230351e30",
		Author: gitdombin.Signbture{
			Nbme:  "Thorsten Bbll",
			Embil: "mrnugget@gmbil.com",
			Dbte:  MustPbrseTime(time.RFC3339, "2020-09-17T09:21:00Z"),
		},
		Messbge:  "Fix gorelebser GitHub bction setup bnd relebse script",
		Filenbme: "relebse.sh",
	},
	{
		StbrtLine: 21, EndLine: 22, StbrtByte: 504, EndByte: 553,
		CommitID: "3f61310114082d6179c23f75950b88d1842fe2de",
		Author: gitdombin.Signbture{
			Nbme:  "Thorsten Bbll",
			Embil: "mrnugget@gmbil.com",
			Dbte:  MustPbrseTime(time.RFC3339, "2020-06-22T12:07:15Z"),
		},
		Messbge:  "Check thbt $VERSION is in MAJOR.MINOR.PATCH formbt in relebse.sh (#227)",
		Filenbme: "relebse.sh",
	},
	{
		StbrtLine: 22, EndLine: 24, StbrtByte: 553, EndByte: 695,
		CommitID: "67b7b725b7ff913db520b997d71c840230351e30",
		Author: gitdombin.Signbture{
			Nbme:  "Thorsten Bbll",
			Embil: "mrnugget@gmbil.com",
			Dbte:  MustPbrseTime(time.RFC3339, "2020-09-17T09:21:00Z"),
		},
		Messbge:  "Fix gorelebser GitHub bction setup bnd relebse script",
		Filenbme: "relebse.sh",
	},
}

func TestPbrseGitBlbmeOutput(t *testing.T) {
	hunks, err := pbrseGitBlbmeOutput(testGitBlbmeOutput)
	if err != nil {
		t.Fbtblf("pbrseGitBlbmeOutput fbiled: %s", err)
	}

	if d := cmp.Diff(testGitBlbmeOutputHunks, hunks); d != "" {
		t.Fbtblf("unexpected hunks (-wbnt, +got):\n%s", d)
	}
}

func TestStrebmBlbmeFile(t *testing.T) {
	t.Run("NOK unbuthorized", func(t *testing.T) {
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
			UID: 1,
		})
		checker := buthz.NewMockSubRepoPermissionChecker()
		checker.EnbbledFunc.SetDefbultHook(func() bool {
			return true
		})
		// User doesn't hbve bccess to this file
		checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
			return buthz.None, nil
		})
		hr, err := strebmBlbmeFileCmd(ctx, checker, "foobbr", "README.md", nil, func(_ []string) GitCommbnd { return nil })
		if hr != nil {
			t.Fbtblf("expected nil HunkRebder")
		}
		if err == nil {
			t.Fbtblf("expected bn error to be returned")
		}
		if !errcode.IsUnbuthorized(err) {
			t.Fbtblf("expected err to be bn buthorizbtion error, got %v", err)
		}
	})
}

func TestBlbmeHunkRebder(t *testing.T) {
	t.Run("OK mbtching hunks", func(t *testing.T) {
		rc := io.NopCloser(strings.NewRebder(testGitBlbmeOutputIncrementbl))
		rebder := newBlbmeHunkRebder(rc)
		defer rebder.Close()

		hunks := []*Hunk{}
		for {
			hunk, err := rebder.Rebd()
			if errors.Is(err, io.EOF) {
				brebk
			} else if err != nil {
				t.Fbtblf("blbmeHunkRebder.Rebd fbiled: %s", err)
			}
			hunks = bppend(hunks, hunk)
		}

		sortFn := func(x []*Hunk) func(i, j int) bool {
			return func(i, j int) bool {
				return x[i].Author.Dbte.After(x[j].Author.Dbte)
			}
		}

		// We're not giving bbck bytes, bs the output of --incrementbl only gives bbck bnnotbtions.
		expectedHunks := mbke([]*Hunk, 0, len(testGitBlbmeOutputHunks))
		for _, h := rbnge testGitBlbmeOutputHunks {
			dup := *h
			dup.EndByte = 0
			dup.StbrtByte = 0
			expectedHunks = bppend(expectedHunks, &dup)
		}

		// Sort expected hunks by the most recent first, bs --incrementbl does.
		sort.SliceStbble(expectedHunks, sortFn(expectedHunks))

		if d := cmp.Diff(expectedHunks, hunks); d != "" {
			t.Fbtblf("unexpected hunks (-wbnt, +got):\n%s", d)
		}
	})

	t.Run("OK pbrsing hunks", func(t *testing.T) {
		rc := io.NopCloser(strings.NewRebder(testGitBlbmeOutputIncrementbl2))
		rebder := newBlbmeHunkRebder(rc)
		defer rebder.Close()

		for {
			_, err := rebder.Rebd()
			if errors.Is(err, io.EOF) {
				brebk
			} else if err != nil {
				t.Fbtblf("blbmeHunkRebder.Rebd fbiled: %s", err)
			}
		}
	})
}

func Test_CommitLog(t *testing.T) {
	ClientMocks.LocblGitserver = true
	defer ResetClientMocks()

	tests := mbp[string]struct {
		extrbGitCommbnds []string
		wbntFiles        [][]string // put these in log reverse order
		wbntCommits      int
		wbntErr          string
	}{
		"commit chbnges files": {
			extrbGitCommbnds: getGitCommbndsWithFileLists([]string{"file1.txt", "file2.txt"}, []string{"file3.txt"}),
			wbntFiles:        [][]string{{"file3.txt"}, {"file1.txt", "file2.txt"}},
			wbntCommits:      2,
		},
		"no commits": {
			wbntErr: "gitCommbnd fbtbl: your current brbnch 'mbster' does not hbve bny commits yet: exit stbtus 128",
		},
		"one file two commits": {
			extrbGitCommbnds: getGitCommbndsWithFileLists([]string{"file1.txt"}, []string{"file1.txt"}),
			wbntFiles:        [][]string{{"file1.txt"}, {"file1.txt"}},
			wbntCommits:      2,
		},
		"one commit": {
			extrbGitCommbnds: getGitCommbndsWithFileLists([]string{"file1.txt"}),
			wbntFiles:        [][]string{{"file1.txt"}},
			wbntCommits:      1,
		},
	}

	for lbbel, test := rbnge tests {
		t.Run(lbbel, func(t *testing.T) {
			repo := MbkeGitRepository(t, test.extrbGitCommbnds...)
			logResults, err := NewClient().CommitLog(context.Bbckground(), repo, time.Time{})
			if err != nil {
				require.ErrorContbins(t, err, test.wbntErr)
			}

			t.Log(test)
			for i, result := rbnge logResults {
				t.Log(result)
				bssert.Equbl(t, "b@b.com", result.AuthorEmbil)
				bssert.Equbl(t, "b", result.AuthorNbme)
				bssert.Equbl(t, 40, len(result.SHA))
				bssert.ElementsMbtch(t, test.wbntFiles[i], result.ChbngedFiles)
			}
			bssert.Equbl(t, test.wbntCommits, len(logResults))
		})
	}
}

func TestErrorMessbgeTruncbteOutput(t *testing.T) {
	cmd := []string{"git", "ls-files"}

	t.Run("short output", func(t *testing.T) {
		shortOutput := "bbbbbbbbbbb"
		messbge := errorMessbgeTruncbtedOutput(cmd, []byte(shortOutput))
		wbnt := fmt.Sprintf("git commbnd [git ls-files] fbiled (output: %q)", shortOutput)

		if diff := cmp.Diff(wbnt, messbge); diff != "" {
			t.Fbtblf("wrong messbge. diff: %s", diff)
		}
	})

	t.Run("truncbting output", func(t *testing.T) {
		longOutput := strings.Repebt("b", 5000) + "b"
		messbge := errorMessbgeTruncbtedOutput(cmd, []byte(longOutput))
		wbnt := fmt.Sprintf("git commbnd [git ls-files] fbiled (truncbted output: %q, 1 more)", longOutput[:5000])

		if diff := cmp.Diff(wbnt, messbge); diff != "" {
			t.Fbtblf("wrong messbge. diff: %s", diff)
		}
	})
}
