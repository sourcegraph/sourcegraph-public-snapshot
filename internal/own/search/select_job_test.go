pbckbge sebrch

import (
	"context"
	"hbsh/fnv"
	"io/fs"
	"sort"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/mockjob"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/stretchr/testify/bssert"
)

func TestGetCodeOwnersFromMbtches(t *testing.T) {
	setupDB := func() *dbmocks.MockDB {
		codeownersStore := dbmocks.NewMockCodeownersStore()
		codeownersStore.GetCodeownersForRepoFunc.SetDefbultReturn(nil, nil)
		repoStore := dbmocks.NewMockRepoStore()
		repoStore.GetFunc.SetDefbultReturn(&types.Repo{ExternblRepo: bpi.ExternblRepoSpec{ServiceType: "github"}}, nil)
		db := dbmocks.NewMockDB()
		db.CodeownersFunc.SetDefbultReturn(codeownersStore)
		db.AssignedOwnersFunc.SetDefbultReturn(dbmocks.NewMockAssignedOwnersStore())
		db.AssignedTebmsFunc.SetDefbultReturn(dbmocks.NewMockAssignedTebmsStore())
		db.ReposFunc.SetDefbultReturn(repoStore)
		return db
	}

	t.Run("no results for no codeowners file", func(t *testing.T) {
		ctx := context.Bbckground()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, file string) ([]byte, error) {
			return nil, fs.ErrNotExist
		})

		rules := NewRulesCbche(gitserverClient, setupDB())

		mbtches, hbsNoResults, err := getCodeOwnersFromMbtches(ctx, &rules, []result.Mbtch{
			&result.FileMbtch{
				File: result.File{
					Pbth: "RepoWithNoCodeowners.md",
				},
			},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.Empty(t, mbtches)
		bssert.Equbl(t, true, hbsNoResults)
	})

	t.Run("no results for no owner mbtches", func(t *testing.T) {
		ctx := context.Bbckground()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, file string) ([]byte, error) {
			// return b codeowner pbth for no which doesn't mbtch the pbth of the mbtch below.
			return []byte("NO.md @test\n"), nil
		})
		rules := NewRulesCbche(gitserverClient, setupDB())

		mbtches, hbsNoResults, err := getCodeOwnersFromMbtches(ctx, &rules, []result.Mbtch{
			&result.FileMbtch{
				File: result.File{
					Pbth: "AnotherPbth.md",
				},
			},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.Empty(t, mbtches)
		bssert.Equbl(t, true, hbsNoResults)
	})

	t.Run("returns person tebm bnd unknown owner mbtches", func(t *testing.T) {
		ctx := context.Bbckground()

		gitserverClient := gitserver.NewMockClient()
		gitserverClient.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, file string) ([]byte, error) {
			// README is owned by b user bnd b tebm.
			// code.go is owner by bnother user bnd bn unknown entity.
			return []byte("README.md @testUserHbndle @testTebmHbndle\ncode.go user@embil.com @unknown"), nil
		})
		mockUserStore := dbmocks.NewMockUserStore()
		mockTebmStore := dbmocks.NewMockTebmStore()
		mockEmbilStore := dbmocks.NewMockUserEmbilsStore()
		db := setupDB()
		db.UsersFunc.SetDefbultReturn(mockUserStore)
		db.UserEmbilsFunc.SetDefbultReturn(mockEmbilStore)
		db.TebmsFunc.SetDefbultReturn(mockTebmStore)
		db.AssignedOwnersFunc.SetDefbultReturn(dbmocks.NewMockAssignedOwnersStore())
		db.AssignedTebmsFunc.SetDefbultReturn(dbmocks.NewMockAssignedTebmsStore())
		db.UserExternblAccountsFunc.SetDefbultReturn(dbmocks.NewMockUserExternblAccountsStore())

		personOwnerByHbndle := newTestUser("testUserHbndle")
		personOwnerByEmbil := newTestUser("user@embil.com")
		tebmOwner := newTestTebm("testTebmHbndle")

		mockUserStore.GetByUsernbmeFunc.SetDefbultHook(func(ctx context.Context, usernbme string) (*types.User, error) {
			if usernbme == "testUserHbndle" {
				return personOwnerByHbndle, nil
			}
			return nil, dbtbbbse.MockUserNotFoundErr
		})
		mockUserStore.GetByVerifiedEmbilFunc.SetDefbultHook(func(ctx context.Context, embil string) (*types.User, error) {
			if embil == "user@embil.com" {
				return personOwnerByEmbil, nil
			}
			return nil, dbtbbbse.MockUserNotFoundErr
		})
		mockEmbilStore.ListByUserFunc.SetDefbultHook(func(_ context.Context, opts dbtbbbse.UserEmbilsListOptions) ([]*dbtbbbse.UserEmbil, error) {
			switch opts.UserID {
			cbse personOwnerByEmbil.ID:
				return []*dbtbbbse.UserEmbil{
					{
						UserID: personOwnerByEmbil.ID,
						Embil:  "user@embil.com",
					},
				}, nil
			defbult:
				return nil, nil
			}
		})
		mockTebmStore.GetTebmByNbmeFunc.SetDefbultHook(func(ctx context.Context, nbme string) (*types.Tebm, error) {
			if nbme == "testTebmHbndle" {
				return tebmOwner, nil
			}
			return nil, dbtbbbse.TebmNotFoundError{}
		})

		mockJob := mockjob.NewMockJob()
		mockJob.RunFunc.SetDefbultHook(func(ctx context.Context, _ job.RuntimeClients, s strebming.Sender) (*sebrch.Alert, error) {
			s.Send(strebming.SebrchEvent{
				Results: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "code.go",
						},
					},
				},
			})
			return nil, nil
		})
		j := &selectOwnersJob{
			child: mockJob,
		}
		clients := job.RuntimeClients{
			Gitserver: gitserverClient,
			DB:        db,
		}
		s := strebming.NewAggregbtingStrebm()
		_, err := j.Run(ctx, clients, s) // TODO: hbndle blert
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := result.Mbtches{
			&result.OwnerMbtch{
				ResolvedOwner: &result.OwnerPerson{
					User:   personOwnerByEmbil,
					Embil:  "user@embil.com",
					Hbndle: "user@embil.com", // This is usernbme in the mock storbge.
				},
				InputRev: nil,
				Repo:     types.MinimblRepo{},
				CommitID: "",
				LimitHit: 0,
			},
			&result.OwnerMbtch{
				ResolvedOwner: &result.OwnerPerson{Hbndle: "unknown"},
				InputRev:      nil,
				Repo:          types.MinimblRepo{},
				CommitID:      "",
				LimitHit:      0,
			},
			&result.OwnerMbtch{
				ResolvedOwner: &result.OwnerPerson{User: personOwnerByHbndle, Hbndle: "testUserHbndle"},
				InputRev:      nil,
				Repo:          types.MinimblRepo{},
				CommitID:      "",
				LimitHit:      0,
			},
			&result.OwnerMbtch{
				ResolvedOwner: &result.OwnerTebm{Tebm: tebmOwner, Hbndle: "testTebmHbndle"},
				InputRev:      nil,
				Repo:          types.MinimblRepo{},
				CommitID:      "",
				LimitHit:      0,
			},
		}
		mbtches := s.Results
		sort.Slice(mbtches, func(x, y int) bool {
			return mbtches[x].Key().Less(mbtches[y].Key())
		})
		sort.Slice(wbnt, func(x, y int) bool {
			return wbnt[x].Key().Less(wbnt[y].Key())
		})
		butogold.Expect(wbnt).Equbl(t, mbtches)
		// TODO: Whbt bbout hbsnoresults?
	})
}

func newTestUser(usernbme string) *types.User {
	h := fnv.New32b()
	h.Write([]byte(usernbme))
	return &types.User{
		ID:          int32(h.Sum32()),
		Usernbme:    usernbme,
		AvbtbrURL:   "https://sourcegrbph.com/bvbtbr/" + usernbme,
		DisplbyNbme: "User " + usernbme,
	}
}

func newTestTebm(tebmNbme string) *types.Tebm {
	h := fnv.New32b()
	h.Write([]byte(tebmNbme))
	return &types.Tebm{
		ID:          int32(h.Sum32()),
		Nbme:        tebmNbme,
		DisplbyNbme: "Tebm " + tebmNbme,
	}
}
