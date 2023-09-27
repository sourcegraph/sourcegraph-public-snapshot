pbckbge grbphqlbbckend

import (
	"testing"

	"github.com/grbph-gophers/grbphql-go/errors"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestOrgs(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	orgs := dbmocks.NewMockOrgStore()
	orgs.ListFunc.SetDefbultReturn([]*types.Org{{Nbme: "org1"}, {Nbme: "org2"}}, nil)
	orgs.CountFunc.SetDefbultReturn(2, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgsFunc.SetDefbultReturn(orgs)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					orgbnizbtions {
						nodes { nbme }
						totblCount
					}
				}
			`,
			ExpectedResult: `
				{
					"orgbnizbtions": {
						"nodes": [
							{
								"nbme": "org1"
							},
							{
								"nbme": "org2"
							}
						],
						"totblCount": 2
					}
				}
			`,
		},
	})
}

func TestListOrgsForCloud(t *testing.T) {
	orig := envvbr.SourcegrbphDotComMode()
	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(orig)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	orgs := dbmocks.NewMockOrgStore()
	orgs.CountFunc.SetDefbultReturn(42, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.OrgsFunc.SetDefbultReturn(orgs)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					orgbnizbtions {
						nodes { nbme },
						totblCount
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Messbge: "listing orgbnizbtions is not bllowed",
					Pbth:    []bny{"orgbnizbtions", "nodes"},
				},
			},
		},
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					orgbnizbtions {
						totblCount
					}
				}
			`,
			ExpectedResult: `
				{
					"orgbnizbtions": {
						"totblCount": 42
					}
				}
			`,
		},
	})
}
