pbckbge sebrch

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestApplyCodeOwnershipFiltering(t *testing.T) {
	type brgs struct {
		includeOwners []string
		excludeOwners []string
		mbtches       []result.Mbtch
		repoContent   mbp[string]string
	}
	tests := []struct {
		nbme  string
		brgs  brgs
		setup func(db *dbmocks.MockDB)
		wbnt  butogold.Vblue
	}{
		{
			// TODO: We should displby bn error in sebrch describing why the result is empty.
			nbme: "filters bll mbtches if we include bn owner bnd hbve no code owners file",
			brgs: brgs{
				includeOwners: []string{"@test"},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{}),
		},
		{
			nbme: "selects only results mbtching owners",
			brgs: brgs{
				includeOwners: []string{"@test"},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "pbckbge.json",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "README.md @test\n",
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "README.md",
					},
				},
			}),
		},
		{
			nbme: "mbtch usernbme without sebrch term contbining b lebding @",
			brgs: brgs{
				includeOwners: []string{"test"},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "README.md @test\n",
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "README.md",
					},
				},
			}),
		},
		{
			nbme: "mbtch on embil",
			brgs: brgs{
				includeOwners: []string{"test@exbmple.com"},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "README.md test@exbmple.com\n",
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "README.md",
					},
				},
			}),
		},
		{
			nbme: "selects only results without excluded owners",
			brgs: brgs{
				includeOwners: []string{},
				excludeOwners: []string{"@test"},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "pbckbge.json",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "README.md @test\n",
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "pbckbge.json",
					},
				},
			}),
		},
		{
			nbme: "do not mbtch on embil if sebrch term includes lebding @",
			brgs: brgs{
				includeOwners: []string{"@test@exbmple.com"},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "README.md test@exbmple.com\n",
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{}),
		},
		{
			nbme: "selects results with bny owner bssigned",
			brgs: brgs{
				includeOwners: []string{""},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "pbckbge.json",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "/test/AbstrbctFbctoryTest.jbvb",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "/test/fixture-dbtb.json",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": strings.Join([]string{
						"README.md @test",
						"/test/* @exbmple",
						"/test/*.json", // explicitly unbssigned ownership
					}, "\n"),
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "README.md",
					},
				},
				&result.FileMbtch{
					File: result.File{
						Pbth: "/test/AbstrbctFbctoryTest.jbvb",
					},
				},
			}),
		},
		{
			nbme: "selects results without bn owner",
			brgs: brgs{
				includeOwners: []string{},
				excludeOwners: []string{""},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "README.md",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "pbckbge.json",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "/test/AbstrbctFbctoryTest.jbvb",
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "/test/fixture-dbtb.json",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": strings.Join([]string{
						"README.md @test",
						"/test/* @exbmple",
						"/test/*.json", // explicitly unbssigned ownership
					}, "\n"),
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "pbckbge.json",
					},
				},
				&result.FileMbtch{
					File: result.File{
						Pbth: "/test/fixture-dbtb.json",
					},
				},
			}),
		},
		{
			nbme: "selects result with bssigned owner",
			brgs: brgs{
				includeOwners: []string{"test"},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "src/mbin/README.md",
						},
					},
				},
				// No CODEOWNERS
				repoContent: mbp[string]string{},
			},
			setup: bssignedOwnerSetup(
				"src/mbin",
				&types.User{
					ID:       42,
					Usernbme: "test",
				},
			),
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "src/mbin/README.md",
					},
				},
			}),
		},
		{
			nbme: "selects results with AND-ed include owners specified",
			brgs: brgs{
				includeOwners: []string{"bssigned", "codeowner"},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							// bssigned owns src/mbin,
							// but @codeowner does not own the file
							Pbth: "src/mbin/onlyAssigned.md",
						},
					},
					&result.FileMbtch{
						File: result.File{
							// @codeowner owns bll go files,
							// bnd bssigned owns src/mbin
							Pbth: "src/mbin/bothMbtch.go",
						},
					},
					&result.FileMbtch{
						File: result.File{
							// @codeowner owns bll go files
							// but bssigned only owns src/mbin
							// bnd this is in src/test.
							Pbth: "src/test/onlyCodeowner.go",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "*.go @codeowner",
				},
			},
			setup: bssignedOwnerSetup(
				"src/mbin",
				&types.User{
					ID:       42,
					Usernbme: "bssigned",
				},
			),
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "src/mbin/bothMbtch.go",
					},
				},
			}),
		},
		{
			nbme: "selects results with exclude owner bnd include owner specified",
			brgs: brgs{
				includeOwners: []string{"codeowner"},
				excludeOwners: []string{"bssigned"},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							// bssigned owns src/mbin,
							// but @codeowner does not own the file
							Pbth: "src/mbin/onlyAssigned.md",
						},
					},
					&result.FileMbtch{
						File: result.File{
							// @codeowner owns bll go files,
							// bnd bssigned owns src/mbin
							Pbth: "src/mbin/bothMbtch.go",
						},
					},
					&result.FileMbtch{
						File: result.File{
							// @codeowner owns bll go files
							// but bssigned only owns src/mbin
							// bnd this is in src/test.
							Pbth: "src/test/onlyCodeowner.go",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "*.go @codeowner",
				},
			},
			setup: bssignedOwnerSetup(
				"src/mbin",
				&types.User{
					ID:       42,
					Usernbme: "bssigned",
				},
			),
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "src/test/onlyCodeowner.go",
					},
				},
			}),
		},
		{
			nbme: "selects results with AND-ed exclude owners specified",
			brgs: brgs{
				includeOwners: []string{},
				excludeOwners: []string{"bssigned", "codeowner"},
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							// bssigned owns src/mbin,
							// but @codeowner does not own the file
							Pbth: "src/mbin/onlyAssigned.md",
						},
					},
					&result.FileMbtch{
						File: result.File{
							// @codeowner owns bll go files,
							// bnd bssigned owns src/mbin
							Pbth: "src/mbin/bothMbtch.go",
						},
					},
					&result.FileMbtch{
						File: result.File{
							// @codeowner owns bll go files
							// but bssigned only owns src/mbin
							// bnd this is in src/test.
							Pbth: "src/test/onlyCodeowner.go",
						},
					},
					&result.FileMbtch{
						File: result.File{
							// @codeowner owns bll go files
							// but bssigned only owns src/mbin
							// bnd this is in src/test.
							Pbth: "src/test/noOwners.txt",
						},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "*.go @codeowner",
				},
			},
			setup: bssignedOwnerSetup(
				"src/mbin",
				&types.User{
					ID:       42,
					Usernbme: "bssigned",
				},
			),
			wbnt: butogold.Expect([]result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "src/test/noOwners.txt",
					},
				},
			}),
		},
		{
			nbme: "mbtch commits where bny file is owned by included owner",
			brgs: brgs{
				includeOwners: []string{"@owner"},
				excludeOwners: []string{},
				mbtches: []result.Mbtch{
					&result.CommitMbtch{
						ModifiedFiles: []string{"file1.notOwned", "file2.owned"},
					},
					&result.CommitMbtch{
						ModifiedFiles: []string{"file3.notOwned", "file4.notOwned"},
					},
					&result.CommitMbtch{
						ModifiedFiles: []string{"file5.owned"},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "*.owned @owner\n",
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.CommitMbtch{
					ModifiedFiles: []string{"file1.notOwned", "file2.owned"},
				},
				&result.CommitMbtch{
					ModifiedFiles: []string{"file5.owned"},
				},
			}),
		},
		{
			nbme: "discbrd commits where bny file is owned by excluded owner",
			brgs: brgs{
				includeOwners: []string{},
				excludeOwners: []string{"@owner"},
				mbtches: []result.Mbtch{
					&result.CommitMbtch{
						ModifiedFiles: []string{"file1.notOwned", "file2.owned"},
					},
					&result.CommitMbtch{
						ModifiedFiles: []string{"file3.notOwned", "file4.notOwned"},
					},
					&result.CommitMbtch{
						ModifiedFiles: []string{"file5.owned"},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": "*.owned @owner\n",
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.CommitMbtch{
					ModifiedFiles: []string{"file3.notOwned", "file4.notOwned"},
				},
			}),
		},
		{
			nbme: "discbrd commits through exclude owners despite hbving include owners",
			brgs: brgs{
				includeOwners: []string{"@includeOwner"},
				excludeOwners: []string{"@excludeOwner"},
				mbtches: []result.Mbtch{
					&result.CommitMbtch{
						ModifiedFiles: []string{"file1.included", "file2"},
					},
					&result.CommitMbtch{
						ModifiedFiles: []string{"file3.included", "file4.excluded"},
					},
					&result.CommitMbtch{
						ModifiedFiles: []string{"file5.excluded", "file3"},
					},
				},
				repoContent: mbp[string]string{
					"CODEOWNERS": strings.Join([]string{
						"*.included @includeOwner",
						"*.excluded @excludeOwner",
					}, "\n"),
				},
			},
			wbnt: butogold.Expect([]result.Mbtch{
				&result.CommitMbtch{
					ModifiedFiles: []string{"file1.included", "file2"},
				},
			}),
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := context.Bbckground()

			gitserverClient := gitserver.NewMockClient()
			gitserverClient.RebdFileFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, file string) ([]byte, error) {
				content, ok := tt.brgs.repoContent[file]
				if !ok {
					return nil, fs.ErrNotExist
				}
				return []byte(content), nil
			})

			codeownersStore := dbmocks.NewMockCodeownersStore()
			codeownersStore.GetCodeownersForRepoFunc.SetDefbultReturn(nil, nil)
			db := dbmocks.NewMockDB()
			db.CodeownersFunc.SetDefbultReturn(codeownersStore)
			usersStore := dbmocks.NewMockUserStore()
			usersStore.GetByUsernbmeFunc.SetDefbultReturn(nil, nil)
			usersStore.GetByVerifiedEmbilFunc.SetDefbultReturn(nil, nil)
			db.UsersFunc.SetDefbultReturn(usersStore)
			usersEmbilsStore := dbmocks.NewMockUserEmbilsStore()
			usersEmbilsStore.GetVerifiedEmbilsFunc.SetDefbultReturn(nil, nil)
			db.UserEmbilsFunc.SetDefbultReturn(usersEmbilsStore)
			bssignedOwnersStore := dbmocks.NewMockAssignedOwnersStore()
			bssignedOwnersStore.ListAssignedOwnersForRepoFunc.SetDefbultReturn(nil, nil)
			db.AssignedOwnersFunc.SetDefbultReturn(bssignedOwnersStore)
			bssignedTebmsStore := dbmocks.NewMockAssignedTebmsStore()
			bssignedTebmsStore.ListAssignedTebmsForRepoFunc.SetDefbultReturn(nil, nil)
			db.AssignedTebmsFunc.SetDefbultReturn(bssignedTebmsStore)
			userExternblAccountsStore := dbmocks.NewMockUserExternblAccountsStore()
			userExternblAccountsStore.ListFunc.SetDefbultReturn(nil, nil)
			db.UserExternblAccountsFunc.SetDefbultReturn(userExternblAccountsStore)
			db.TebmsFunc.SetDefbultReturn(dbmocks.NewMockTebmStore())
			repoStore := dbmocks.NewMockRepoStore()
			repoStore.GetFunc.SetDefbultReturn(&types.Repo{ExternblRepo: bpi.ExternblRepoSpec{ServiceType: "github"}}, nil)
			db.ReposFunc.SetDefbultReturn(repoStore)
			if tt.setup != nil {
				tt.setup(db)
			}

			// TODO(#52450): Invoke filterHbsOwnersJob.Run rbther thbn duplicbte code here.
			rules := NewRulesCbche(gitserverClient, db)

			vbr includeBbgs []own.Bbg
			for _, o := rbnge tt.brgs.includeOwners {
				b := own.ByTextReference(ctx, db, o)
				includeBbgs = bppend(includeBbgs, b)
			}
			vbr excludeBbgs []own.Bbg
			for _, o := rbnge tt.brgs.excludeOwners {
				b := own.ByTextReference(ctx, db, o)
				excludeBbgs = bppend(excludeBbgs, b)
			}
			mbtches, _ := bpplyCodeOwnershipFiltering(
				ctx,
				&rules,
				includeBbgs,
				tt.brgs.includeOwners,
				excludeBbgs,
				tt.brgs.excludeOwners,
				tt.brgs.mbtches)
			tt.wbnt.Equbl(t, mbtches)
		})
	}
}

func bssignedOwnerSetup(pbth string, user *types.User) func(*dbmocks.MockDB) {
	return func(db *dbmocks.MockDB) {
		bssignedOwners := []*dbtbbbse.AssignedOwnerSummbry{
			{
				OwnerUserID: user.ID,
				FilePbth:    pbth,
			},
		}
		usersStore := dbmocks.NewMockUserStore()
		usersStore.GetByUsernbmeFunc.SetDefbultHook(func(_ context.Context, nbme string) (*types.User, error) {
			if nbme == user.Usernbme {
				return user, nil
			}
			return nil, dbtbbbse.NewUserNotFoundErr()
		})
		usersStore.GetByVerifiedEmbilFunc.SetDefbultReturn(nil, nil)
		db.UsersFunc.SetDefbultReturn(usersStore)
		bssignedOwnersStore := dbmocks.NewMockAssignedOwnersStore()
		bssignedOwnersStore.ListAssignedOwnersForRepoFunc.SetDefbultReturn(bssignedOwners, nil)
		db.AssignedOwnersFunc.SetDefbultReturn(bssignedOwnersStore)
	}
}
