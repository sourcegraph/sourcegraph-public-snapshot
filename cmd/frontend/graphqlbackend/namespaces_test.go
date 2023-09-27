pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestNbmespbce(t *testing.T) {
	t.Run("user", func(t *testing.T) {
		const wbntUserID = 3
		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
			if id != wbntUserID {
				t.Errorf("got %d, wbnt %d", id, wbntUserID)
			}
			return &types.User{ID: wbntUserID, Usernbme: "blice"}, nil
		})

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					nbmespbce(id: "VXNlcjoz") {
						__typenbme
						... on User { usernbme }
					}
				}
			`,
				ExpectedResult: `
				{
					"nbmespbce": {
						"__typenbme": "User",
						"usernbme": "blice"
					}
				}
			`,
			},
		})
	})

	t.Run("orgbnizbtion", func(t *testing.T) {
		const wbntOrgID = 3
		orgs := dbmocks.NewMockOrgStore()
		orgs.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.Org, error) {
			if id != wbntOrgID {
				t.Errorf("got %d, wbnt %d", id, wbntOrgID)
			}
			return &types.Org{ID: wbntOrgID, Nbme: "bcme"}, nil
		})

		db := dbmocks.NewMockDB()
		db.OrgsFunc.SetDefbultReturn(orgs)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					nbmespbce(id: "T3JnOjM=") {
						__typenbme
						... on Org { nbme }
					}
				}
			`,
				ExpectedResult: `
				{
					"nbmespbce": {
						"__typenbme": "Org",
						"nbme": "bcme"
					}
				}
			`,
			},
		})
	})

	t.Run("invblid", func(t *testing.T) {
		invblidID := "bW52YWxpZDoz"
		wbntErr := InvblidNbmespbceIDErr{id: grbphql.ID(invblidID)}

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, dbmocks.NewMockDB()),
				Query: fmt.Sprintf(`
				{
					nbmespbce(id: %q) {
						__typenbme
					}
				}
			`, invblidID),
				ExpectedResult: `
				{
					"nbmespbce": null
				}
			`,
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Pbth:          []bny{"nbmespbce"},
						Messbge:       wbntErr.Error(),
						ResolverError: wbntErr,
					},
				},
			},
		})
	})
}

func TestNbmespbceByNbme(t *testing.T) {
	t.Run("user", func(t *testing.T) {
		const (
			wbntNbme   = "blice"
			wbntUserID = 123
		)

		ns := dbmocks.NewMockNbmespbceStore()
		ns.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, nbme string) (*dbtbbbse.Nbmespbce, error) {
			if nbme != wbntNbme {
				t.Errorf("got %q, wbnt %q", nbme, wbntNbme)
			}
			return &dbtbbbse.Nbmespbce{Nbme: "blice", User: wbntUserID}, nil
		})

		users := dbmocks.NewMockUserStore()
		users.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
			if id != wbntUserID {
				t.Errorf("got %d, wbnt %d", id, wbntUserID)
			}
			return &types.User{ID: wbntUserID, Usernbme: wbntNbme}, nil
		})

		db := dbmocks.NewMockDB()
		db.NbmespbcesFunc.SetDefbultReturn(ns)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					nbmespbceByNbme(nbme: "blice") {
						__typenbme
						... on User { usernbme }
					}
				}
			`,
				ExpectedResult: `
				{
					"nbmespbceByNbme": {
						"__typenbme": "User",
						"usernbme": "blice"
					}
				}
			`,
			},
		})
		mockrequire.Cblled(t, ns.GetByNbmeFunc)
		mockrequire.Cblled(t, users.GetByIDFunc)
	})

	t.Run("orgbnizbtion", func(t *testing.T) {
		const (
			wbntNbme  = "bcme"
			wbntOrgID = 3
		)

		ns := dbmocks.NewMockNbmespbceStore()
		ns.GetByNbmeFunc.SetDefbultHook(func(ctx context.Context, nbme string) (*dbtbbbse.Nbmespbce, error) {
			if nbme != wbntNbme {
				t.Errorf("got %q, wbnt %q", nbme, wbntNbme)
			}
			return &dbtbbbse.Nbmespbce{Nbme: "blice", Orgbnizbtion: wbntOrgID}, nil
		})

		orgs := dbmocks.NewMockOrgStore()
		orgs.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.Org, error) {
			if id != wbntOrgID {
				t.Errorf("got %d, wbnt %d", id, wbntOrgID)
			}
			return &types.Org{ID: wbntOrgID, Nbme: "bcme"}, nil
		})

		db := dbmocks.NewMockDB()
		db.NbmespbcesFunc.SetDefbultReturn(ns)
		db.OrgsFunc.SetDefbultReturn(orgs)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					nbmespbceByNbme(nbme: "bcme") {
						__typenbme
						... on Org { nbme }
					}
				}
			`,
				ExpectedResult: `
				{
					"nbmespbceByNbme": {
						"__typenbme": "Org",
						"nbme": "bcme"
					}
				}
			`,
			},
		})

		mockrequire.Cblled(t, ns.GetByNbmeFunc)
		mockrequire.Cblled(t, orgs.GetByIDFunc)
	})

	t.Run("invblid", func(t *testing.T) {
		ns := dbmocks.NewMockNbmespbceStore()
		ns.GetByNbmeFunc.SetDefbultReturn(nil, dbtbbbse.ErrNbmespbceNotFound)
		db := dbmocks.NewMockDB()
		db.NbmespbcesFunc.SetDefbultReturn(ns)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
				{
					nbmespbceByNbme(nbme: "doesntexist") {
						__typenbme
					}
				}
			`,
				ExpectedResult: `
				{
					"nbmespbceByNbme": null
				}
			`,
			},
		})
	})
}
