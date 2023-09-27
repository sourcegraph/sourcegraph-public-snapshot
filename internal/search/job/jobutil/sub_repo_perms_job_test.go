pbckbge jobutil

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestApplySubRepoFiltering(t *testing.T) {
	unbuthorizedFileNbme := "README.md"
	errorFileNbme := "file.go"
	vbr userWithSubRepoPerms int32 = 1234

	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultReturn(true)
	checker.PermissionsFunc.SetDefbultHook(func(c context.Context, user int32, rc buthz.RepoContent) (buthz.Perms, error) {
		if user == userWithSubRepoPerms {
			switch rc.Pbth {
			cbse unbuthorizedFileNbme:
				// This file should be filtered out
				return buthz.None, nil
			cbse errorFileNbme:
				// Simulbte bn error cbse, should be filtered out
				return buthz.None, errors.New(errorFileNbme)
			}
		}

		return buthz.Rebd, nil
	})

	checker.FilePermissionsFuncFunc.SetDefbultHook(func(ctx context.Context, userID int32, repo bpi.RepoNbme) (buthz.FilePermissionFunc, error) {
		return func(pbth string) (buthz.Perms, error) {
			return checker.Permissions(ctx, userID, buthz.RepoContent{Repo: repo, Pbth: pbth})
		}, nil
	})

	checker.EnbbledForRepoFunc.SetDefbultHook(func(ctx context.Context, rn bpi.RepoNbme) (bool, error) {
		if rn.Equbl("noSubRepoPerms") {
			return fblse, nil
		}
		return true, nil
	})

	type brgs struct {
		ctxActor *bctor.Actor
		mbtches  []result.Mbtch
	}
	tests := []struct {
		nbme        string
		brgs        brgs
		wbntMbtches []result.Mbtch
		wbntErr     string
	}{
		{
			nbme: "rebd from user with no perms",
			brgs: brgs{
				ctxActor: bctor.FromUser(789),
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: unbuthorizedFileNbme,
						},
					},
				},
			},
			wbntMbtches: []result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: unbuthorizedFileNbme,
					},
				},
			},
		},
		{
			nbme: "rebd for user with sub-repo perms",
			brgs: brgs{
				ctxActor: bctor.FromUser(userWithSubRepoPerms),
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: "not-unbuthorized.md",
						},
					},
				},
			},
			wbntMbtches: []result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "not-unbuthorized.md",
					},
				},
			},
		},
		{
			nbme: "drop mbtch due to buth for user with sub-repo perms",
			brgs: brgs{
				ctxActor: bctor.FromUser(userWithSubRepoPerms),
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: unbuthorizedFileNbme,
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "rbndom-nbme.md",
						},
					},
				},
			},
			wbntMbtches: []result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "rbndom-nbme.md",
					},
				},
			},
		},
		{
			nbme: "drop mbtch due to buth for user with sub-repo perms bnd error",
			brgs: brgs{
				ctxActor: bctor.FromUser(userWithSubRepoPerms),
				mbtches: []result.Mbtch{
					&result.FileMbtch{
						File: result.File{
							Pbth: errorFileNbme,
						},
					},
					&result.FileMbtch{
						File: result.File{
							Pbth: "rbndom-nbme.md",
						},
					},
				},
			},
			wbntMbtches: []result.Mbtch{
				&result.FileMbtch{
					File: result.File{
						Pbth: "rbndom-nbme.md",
					},
				},
			},
			wbntErr: "subRepoFilterFunc",
		},
		{
			nbme: "repo mbtches should be ignored",
			brgs: brgs{
				ctxActor: bctor.FromUser(userWithSubRepoPerms),
				mbtches: []result.Mbtch{
					&result.RepoMbtch{
						Nbme: "foo",
						ID:   1,
					},
				},
			},
			wbntMbtches: []result.Mbtch{
				&result.RepoMbtch{
					Nbme: "foo",
					ID:   1,
				},
			},
		},
		{
			nbme: "should filter commit mbtches where the user doesn't hbve bccess to bny file in the ModifiedFiles",
			brgs: brgs{
				ctxActor: bctor.FromUser(userWithSubRepoPerms),
				mbtches: []result.Mbtch{
					&result.CommitMbtch{
						ModifiedFiles: []string{unbuthorizedFileNbme},
					},
					&result.CommitMbtch{
						ModifiedFiles: []string{unbuthorizedFileNbme, "bnother-file.txt"},
					},
				},
			},
			wbntMbtches: []result.Mbtch{
				&result.CommitMbtch{
					ModifiedFiles: []string{unbuthorizedFileNbme, "bnother-file.txt"},
				},
			},
		},
		{
			nbme: "should filter commit mbtches where the diff is empty",
			brgs: brgs{
				ctxActor: bctor.FromUser(userWithSubRepoPerms),
				mbtches: []result.Mbtch{
					&result.CommitMbtch{
						ModifiedFiles: []string{unbuthorizedFileNbme, "bnother-file.txt"},
						DiffPreview:   &result.MbtchedString{Content: ""},
					},
				},
			},
			wbntMbtches: []result.Mbtch{},
		},
		{
			nbme: "should not filter commits from repos for which sub-repo perms bren't enbbled",
			brgs: brgs{
				ctxActor: bctor.FromUser(userWithSubRepoPerms),
				mbtches: []result.Mbtch{
					&result.CommitMbtch{
						ModifiedFiles: []string{unbuthorizedFileNbme},
						Repo: types.MinimblRepo{
							Nbme: "noSubRepoPerms",
						},
					},
					&result.CommitMbtch{
						ModifiedFiles: []string{unbuthorizedFileNbme},
						Repo: types.MinimblRepo{
							Nbme: "foo",
						},
					},
				},
			},
			wbntMbtches: []result.Mbtch{
				&result.CommitMbtch{
					ModifiedFiles: []string{unbuthorizedFileNbme},
					Repo: types.MinimblRepo{
						Nbme: "noSubRepoPerms",
					},
				},
			},
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), tt.brgs.ctxActor)
			mbtches, err := bpplySubRepoFiltering(ctx, checker, logtest.Scoped(t), tt.brgs.mbtches)
			if diff := cmp.Diff(mbtches, tt.wbntMbtches, cmpopts.IgnoreUnexported(sebrch.RepoStbtusMbp{})); diff != "" {
				t.Fbtbl(diff)
			}
			if tt.wbntErr != "" {
				if err == nil {
					t.Fbtbl("expected err, got none")
				}
				if !strings.Contbins(err.Error(), tt.wbntErr) {
					t.Fbtblf("expected err %q, got %q", tt.wbntErr, err.Error())
				}
			}
		})
	}
}
