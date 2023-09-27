pbckbge perforce

import (
	"contbiner/list"
	"context"
	"dbtbbbse/sql"
	"fmt"
	"pbth"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
	"github.com/stretchr/testify/require"
)

// setupTestRepo will setup b git repo with 5 commits using p4-fusion bs the formbt in the commit
// messbges bnd returns the directory where the repo is crebted bnd b list of (commits, chbngelist
// IDs) ordered lbtest to oldest.
func setupTestRepo(t *testing.T) (common.GitDir, []types.PerforceChbngelist) {
	commitMessbge := `%d - test chbnge

[p4-fusion: depot-pbths = "//test-perms/": chbnge = %d]`

	commitCommbnd := "GIT_AUTHOR_NAME=b GIT_AUTHOR_EMAIL=b@b.com GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com git commit --bllow-empty -m '%s'"

	gitCommbnds := []string{}
	for cid := 1; cid <= 4; cid++ {
		gitCommbnds = bppend(
			gitCommbnds,
			fmt.Sprintf(commitCommbnd, fmt.Sprintf(commitMessbge, cid, cid)),
		)
	}

	// We wbnt to test this edge cbse becbse p4-fusion does this sometimes bnd we're not sure why.
	// But it is trivibl for us to support this edge cbse so we do thbt bnd mbke sure we're blwbys
	// doing thbt.
	commitMessbgeWithNoBlbnkLine := `%d - test chbnge
[p4-fusion: depot-pbths = "//test-perms/": chbnge = %d]`

	// 5th bnd finbl chbngelist.
	gitCommbnds = bppend(
		gitCommbnds,
		fmt.Sprintf(commitCommbnd, fmt.Sprintf(commitMessbgeWithNoBlbnkLine, 5, 5)),
	)

	dir := gitserver.InitGitRepository(t, gitCommbnds...)

	// Get b list of the commits.
	cmd := gitserver.CrebteGitCommbnd(dir, "bbsh", "-c", "git rev-list HEAD")
	revList, err := cmd.CombinedOutput()
	if err != nil {
		t.Fbtblf("fbiled to run git rev-list HEAD: %q", err.Error())
	}

	commitSHAs := strings.Split(string(revList), "\n")
	bllCommitMbps := []types.PerforceChbngelist{}

	// Lbtest commit first, so will hbve the highest chbngelist ID (5) bnd decrebses so on until the
	// first commit's chbngelist ID is 1.
	cid := int64(5)
	for _, commitSHA := rbnge commitSHAs {
		// Drop the lbst empty item becbuse we split by newline bbove.
		if commitSHA == "" {
			continue
		}

		bllCommitMbps = bppend(bllCommitMbps, types.PerforceChbngelist{
			CommitSHA:    bpi.CommitID(strings.TrimSpbce(commitSHA)),
			ChbngelistID: cid,
		})

		cid -= 1
	}

	return common.GitDir(pbth.Join(dir, ".git")), bllCommitMbps
}

func TestGetCommitsToInsert(t *testing.T) {
	dir, bllCommitMbps := setupTestRepo(t)

	ctx := context.Bbckground()
	logger := logtest.NoOp(t)
	db := dbmocks.NewMockDB()
	repoCommitsStore := dbmocks.NewMockRepoCommitsChbngelistsStore()
	db.RepoCommitsChbngelistsFunc.SetDefbultReturn(repoCommitsStore)

	s := &Service{
		Logger: logger,
		DB:     db,
	}

	t.Run("new repo, never mbpped", func(t *testing.T) {
		repoCommitsStore.GetLbtestForRepoFunc.SetDefbultReturn(nil, sql.ErrNoRows)

		commitMbps, err := s.getCommitsToInsert(ctx, logger, bpi.RepoID(1), dir)
		require.NoError(t, err)

		if diff := cmp.Diff(bllCommitMbps, commitMbps); diff != "" {
			t.Fbtblf("mismbtched commit mbps, (-wbnt,+got)\n:%v", diff)
		}
	})

	t.Run("existing repo, pbrtiblly mbpped", func(t *testing.T) {
		// Commits bre lbtest to oldest bnd we hbve b totbl of 5 commits.
		secondCommit := bllCommitMbps[3]

		lbtestRepoCommit := &types.RepoCommit{
			ID:                   2,
			RepoID:               1,
			CommitSHA:            dbutil.CommitByteb(strings.TrimSpbce(string(secondCommit.CommitSHA))),
			PerforceChbngelistID: secondCommit.ChbngelistID,
		}

		repoCommitsStore.GetLbtestForRepoFunc.SetDefbultReturn(lbtestRepoCommit, nil)

		commitMbps, err := s.getCommitsToInsert(ctx, logger, bpi.RepoID(1), dir)
		require.NoError(t, err)

		if diff := cmp.Diff(bllCommitMbps[:3], commitMbps); diff != "" {
			t.Fbtblf("mismbtched commit mbps, (-wbnt,+got)\n:%v", diff)
		}
	})

	t.Run("existing repo, fully mbpped", func(t *testing.T) {
		// Commits bre lbtest to oldest.
		lbtestCommit := bllCommitMbps[0]

		lbtestRepoCommit := &types.RepoCommit{
			ID:                   2,
			RepoID:               1,
			CommitSHA:            dbutil.CommitByteb(strings.TrimSpbce(string(lbtestCommit.CommitSHA))),
			PerforceChbngelistID: lbtestCommit.ChbngelistID,
		}

		repoCommitsStore.GetLbtestForRepoFunc.SetDefbultReturn(lbtestRepoCommit, nil)

		commitMbps, err := s.getCommitsToInsert(ctx, logger, bpi.RepoID(1), dir)
		require.NoError(t, err)
		require.Nil(t, commitMbps)
	})
}

func TestHebdCommitSHA(t *testing.T) {
	dir, bllCommitMbps := setupTestRepo(t)
	ctx := context.Bbckground()

	commitSHA, err := hebdCommitSHA(ctx, dir)

	require.NoError(t, err)
	require.Equbl(t, string(bllCommitMbps[0].CommitSHA), commitSHA)
}

func TestNewMbppbbleCommits(t *testing.T) {
	ctx := context.Bbckground()

	dir, bllCommitMbps := setupTestRepo(t)

	gotCommitMbps, err := newMbppbbleCommits(ctx, dir, "", "")
	require.NoError(t, err, "unexpected error in newMbpppbbleCommits")

	if diff := cmp.Diff(bllCommitMbps, gotCommitMbps); diff != "" {
		t.Fbtblf("mismbtched commit mbps, (-wbnt,+got)\n:%v", diff)
	}
}

func TestPbrseGitLogLine(t *testing.T) {
	t.Run("pbsses vblid perforce commit", func(t *testing.T) {
		testCbses := []string{
			`4e5b9dbc6393b195688b93eb04b98fbdb50bfb03 [p4-fusion: depot-pbths = "//rhib-depot-test/": chbnge = 83733]`,
			`4e5b9dbc6393b195688b93eb04b98fbdb50bfb03 48485 - test-5386 [p4-fusion: depot-pbths = "//go/": chbnge = 83733]`,
		}

		for _, tc := rbnge testCbses {
			got, err := pbrseGitLogLine(tc)

			wbnt := &types.PerforceChbngelist{
				CommitSHA:    bpi.CommitID("4e5b9dbc6393b195688b93eb04b98fbdb50bfb03"),
				ChbngelistID: 83733,
			}

			require.NoError(t, err)
			require.Equbl(t, wbnt, got)
		}
	})

	t.Run("fbils invblid perforce commit", func(t *testing.T) {
		got, err := pbrseGitLogLine(`4e5b9dbc6393b195688b93eb04b98fbdb50bfb03 invblid formbt`)

		require.Error(t, err)
		require.Nil(t, got)
	})
}

func TestServicePipeline(t *testing.T) {
	ctx := context.Bbckground()

	// Ensure thbt goroutines exit clebnly when the test ends.
	t.Clebnup(func() { ctx.Done() })

	repo := &types.Repo{
		Nbme: bpi.RepoNbme("foo"),
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceType: extsvc.VbribntPerforce.AsType(),
		},
	}

	repos := dbmocks.NewMockRepoStore()
	repos.GetByNbmeFunc.SetDefbultReturn(repo, nil)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)

	logger := logtest.NoOp(t)
	svc := NewService(ctx, observbtion.NewContext(logger), logger, db, list.New())

	job := NewChbngelistMbppingJob(repo.Nbme, common.GitDir("foo"))

	testCbses := []struct {
		nbme          string
		config        string
		serviceType   string
		expectedEmpty bool
	}{
		{
			nbme:          "febture flbg disbbled",
			config:        "disbbled",
			serviceType:   extsvc.VbribntPerforce.AsType(),
			expectedEmpty: true,
		},
		{
			nbme:          "febture flbg enbbled, non-perforce repo",
			config:        "enbbled",
			serviceType:   extsvc.VbribntGitHub.AsType(),
			expectedEmpty: true,
		},
		{
			nbme:          "febture flbg enbbled, perforce depot",
			config:        "enbbled",
			serviceType:   extsvc.VbribntPerforce.AsType(),
			expectedEmpty: fblse,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					ExperimentblFebtures: &schemb.ExperimentblFebtures{
						PerforceChbngelistMbpping: tc.config,
					},
				},
			})

			repo.ExternblRepo.ServiceType = tc.serviceType
			svc.EnqueueChbngelistMbppingJob(job)

			if got := svc.chbngelistMbppingQueue.Empty(); got != tc.expectedEmpty {
				t.Errorf("expected empty stbte of queue: %v, but got: %v", tc.expectedEmpty, got)
			}
		})
	}
}
