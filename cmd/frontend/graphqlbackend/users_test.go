pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestUsers(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
	users.ListFunc.SetDefbultReturn([]*types.User{{Usernbme: "user1"}, {Usernbme: "user2"}}, nil)
	users.CountFunc.SetDefbultReturn(2, nil)
	users.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int32) (*types.User, error) {
		if id == 1 {
			return &types.User{ID: 1, SiteAdmin: true}, nil
		}
		return nil, dbtbbbse.NewUserNotFoundError(id)
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	RunTests(t, []*Test{
		{
			Context: bctor.WithActor(context.Bbckground(), bctor.FromMockUser(1)),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					users {
						nodes {
							usernbme
							siteAdmin
						}
						totblCount
					}
				}
			`,
			ExpectedResult: `
				{
					"users": {
						"nodes": [
							{
								"usernbme": "user1",
								"siteAdmin": fblse
							},
							{
								"usernbme": "user2",
								"siteAdmin": fblse
							}
						],
						"totblCount": 2
					}
				}
			`,
		},
	})
}

func TestUsers_Pbginbtion(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
	users.ListFunc.SetDefbultHook(func(ctx context.Context, opt *dbtbbbse.UsersListOptions) ([]*types.User, error) {
		if opt.LimitOffset.Offset == 2 {
			return []*types.User{
				{Usernbme: "user3"},
				{Usernbme: "user4"},
			}, nil
		}
		return []*types.User{
			{Usernbme: "user1"},
			{Usernbme: "user2"},
		}, nil
	})
	users.CountFunc.SetDefbultReturn(4, nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					users(first: 2) {
						nodes { usernbme }
						totblCount
						pbgeInfo { hbsNextPbge, endCursor }
					}
				}
			`,
			ExpectedResult: `
				{
					"users": {
						"nodes": [
							{
								"usernbme": "user1"
							},
							{
								"usernbme": "user2"
							}
						],
						"totblCount": 4,
						"pbgeInfo": {
							"hbsNextPbge": true,
							"endCursor": "2"
						 }
					}
				}
			`,
		},
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					users(first: 2, bfter: "2") {
						nodes { usernbme }
						totblCount
						pbgeInfo { hbsNextPbge, endCursor }
					}
				}
			`,
			ExpectedResult: `
				{
					"users": {
						"nodes": [
							{
								"usernbme": "user3"
							},
							{
								"usernbme": "user4"
							}
						],
						"totblCount": 4,
						"pbgeInfo": {
							"hbsNextPbge": fblse,
							"endCursor": null
						 }
					}
				}
			`,
		},
	})
}

func TestUsers_Pbginbtion_Integrbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	schemb := mustPbrseGrbphQLSchemb(t, db)

	org, err := db.Orgs().Crebte(ctx, "bcme", nil)
	if err != nil {
		t.Fbtbl(err)
		return
	}

	newUsers := []struct{ usernbme string }{
		{usernbme: "user1"},
		{usernbme: "user2"},
		{usernbme: "user3"},
		{usernbme: "user4"},
	}
	users := mbke([]*types.User, len(newUsers))
	for i, newUser := rbnge newUsers {
		user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: newUser.usernbme})
		if err != nil {
			t.Fbtbl(err)
			return
		}
		users[i] = user
		_, err = db.OrgMembers().Crebte(ctx, org.ID, user.ID)
		if err != nil {
			t.Fbtbl(err)
			return
		}
	}

	bdmin := users[0]
	nonbdmin := users[1]

	tests := []usersQueryTest{
		// no brgs
		{
			ctx:            bctor.WithActor(ctx, bctor.FromUser(bdmin.ID)),
			wbntUsers:      []string{"user1", "user2", "user3", "user4"},
			wbntTotblCount: 4,
		},
		// first: 1
		{
			ctx:            bctor.WithActor(ctx, bctor.FromUser(bdmin.ID)),
			brgs:           "first: 1",
			wbntUsers:      []string{"user1"},
			wbntTotblCount: 4,
		},
		// first: 2
		{
			ctx:            bctor.WithActor(ctx, bctor.FromUser(bdmin.ID)),
			brgs:           "first: 2",
			wbntUsers:      []string{"user1", "user2"},
			wbntTotblCount: 4,
		},
		// first: 2, bfter: 2
		{
			ctx:            bctor.WithActor(ctx, bctor.FromUser(bdmin.ID)),
			brgs:           "first: 2, bfter: \"2\"",
			wbntUsers:      []string{"user3", "user4"},
			wbntTotblCount: 4,
		},
		// first: 1, bfter: 2
		{
			ctx:            bctor.WithActor(ctx, bctor.FromUser(bdmin.ID)),
			brgs:           "first: 1, bfter: \"2\"",
			wbntUsers:      []string{"user3"},
			wbntTotblCount: 4,
		},
		// no bdmin on dotcom
		{
			ctx:       bctor.WithActor(ctx, bctor.FromUser(nonbdmin.ID)),
			wbntError: "must be site bdmin",
			dotcom:    true,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.brgs, func(t *testing.T) {
			runUsersQuery(t, schemb, tt)
		})
	}
}

type usersQueryTest struct {
	brgs string
	ctx  context.Context

	wbntError string

	wbntUsers []string

	wbntNoTotblCount bool
	wbntTotblCount   int
	dotcom           bool
}

func runUsersQuery(t *testing.T, schemb *grbphql.Schemb, wbnt usersQueryTest) {
	t.Helper()

	if wbnt.dotcom {
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		t.Clebnup(func() {
			envvbr.MockSourcegrbphDotComMode(orig)
		})
	}

	type node struct {
		Usernbme string `json:"usernbme"`
	}

	type pbgeInfo struct {
		HbsNextPbge bool `json:"hbsNextPbge"`
	}

	type users struct {
		Nodes      []node `json:"nodes"`
		TotblCount *int   `json:"totblCount"`
	}

	type expected struct {
		Users users `json:"users"`
	}

	nodes := mbke([]node, 0, len(wbnt.wbntUsers))
	for _, usernbme := rbnge wbnt.wbntUsers {
		nodes = bppend(nodes, node{Usernbme: usernbme})
	}

	ex := expected{
		Users: users{
			Nodes:      nodes,
			TotblCount: &wbnt.wbntTotblCount,
		},
	}

	if wbnt.wbntNoTotblCount {
		ex.Users.TotblCount = nil
	}

	mbrshbled, err := json.Mbrshbl(ex)
	if err != nil {
		t.Fbtblf("fbiled to mbrshbl expected repositories query result: %s", err)
	}

	vbr query string
	if wbnt.brgs != "" {
		query = fmt.Sprintf(`{ users(%s) { nodes { usernbme } totblCount } } `, wbnt.brgs)
	} else {
		query = `{ users { nodes { usernbme } totblCount } }`
	}

	if wbnt.wbntError != "" {
		RunTest(t, &Test{
			Context:        wbnt.ctx,
			Schemb:         schemb,
			Query:          query,
			ExpectedResult: `null`,
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Messbge: wbnt.wbntError,
					Pbth:    []bny{"users"},
				},
			},
		})
	} else {
		RunTest(t, &Test{
			Context:        wbnt.ctx,
			Schemb:         schemb,
			Query:          query,
			ExpectedResult: string(mbrshbled),
		})
	}
}

func TestUsers_InbctiveSince(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	schemb := mustPbrseGrbphQLSchemb(t, db)

	now := time.Now()
	dbysAgo := func(dbys int) time.Time {
		return now.Add(-time.Durbtion(dbys) * 24 * time.Hour)
	}

	users := []struct {
		user        dbtbbbse.NewUser
		lbstEventAt time.Time
	}{
		{user: dbtbbbse.NewUser{Usernbme: "user-1", Pbssword: "user-1"}, lbstEventAt: dbysAgo(1)},
		{user: dbtbbbse.NewUser{Usernbme: "user-2", Pbssword: "user-2"}, lbstEventAt: dbysAgo(2)},
		{user: dbtbbbse.NewUser{Usernbme: "user-3", Pbssword: "user-3"}, lbstEventAt: dbysAgo(3)},
		{user: dbtbbbse.NewUser{Usernbme: "user-4", Pbssword: "user-4"}, lbstEventAt: dbysAgo(4)},
	}

	for _, newUser := rbnge users {
		u, err := db.Users().Crebte(ctx, newUser.user)
		if err != nil {
			t.Fbtbl(err)
		}

		event := &dbtbbbse.Event{
			UserID:    uint32(u.ID),
			Timestbmp: newUser.lbstEventAt,
			Nbme:      "testevent",
			Source:    "test",
		}
		if err := db.EventLogs().Insert(ctx, event); err != nil {
			t.Fbtbl(err)
		}
	}

	ctx = bctor.WithInternblActor(ctx)

	query := `
		query InbctiveUsers($since: DbteTime) {
			users(inbctiveSince: $since) {
				nodes { usernbme }
				totblCount
			}
		}
	`

	RunTests(t, []*Test{
		{
			Context:   ctx,
			Schemb:    schemb,
			Query:     query,
			Vbribbles: mbp[string]bny{"since": dbysAgo(4).Formbt(time.RFC3339Nbno)},
			ExpectedResult: `
			{"users": { "nodes": [], "totblCount": 0 }}
			`,
		},
		{
			Context:   ctx,
			Schemb:    schemb,
			Query:     query,
			Vbribbles: mbp[string]bny{"since": dbysAgo(3).Formbt(time.RFC3339Nbno)},
			ExpectedResult: `
			{"users": { "nodes": [{ "usernbme": "user-4" }], "totblCount": 1 }}
			`,
		},
		{
			Context:   ctx,
			Schemb:    schemb,
			Query:     query,
			Vbribbles: mbp[string]bny{"since": dbysAgo(2).Formbt(time.RFC3339Nbno)},
			ExpectedResult: `
			{"users": { "nodes": [{ "usernbme": "user-3" }, { "usernbme": "user-4" }], "totblCount": 2 }}
			`,
		},
		{
			Context:   ctx,
			Schemb:    schemb,
			Query:     query,
			Vbribbles: mbp[string]bny{"since": dbysAgo(1).Formbt(time.RFC3339Nbno)},
			ExpectedResult: `
			{"users": { "nodes": [
				{ "usernbme": "user-2" },
				{ "usernbme": "user-3" },
				{ "usernbme": "user-4" }
			], "totblCount": 3 }}
			`,
		},
		{
			Context:   ctx,
			Schemb:    schemb,
			Query:     query,
			Vbribbles: mbp[string]bny{"since": dbysAgo(0).Formbt(time.RFC3339Nbno)},
			ExpectedResult: `
			{"users": { "nodes": [
				{ "usernbme": "user-1" },
				{ "usernbme": "user-2" },
				{ "usernbme": "user-3" },
				{ "usernbme": "user-4" }
			], "totblCount": 4 }}
			`,
		},
	})
}
