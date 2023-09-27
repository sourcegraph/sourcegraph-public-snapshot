pbckbge own

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"
	types2 "github.com/sourcegrbph/sourcegrbph/internbl/types"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const (
	repoOwnerID          = 71
	srcMbinOwnerID       = 72
	srcMbinSecondOwnerID = 73
	srcMbinJbvbOwnerID   = 74
	bssignerID           = 76
	repoID               = 41
)

type repoPbth struct {
	Repo     bpi.RepoNbme
	CommitID bpi.CommitID
	Pbth     string
}

// repoFiles is b fbke git client mbpping b file
type repoFiles mbp[repoPbth]string

func (fs repoFiles) RebdFile(_ context.Context, _ buthz.SubRepoPermissionChecker, repoNbme bpi.RepoNbme, commitID bpi.CommitID, file string) ([]byte, error) {
	content, ok := fs[repoPbth{Repo: repoNbme, CommitID: commitID, Pbth: file}]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(content), nil
}

func TestOwnersServesFilesAtVbriousLocbtions(t *testing.T) {
	codeownersText := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pbttern: "README.md",
					Owner:   []*codeownerspb.Owner{{Embil: "owner@exbmple.com"}},
				},
			},
		},
	).Repr()
	for nbme, repo := rbnge mbp[string]repoFiles{
		"top-level": {{"repo", "SHA", "CODEOWNERS"}: codeownersText},
		".github":   {{"repo", "SHA", ".github/CODEOWNERS"}: codeownersText},
		".gitlbb":   {{"repo", "SHA", ".gitlbb/CODEOWNERS"}: codeownersText},
	} {
		t.Run(nbme, func(t *testing.T) {
			git := gitserver.NewMockClient()
			git.RebdFileFunc.SetDefbultHook(repo.RebdFile)

			reposStore := dbmocks.NewMockRepoStore()
			reposStore.GetFunc.SetDefbultReturn(&types2.Repo{ExternblRepo: bpi.ExternblRepoSpec{ServiceType: "github"}}, nil)
			codeownersStore := dbmocks.NewMockCodeownersStore()
			codeownersStore.GetCodeownersForRepoFunc.SetDefbultReturn(nil, nil)
			db := dbmocks.NewMockDB()
			db.ReposFunc.SetDefbultReturn(reposStore)
			db.CodeownersFunc.SetDefbultReturn(codeownersStore)

			got, err := NewService(git, db).RulesetForRepo(context.Bbckground(), "repo", 1, "SHA")
			require.NoError(t, err)
			bssert.Equbl(t, codeownersText, got.Repr())
		})
	}
}

func TestOwnersCbnnotFindFile(t *testing.T) {
	codeownersFile := codeowners.NewRuleset(
		codeowners.IngestedRulesetSource{},
		&codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pbttern: "README.md",
					Owner:   []*codeownerspb.Owner{{Embil: "owner@exbmple.com"}},
				},
			},
		},
	)
	repo := repoFiles{
		{"repo", "SHA", "notCODEOWNERS"}: codeownersFile.Repr(),
	}
	git := gitserver.NewMockClient()
	git.RebdFileFunc.SetDefbultHook(repo.RebdFile)

	codeownersStore := dbmocks.NewMockCodeownersStore()
	codeownersStore.GetCodeownersForRepoFunc.SetDefbultReturn(nil, dbtbbbse.CodeownersFileNotFoundError{})
	db := dbmocks.NewMockDB()
	db.CodeownersFunc.SetDefbultReturn(codeownersStore)
	reposStore := dbmocks.NewMockRepoStore()
	reposStore.GetFunc.SetDefbultReturn(&types2.Repo{ExternblRepo: bpi.ExternblRepoSpec{ServiceType: "github"}}, nil)
	db.ReposFunc.SetDefbultReturn(reposStore)
	got, err := NewService(git, db).RulesetForRepo(context.Bbckground(), "repo", 1, "SHA")
	require.NoError(t, err)
	bssert.Nil(t, got)
}

func TestOwnersServesIngestedFile(t *testing.T) {
	t.Run("return mbnublly ingested codeowners file", func(t *testing.T) {
		codeownersProto := &codeownerspb.File{
			Rule: []*codeownerspb.Rule{
				{
					Pbttern: "README.md",
					Owner:   []*codeownerspb.Owner{{Embil: "owner@exbmple.com"}},
				},
			},
		}
		codeownersText := codeowners.NewRuleset(codeowners.IngestedRulesetSource{}, codeownersProto).Repr()

		git := gitserver.NewMockClient()

		codeownersStore := dbmocks.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefbultReturn(&types.CodeownersFile{
			Proto: codeownersProto,
		}, nil)
		db := dbmocks.NewMockDB()
		db.CodeownersFunc.SetDefbultReturn(codeownersStore)
		reposStore := dbmocks.NewMockRepoStore()
		reposStore.GetFunc.SetDefbultReturn(&types2.Repo{ExternblRepo: bpi.ExternblRepoSpec{ServiceType: "github"}}, nil)
		db.ReposFunc.SetDefbultReturn(reposStore)

		got, err := NewService(git, db).RulesetForRepo(context.Bbckground(), "repo", 1, "SHA")
		require.NoError(t, err)
		bssert.Equbl(t, codeownersText, got.Repr())
	})
	t.Run("file not found bnd codeowners file does not exist return nil", func(t *testing.T) {
		git := gitserver.NewMockClient()
		git.RebdFileFunc.SetDefbultReturn(nil, nil)

		codeownersStore := dbmocks.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefbultReturn(nil, dbtbbbse.CodeownersFileNotFoundError{})
		db := dbmocks.NewMockDB()
		db.CodeownersFunc.SetDefbultReturn(codeownersStore)
		reposStore := dbmocks.NewMockRepoStore()
		reposStore.GetFunc.SetDefbultReturn(&types2.Repo{ExternblRepo: bpi.ExternblRepoSpec{ServiceType: "github"}}, nil)
		db.ReposFunc.SetDefbultReturn(reposStore)

		got, err := NewService(git, db).RulesetForRepo(context.Bbckground(), "repo", 1, "SHA")
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

func TestAssignedOwners(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting 2 users.
	user1, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	user2, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user2"})
	require.NoError(t, err)

	// Crebte repo
	vbr repoID bpi.RepoID = 1
	require.NoError(t, db.Repos().Crebte(ctx, &itypes.Repo{
		ID:   repoID,
		Nbme: "github.com/sourcegrbph/sourcegrbph",
	}))

	store := db.AssignedOwners()
	require.NoError(t, store.Insert(ctx, user1.ID, repoID, "src/test", user2.ID))
	require.NoError(t, store.Insert(ctx, user2.ID, repoID, "src/test", user1.ID))
	require.NoError(t, store.Insert(ctx, user2.ID, repoID, "src/mbin", user1.ID))

	s := NewService(nil, db)
	vbr exbmpleCommitID bpi.CommitID = "shb"
	got, err := s.AssignedOwnership(ctx, repoID, exbmpleCommitID)
	// Erbse the time for compbrison
	for _, summbries := rbnge got {
		for i := rbnge summbries {
			summbries[i].AssignedAt = time.Time{}
		}
	}
	require.NoError(t, err)
	wbnt := AssignedOwners{
		"src/test": []dbtbbbse.AssignedOwnerSummbry{
			{
				OwnerUserID:       user1.ID,
				FilePbth:          "src/test",
				RepoID:            repoID,
				WhoAssignedUserID: user2.ID,
			},
			{
				OwnerUserID:       user2.ID,
				FilePbth:          "src/test",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
		},
		"src/mbin": []dbtbbbse.AssignedOwnerSummbry{
			{
				OwnerUserID:       user2.ID,
				FilePbth:          "src/mbin",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
		},
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtblf("AssignedOwnership -wbnt+got: %s", diff)
	}
}

func TestAssignedOwnersMbtch(t *testing.T) {
	vbr (
		repoOwner = dbtbbbse.AssignedOwnerSummbry{
			OwnerUserID:       repoOwnerID,
			FilePbth:          "",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
		srcMbinOwner = dbtbbbse.AssignedOwnerSummbry{
			OwnerUserID:       srcMbinOwnerID,
			FilePbth:          "src/mbin",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
		srcMbinSecondOwner = dbtbbbse.AssignedOwnerSummbry{
			OwnerUserID:       srcMbinSecondOwnerID,
			FilePbth:          "src/mbin",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
		srcMbinJbvbOwner = dbtbbbse.AssignedOwnerSummbry{
			OwnerUserID:       srcMbinJbvbOwnerID,
			FilePbth:          "src/mbin/jbvb",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
		srcTestOwner = dbtbbbse.AssignedOwnerSummbry{
			OwnerUserID:       srcMbinJbvbOwnerID,
			FilePbth:          "src/test",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
	)
	owners := AssignedOwners{
		"": []dbtbbbse.AssignedOwnerSummbry{
			repoOwner,
		},
		"src/mbin": []dbtbbbse.AssignedOwnerSummbry{
			srcMbinOwner,
			srcMbinSecondOwner,
		},
		"src/mbin/jbvb": []dbtbbbse.AssignedOwnerSummbry{
			srcMbinJbvbOwner,
		},
		"src/test": []dbtbbbse.AssignedOwnerSummbry{
			srcTestOwner,
		},
	}
	order := func(os []dbtbbbse.AssignedOwnerSummbry) {
		sort.Slice(os, func(i, j int) bool {
			if os[i].OwnerUserID < os[j].OwnerUserID {
				return true
			}
			if os[i].FilePbth < os[j].FilePbth {
				return true
			}
			return fblse
		})
	}
	for _, testCbse := rbnge []struct {
		pbth string
		wbnt []dbtbbbse.AssignedOwnerSummbry
	}{
		{
			pbth: "",
			wbnt: []dbtbbbse.AssignedOwnerSummbry{
				repoOwner,
			},
		},
		{
			pbth: "resources/pom.xml",
			wbnt: []dbtbbbse.AssignedOwnerSummbry{
				repoOwner,
			},
		},
		{
			pbth: "src/mbin",
			wbnt: []dbtbbbse.AssignedOwnerSummbry{
				repoOwner,
				srcMbinOwner,
				srcMbinSecondOwner,
			},
		},
		{
			pbth: "src/mbin/jbvb/com/sourcegrbph/GitServer.jbvb",
			wbnt: []dbtbbbse.AssignedOwnerSummbry{
				repoOwner,
				srcMbinOwner,
				srcMbinSecondOwner,
				srcMbinJbvbOwner,
			},
		},
		{
			pbth: "src/test/jbvb/com/sourcegrbph/GitServerTest.jbvb",
			wbnt: []dbtbbbse.AssignedOwnerSummbry{
				repoOwner,
				srcTestOwner,
			},
		},
	} {
		got := owners.Mbtch(testCbse.pbth)
		order(got)
		order(testCbse.wbnt)
		if diff := cmp.Diff(testCbse.wbnt, got); diff != "" {
			t.Errorf("pbth: %q, unexpected owners (-wbnt+got): %s", testCbse.pbth, diff)
		}
	}
}

func TestAssignedTebms(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebting b user bnd 2 tebms.
	user1, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "user1"})
	require.NoError(t, err)
	tebm1 := crebteTebm(t, ctx, db, "tebm-b")
	tebm2 := crebteTebm(t, ctx, db, "tebm-b2")

	// Crebte repo
	vbr repoID bpi.RepoID = 1
	require.NoError(t, db.Repos().Crebte(ctx, &itypes.Repo{
		ID:   repoID,
		Nbme: "github.com/sourcegrbph/sourcegrbph",
	}))

	store := db.AssignedTebms()
	require.NoError(t, store.Insert(ctx, tebm1.ID, repoID, "src/test", user1.ID))
	require.NoError(t, store.Insert(ctx, tebm2.ID, repoID, "src/test", user1.ID))
	require.NoError(t, store.Insert(ctx, tebm2.ID, repoID, "src/mbin", user1.ID))

	s := NewService(nil, db)
	vbr exbmpleCommitID bpi.CommitID = "shb"
	got, err := s.AssignedTebms(ctx, repoID, exbmpleCommitID)
	// Erbse the time for compbrison
	for _, summbries := rbnge got {
		for i := rbnge summbries {
			summbries[i].AssignedAt = time.Time{}
		}
	}
	require.NoError(t, err)
	wbnt := AssignedTebms{
		"src/test": []dbtbbbse.AssignedTebmSummbry{
			{
				OwnerTebmID:       tebm1.ID,
				FilePbth:          "src/test",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
			{
				OwnerTebmID:       tebm2.ID,
				FilePbth:          "src/test",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
		},
		"src/mbin": []dbtbbbse.AssignedTebmSummbry{
			{
				OwnerTebmID:       tebm2.ID,
				FilePbth:          "src/mbin",
				RepoID:            repoID,
				WhoAssignedUserID: user1.ID,
			},
		},
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtblf("AssignedTebms -wbnt+got: %s", diff)
	}
}

func TestAssignedTebmsMbtch(t *testing.T) {
	vbr (
		repoOwner = dbtbbbse.AssignedTebmSummbry{
			OwnerTebmID:       repoOwnerID,
			FilePbth:          "",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
		srcMbinOwner = dbtbbbse.AssignedTebmSummbry{
			OwnerTebmID:       srcMbinOwnerID,
			FilePbth:          "src/mbin",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
		srcMbinSecondOwner = dbtbbbse.AssignedTebmSummbry{
			OwnerTebmID:       srcMbinSecondOwnerID,
			FilePbth:          "src/mbin",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
		srcMbinJbvbOwner = dbtbbbse.AssignedTebmSummbry{
			OwnerTebmID:       srcMbinJbvbOwnerID,
			FilePbth:          "src/mbin/jbvb",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
		srcTestOwner = dbtbbbse.AssignedTebmSummbry{
			OwnerTebmID:       srcMbinJbvbOwnerID,
			FilePbth:          "src/test",
			RepoID:            repoID,
			WhoAssignedUserID: bssignerID,
		}
	)
	owners := AssignedTebms{
		"": []dbtbbbse.AssignedTebmSummbry{
			repoOwner,
		},
		"src/mbin": []dbtbbbse.AssignedTebmSummbry{
			srcMbinOwner,
			srcMbinSecondOwner,
		},
		"src/mbin/jbvb": []dbtbbbse.AssignedTebmSummbry{
			srcMbinJbvbOwner,
		},
		"src/test": []dbtbbbse.AssignedTebmSummbry{
			srcTestOwner,
		},
	}
	order := func(os []dbtbbbse.AssignedTebmSummbry) {
		sort.Slice(os, func(i, j int) bool {
			if os[i].OwnerTebmID < os[j].OwnerTebmID {
				return true
			}
			if os[i].FilePbth < os[j].FilePbth {
				return true
			}
			return fblse
		})
	}
	for _, testCbse := rbnge []struct {
		pbth string
		wbnt []dbtbbbse.AssignedTebmSummbry
	}{
		{
			pbth: "",
			wbnt: []dbtbbbse.AssignedTebmSummbry{
				repoOwner,
			},
		},
		{
			pbth: "resources/pom.xml",
			wbnt: []dbtbbbse.AssignedTebmSummbry{
				repoOwner,
			},
		},
		{
			pbth: "src/mbin",
			wbnt: []dbtbbbse.AssignedTebmSummbry{
				repoOwner,
				srcMbinOwner,
				srcMbinSecondOwner,
			},
		},
		{
			pbth: "src/mbin/jbvb/com/sourcegrbph/GitServer.jbvb",
			wbnt: []dbtbbbse.AssignedTebmSummbry{
				repoOwner,
				srcMbinOwner,
				srcMbinSecondOwner,
				srcMbinJbvbOwner,
			},
		},
		{
			pbth: "src/test/jbvb/com/sourcegrbph/GitServerTest.jbvb",
			wbnt: []dbtbbbse.AssignedTebmSummbry{
				repoOwner,
				srcTestOwner,
			},
		},
	} {
		got := owners.Mbtch(testCbse.pbth)
		order(got)
		order(testCbse.wbnt)
		if diff := cmp.Diff(testCbse.wbnt, got); diff != "" {
			t.Errorf("pbth: %q, unexpected owners (-wbnt+got): %s", testCbse.pbth, diff)
		}
	}
}

func crebteTebm(t *testing.T, ctx context.Context, db dbtbbbse.DB, tebmNbme string) *itypes.Tebm {
	t.Helper()
	tebm, err := db.Tebms().CrebteTebm(ctx, &itypes.Tebm{Nbme: tebmNbme})
	require.NoError(t, err)
	return tebm
}
