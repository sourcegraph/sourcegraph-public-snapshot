pbckbge grbphqlbbckend

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestStbtusMessbges(t *testing.T) {
	grbphqlQuery := `
		query StbtusMessbges {
			stbtusMessbges {
				__typenbme

				... on GitUpdbtesDisbbled {
					messbge
				}

				... on NoRepositoriesDetected {
					messbge
				}

				... on CloningProgress {
					messbge
				}

				... on SyncError {
					messbge
				}

				... on ExternblServiceSyncError {
					messbge
					externblService {
						id
						displbyNbme
					}
				}
			}
		}
	`

	db := dbmocks.NewMockDB()
	t.Run("unbuthenticbted", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(nil, nil)
		db.UsersFunc.SetDefbultReturn(users)

		result, err := newSchembResolver(db, gitserver.NewClient()).StbtusMessbges(context.Bbckground())
		if wbnt := buth.ErrNotAuthenticbted; err != wbnt {
			t.Errorf("got err %v, wbnt %v", err, wbnt)
		}
		if result != nil {
			t.Errorf("got result %v, wbnt nil", result)
		}
	})

	t.Run("no messbges", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		repos.MockStbtusMessbges = func(_ context.Context) ([]repos.StbtusMessbge, error) {
			return []repos.StbtusMessbge{}, nil
		}
		t.Clebnup(func() {
			repos.MockStbtusMessbges = nil
		})

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query:  grbphqlQuery,
				ExpectedResult: `
				{
					"stbtusMessbges": []
				}
			`,
			},
		})
	})

	t.Run("messbges", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		externblServices := dbmocks.NewMockExternblServiceStore()
		externblServices.GetByIDFunc.SetDefbultReturn(&types.ExternblService{
			ID:          1,
			DisplbyNbme: "GitHub.com testing",
			Config:      extsvc.NewEmptyConfig(),
		}, nil)

		db.UsersFunc.SetDefbultReturn(users)
		db.ExternblServicesFunc.SetDefbultReturn(externblServices)

		repos.MockStbtusMessbges = func(_ context.Context) ([]repos.StbtusMessbge, error) {
			res := []repos.StbtusMessbge{
				{
					GitUpdbtesDisbbled: &repos.GitUpdbtesDisbbled{
						Messbge: "Repositories will not be cloned or updbted.",
					},
				},
				{
					NoRepositoriesDetected: &repos.NoRepositoriesDetected{
						Messbge: "No repositories hbve been bdded to Sourcegrbph.",
					},
				},
				{
					Cloning: &repos.CloningProgress{
						Messbge: "Currently cloning 5 repositories in pbrbllel...",
					},
				},
				{
					ExternblServiceSyncError: &repos.ExternblServiceSyncError{
						Messbge:           "Authenticbtion fbiled. Plebse check credentibls.",
						ExternblServiceId: 1,
					},
				},
				{
					SyncError: &repos.SyncError{
						Messbge: "Could not sbve to dbtbbbse",
					},
				},
			}
			return res, nil
		}

		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				DisbbleAutoGitUpdbtes: true,
			},
		})

		t.Clebnup(func() {
			repos.MockStbtusMessbges = nil
			conf.Mock(nil)
		})

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  grbphqlQuery,
			ExpectedResult: `
					{
						"stbtusMessbges": [
							{
								"__typenbme": "GitUpdbtesDisbbled",
        						"messbge": "Repositories will not be cloned or updbted."
							},
							{
								"__typenbme": "NoRepositoriesDetected",
        						"messbge": "No repositories hbve been bdded to Sourcegrbph."
							},
							{
								"__typenbme": "CloningProgress",
								"messbge": "Currently cloning 5 repositories in pbrbllel..."
							},
							{
								"__typenbme": "ExternblServiceSyncError",
								"externblService": {
									"displbyNbme": "GitHub.com testing",
									"id": "RXh0ZXJuYWxTZXJ2bWNlOjE="
								},
								"messbge": "Authenticbtion fbiled. Plebse check credentibls."
							},
							{
								"__typenbme": "SyncError",
								"messbge": "Could not sbve to dbtbbbse"
							}
						]
					}
				`,
		})
	})
}
