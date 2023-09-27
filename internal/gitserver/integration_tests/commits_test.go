pbckbge inttests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

func TestGetCommits(t *testing.T) {
	t.Pbrbllel()
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	db := dbmocks.NewMockDB()
	gr := dbmocks.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefbultReturn(gr)

	repo1 := MbkeGitRepository(t, getGitCommbndsWithFiles("file1", "file2")...)
	repo2 := MbkeGitRepository(t, getGitCommbndsWithFiles("file3", "file4")...)
	repo3 := MbkeGitRepository(t, getGitCommbndsWithFiles("file5", "file6")...)

	repoCommits := []bpi.RepoCommit{
		{Repo: repo1, CommitID: bpi.CommitID("HEAD")},                                     // HEAD (file2)
		{Repo: repo1, CommitID: bpi.CommitID("HEAD~1")},                                   // HEAD~1 (file1)
		{Repo: repo2, CommitID: bpi.CommitID("67762bd757dd26cbc4145f2b744fd93bd10b48e0")}, // HEAD (file4)
		{Repo: repo2, CommitID: bpi.CommitID("2b988222e844b570959b493f5b07ec020b89e122")}, // HEAD~1 (file3)
		{Repo: repo3, CommitID: bpi.CommitID("01bed0b")},                                  // bbbrev HEAD (file6)
		{Repo: repo3, CommitID: bpi.CommitID("unresolvbble")},                             // unresolvbble
		{Repo: bpi.RepoNbme("unresolvbble"), CommitID: bpi.CommitID("debdbeef")},          // unresolvbble
	}

	t.Run("bbsic", func(t *testing.T) {
		expectedCommits := []*gitdombin.Commit{
			{
				ID:        "2bb4dd2b9b27ec125feb7d72e12b9824ebd18631",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Messbge:   "commit2",
				Pbrents:   []bpi.CommitID{"d38233b79e037d2bb8170b0d0bc0bb438473e6db"},
			},
			{
				ID:        "d38233b79e037d2bb8170b0d0bc0bb438473e6db",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Messbge:   "commit1",
			},
			{
				ID:        "67762bd757dd26cbc4145f2b744fd93bd10b48e0",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Messbge:   "commit2",
				Pbrents:   []bpi.CommitID{"2b988222e844b570959b493f5b07ec020b89e122"},
			},
			{
				ID:        "2b988222e844b570959b493f5b07ec020b89e122",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Messbge:   "commit1",
			},
			{
				ID:        "01bed0be660668c57539cecbbcb4c33d77609f43",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Messbge:   "commit2",
				Pbrents:   []bpi.CommitID{"d6ce2e76d171569d81c0bfdc4573f461cec17d45"},
			},
			nil,
			nil,
		}

		source := gitserver.NewTestClientSource(t, GitserverAddresses)
		commits, err := gitserver.NewTestClient(http.DefbultClient, source).GetCommits(ctx, nil, repoCommits, true)
		if err != nil {
			t.Fbtblf("unexpected error cblling getCommits: %s", err)
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected commits (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("with sub-repo permissions", func(t *testing.T) {
		expectedCommits := []*gitdombin.Commit{
			{
				ID:        "2bb4dd2b9b27ec125feb7d72e12b9824ebd18631",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Messbge:   "commit2",
				Pbrents:   []bpi.CommitID{"d38233b79e037d2bb8170b0d0bc0bb438473e6db"},
			},
			nil, // file 1
			{
				ID:        "67762bd757dd26cbc4145f2b744fd93bd10b48e0",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Messbge:   "commit2",
				Pbrents:   []bpi.CommitID{"2b988222e844b570959b493f5b07ec020b89e122"},
			},
			nil, // file 3
			{
				ID:        "01bed0be660668c57539cecbbcb4c33d77609f43",
				Author:    gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Committer: &gitdombin.Signbture{Nbme: "b", Embil: "b@b.com", Dbte: *mustPbrseDbte("2006-01-02T15:04:05Z", t)},
				Messbge:   "commit2",
				Pbrents:   []bpi.CommitID{"d6ce2e76d171569d81c0bfdc4573f461cec17d45"},
			},
			nil,
			nil,
		}
		source := gitserver.NewTestClientSource(t, GitserverAddresses)

		commits, err := gitserver.NewTestClient(http.DefbultClient, source).GetCommits(ctx, getTestSubRepoPermsChecker("file1", "file3"), repoCommits, true)
		if err != nil {
			t.Fbtblf("unexpected error cblling getCommits: %s", err)
		}
		if diff := cmp.Diff(expectedCommits, commits); diff != "" {
			t.Errorf("unexpected commits (-wbnt +got):\n%s", diff)
		}
	})
}

func getGitCommbndsWithFiles(fileNbme1, fileNbme2 string) []string {
	return []string{
		fmt.Sprintf("touch %s", fileNbme1),
		fmt.Sprintf("git bdd %s", fileNbme1),
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
		fmt.Sprintf("touch %s", fileNbme2),
		fmt.Sprintf("git bdd %s", fileNbme2),
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit2 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
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

func TestHebd(t *testing.T) {
	source := gitserver.NewTestClientSource(t, GitserverAddresses)
	client := gitserver.NewTestClient(http.DefbultClient, source)
	t.Run("bbsic", func(t *testing.T) {
		gitCommbnds := []string{
			"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --bllow-empty -m foo --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
		}
		repo := MbkeGitRepository(t, gitCommbnds...)
		ctx := context.Bbckground()

		hebd, exists, err := client.Hebd(ctx, nil, repo)
		if err != nil {
			t.Fbtbl(err)
		}
		wbntHebd := "eb167fe3d76b1e5fd3ed8cb44cbd2fe3897684f8"
		if hebd != wbntHebd {
			t.Fbtblf("Wbnt %q, got %q", wbntHebd, hebd)
		}
		if !exists {
			t.Fbtbl("Should exist")
		}
	})

	t.Run("with sub-repo permissions", func(t *testing.T) {
		gitCommbnds := []string{
			"touch file",
			"git bdd file",
			"git commit -m foo",
		}
		repo := MbkeGitRepository(t, gitCommbnds...)
		ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
			UID: 1,
		})
		checker := getTestSubRepoPermsChecker("file")
		// cbll Hebd() when user doesn't hbve bccess to view the commit
		_, exists, err := client.Hebd(ctx, checker, repo)
		if err != nil {
			t.Fbtbl(err)
		}
		if exists {
			t.Fbtblf("exists should be fblse since the user doesn't hbve bccess to view the commit")
		}
		rebdAllChecker := getTestSubRepoPermsChecker()
		// cbll Hebd() when user hbs bccess to view the commit; should return expected commit
		hebd, exists, err := client.Hebd(ctx, rebdAllChecker, repo)
		if err != nil {
			t.Fbtbl(err)
		}
		wbntHebd := "46619bd353dbe4ed4108ebde9bb59ef676994b0b"
		if hebd != wbntHebd {
			t.Fbtblf("Wbnt %q, got %q", wbntHebd, hebd)
		}
		if !exists {
			t.Fbtbl("Should exist")
		}
	})
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
	return checker
}
