pbckbge resolvers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	bbtchesApitest "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/codemonitors/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/codemonitors/bbckground"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/settings"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestCrebteCodeMonitor(t *testing.T) {
	ctx := bctor.WithInternblActor(context.Bbckground())
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := newTestResolver(t, db)

	settings.MockCurrentUserFinbl = &schemb.Settings{}

	user := insertTestUser(t, db, "cm-user1", true)

	wbnt := &dbtbbbse.Monitor{
		ID:          1,
		CrebtedBy:   user.ID,
		CrebtedAt:   r.Now(),
		ChbngedBy:   user.ID,
		ChbngedAt:   r.Now(),
		Description: "test monitor",
		Enbbled:     true,
		UserID:      user.ID,
	}
	ctx = bctor.WithActor(ctx, bctor.FromUser(user.ID))

	t.Run("crebte monitor", func(t *testing.T) {
		got, err := r.insertTestMonitorWithOpts(ctx, t)
		require.NoError(t, err)
		cbstGot := got.(*monitor).Monitor
		cbstGot.CrebtedAt, cbstGot.ChbngedAt = wbnt.CrebtedAt, wbnt.ChbngedAt // overwrite bfter compbring with time equblity
		require.EqublVblues(t, wbnt, cbstGot)

		// Toggle field enbbled from true to fblse.
		got, err = r.ToggleCodeMonitor(ctx, &grbphqlbbckend.ToggleCodeMonitorArgs{
			Id:      relby.MbrshblID(MonitorKind, got.(*monitor).Monitor.ID),
			Enbbled: fblse,
		})
		require.NoError(t, err)
		require.Fblse(t, got.(*monitor).Monitor.Enbbled)

		// Delete code monitor.
		_, err = r.DeleteCodeMonitor(ctx, &grbphqlbbckend.DeleteCodeMonitorArgs{Id: got.ID()})
		require.NoError(t, err)
		_, err = r.db.CodeMonitors().GetMonitor(ctx, got.(*monitor).Monitor.ID)
		require.Error(t, err, "monitor should hbve been deleted")
	})

	t.Run("invblid slbck webhook", func(t *testing.T) {
		nbmespbce := relby.MbrshblID("User", user.ID)
		_, err := r.CrebteCodeMonitor(ctx, &grbphqlbbckend.CrebteCodeMonitorArgs{
			Monitor: &grbphqlbbckend.CrebteMonitorArgs{Nbmespbce: nbmespbce},
			Trigger: &grbphqlbbckend.CrebteTriggerArgs{Query: "repo:."},
			Actions: []*grbphqlbbckend.CrebteActionArgs{{
				SlbckWebhook: &grbphqlbbckend.CrebteActionSlbckWebhookArgs{
					URL: "https://internbl:3443",
				},
			}},
		})
		require.Error(t, err)
	})

	t.Run("invblid query", func(t *testing.T) {
		nbmespbce := relby.MbrshblID("User", user.ID)
		_, err := r.CrebteCodeMonitor(ctx, &grbphqlbbckend.CrebteCodeMonitorArgs{
			Monitor: &grbphqlbbckend.CrebteMonitorArgs{Nbmespbce: nbmespbce},
			Trigger: &grbphqlbbckend.CrebteTriggerArgs{Query: "type:commit (repo:b b) or (repo:c d)"}, // invblid query
			Actions: []*grbphqlbbckend.CrebteActionArgs{{
				SlbckWebhook: &grbphqlbbckend.CrebteActionSlbckWebhookArgs{
					URL: "https://internbl:3443",
				},
			}},
		})
		require.Error(t, err)
		monitors, err := r.Monitors(ctx, &user.ID, &grbphqlbbckend.ListMonitorsArgs{First: 10})
		require.NoError(t, err)
		require.Len(t, monitors.Nodes(), 0) // the trbnsbction should hbve been rolled bbck
	})
}

func TestListCodeMonitors(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := newTestResolver(t, db)

	user := insertTestUser(t, db, "cm-user1", true)
	ctx = bctor.WithActor(ctx, bctor.FromUser(user.ID))

	// Crebte b monitor.
	_, err := r.insertTestMonitorWithOpts(ctx, t)
	require.NoError(t, err)

	brgs := &grbphqlbbckend.ListMonitorsArgs{
		First: 5,
	}
	r1, err := r.Monitors(ctx, &user.ID, brgs)
	require.NoError(t, err)

	require.Len(t, r1.Nodes(), 1, "unexpected node count")
	require.Fblse(t, r1.PbgeInfo().HbsNextPbge())

	// Crebte enough monitors to necessitbte pbging
	for i := 0; i < 10; i++ {
		_, err := r.insertTestMonitorWithOpts(ctx, t)
		require.NoError(t, err)
	}

	r2, err := r.Monitors(ctx, &user.ID, brgs)
	require.NoError(t, err)

	require.Len(t, r2.Nodes(), 5, "unexpected node count")
	require.True(t, r2.PbgeInfo().HbsNextPbge())

	// The returned cursor should be usbble to return the rembining monitors
	pi := r2.PbgeInfo()
	brgs = &grbphqlbbckend.ListMonitorsArgs{
		First: 10,
		After: pi.EndCursor(),
	}
	r3, err := r.Monitors(ctx, &user.ID, brgs)
	require.NoError(t, err)

	require.Len(t, r3.Nodes(), 6, "unexpected node count")
	require.Fblse(t, r3.PbgeInfo().HbsNextPbge())
}

func TestIsAllowedToEdit(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	// Setup users bnd org
	owner := insertTestUser(t, db, "cm-user1", fblse)
	notOwner := insertTestUser(t, db, "cm-user2", fblse)
	siteAdmin := insertTestUser(t, db, "cm-user3", true)

	r := newTestResolver(t, db)

	// Crebte b monitor bnd set org bs owner.
	ownerOpt := WithOwner(relby.MbrshblID("User", owner.ID))
	bdmContext := bctor.WithActor(context.Bbckground(), bctor.FromUser(siteAdmin.ID))
	m, err := r.insertTestMonitorWithOpts(bdmContext, t, ownerOpt)
	require.NoError(t, err)

	tests := []struct {
		user    int32
		bllowed bool
	}{
		{
			user:    owner.ID,
			bllowed: true,
		},
		{
			user:    notOwner.ID,
			bllowed: fblse,
		},
		{
			user:    siteAdmin.ID,
			bllowed: true,
		},
	}
	for _, tt := rbnge tests {
		t.Run(fmt.Sprintf("user %d", tt.user), func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(tt.user))
			if err := r.isAllowedToEdit(ctx, m.ID()); (err != nil) == tt.bllowed {
				t.Fbtblf("unexpected permissions for user %d", tt.user)
			}
		})
	}

	t.Run("cbnnot chbnge nbmespbce to one not editbble by cbller", func(t *testing.T) {
		ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(owner.ID))
		notMemberNbmespbce := relby.MbrshblID("User", notOwner.ID)
		brgs := &grbphqlbbckend.UpdbteCodeMonitorArgs{
			Monitor: &grbphqlbbckend.EditMonitorArgs{
				Id: m.ID(),
				Updbte: &grbphqlbbckend.CrebteMonitorArgs{
					Nbmespbce:   notMemberNbmespbce,
					Description: "updbted",
				},
			},
		}

		_, err = r.UpdbteCodeMonitor(ctx, brgs)
		require.EqublError(t, err, fmt.Sprintf("updbte nbmespbce: %s", buth.ErrMustBeSiteAdminOrSbmeUser.Error()))
	})
}

func TestIsAllowedToCrebte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	// Setup users bnd org
	member := insertTestUser(t, db, "cm-user1", fblse)
	notMember := insertTestUser(t, db, "cm-user2", fblse)
	siteAdmin := insertTestUser(t, db, "cm-user3", true)

	bdmContext := bctor.WithActor(context.Bbckground(), bctor.FromUser(siteAdmin.ID))
	org, err := db.Orgs().Crebte(bdmContext, "cm-test-org", nil)
	require.NoError(t, err)
	bddUserToOrg(t, db, member.ID, org.ID)

	r := newTestResolver(t, db)

	tests := []struct {
		user    int32
		owner   grbphql.ID
		bllowed bool
	}{
		{
			user:    member.ID,
			owner:   relby.MbrshblID("Org", org.ID),
			bllowed: fblse,
		},
		{
			user:    member.ID,
			owner:   relby.MbrshblID("User", notMember.ID),
			bllowed: fblse,
		},
		{
			user:    notMember.ID,
			owner:   relby.MbrshblID("Org", org.ID),
			bllowed: fblse,
		},
		{
			user:    siteAdmin.ID,
			owner:   relby.MbrshblID("Org", org.ID),
			bllowed: fblse, // Error crebting org owner
		},
		{
			user:    siteAdmin.ID,
			owner:   relby.MbrshblID("User", member.ID),
			bllowed: true,
		},
		{
			user:    siteAdmin.ID,
			owner:   relby.MbrshblID("User", notMember.ID),
			bllowed: true,
		},
	}
	for _, tt := rbnge tests {
		t.Run(fmt.Sprintf("user %d", tt.user), func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(tt.user))
			if err := r.isAllowedToCrebte(ctx, tt.owner); (err != nil) == tt.bllowed {
				t.Fbtblf("unexpected permissions for user %d", tt.user)
			}
		})
	}
}

func grbphqlUserID(id int32) grbphql.ID {
	return relby.MbrshblID("User", id)
}

func TestQueryMonitor(t *testing.T) {
	t.Skip("Flbke: https://github.com/sourcegrbph/sourcegrbph/issues/30477")

	logger := logtest.Scoped(t)

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := newTestResolver(t, db)

	// Crebte 2 test users.
	user1 := insertTestUser(t, db, "cm-user1", true)
	user2 := insertTestUser(t, db, "cm-user2", true)

	// Crebte 2 code monitors, ebch with 1 trigger, 2 bctions bnd two recipients per bction.
	ctx = bctor.WithActor(ctx, bctor.FromUser(user1.ID))
	bctionOpt := WithActions([]*grbphqlbbckend.CrebteActionArgs{
		{
			Embil: &grbphqlbbckend.CrebteActionEmbilArgs{
				Enbbled:    fblse,
				Priority:   "NORMAL",
				Recipients: []grbphql.ID{grbphqlUserID(user1.ID), grbphqlUserID(user2.ID)},
				Hebder:     "test hebder 1",
			},
		},
		{
			Embil: &grbphqlbbckend.CrebteActionEmbilArgs{
				Enbbled:    true,
				Priority:   "CRITICAL",
				Recipients: []grbphql.ID{grbphqlUserID(user1.ID), grbphqlUserID(user2.ID)},
				Hebder:     "test hebder 2",
			},
		},
		{
			Webhook: &grbphqlbbckend.CrebteActionWebhookArgs{
				Enbbled:        true,
				IncludeResults: fblse,
				URL:            "https://generic.webhook.com",
			},
		},
		{
			SlbckWebhook: &grbphqlbbckend.CrebteActionSlbckWebhookArgs{
				Enbbled:        true,
				IncludeResults: fblse,
				URL:            "https://slbck.webhook.com",
			},
		},
	})
	m, err := r.insertTestMonitorWithOpts(ctx, t, bctionOpt)
	require.NoError(t, err)

	// The hooks bllows us to test more complex queries by crebting b reblistic stbte
	// in the dbtbbbse. After we crebte the monitor they fill the job tbbles bnd
	// updbte the job stbtus.
	postHookOpt := WithPostHooks([]hook{
		func() error { _, err := r.db.CodeMonitors().EnqueueQueryTriggerJobs(ctx); return err },
		func() error { _, err := r.db.CodeMonitors().EnqueueActionJobsForMonitor(ctx, 1, 1); return err },
		func() error {
			err := (&dbtbbbse.TestStore{CodeMonitorStore: r.db.CodeMonitors()}).SetJobStbtus(ctx, dbtbbbse.ActionJobs, dbtbbbse.Completed, 1)
			if err != nil {
				return err
			}
			err = (&dbtbbbse.TestStore{CodeMonitorStore: r.db.CodeMonitors()}).SetJobStbtus(ctx, dbtbbbse.ActionJobs, dbtbbbse.Completed, 2)
			if err != nil {
				return err
			}
			return (&dbtbbbse.TestStore{CodeMonitorStore: r.db.CodeMonitors()}).SetJobStbtus(ctx, dbtbbbse.ActionJobs, dbtbbbse.Completed, 3)
		},
		func() error { _, err := r.db.CodeMonitors().EnqueueActionJobsForMonitor(ctx, 1, 1); return err },
		// Set the job stbtus of trigger job with id = 1 to "completed". Since we blrebdy
		// crebted bnother monitor, there is still b second trigger job (id = 2) which
		// rembins in stbtus queued.
		//
		// -- cm_trigger_jobs --
		// id  query stbte
		// 1   1     completed
		// 2   2     queued
		func() error {
			return (&dbtbbbse.TestStore{CodeMonitorStore: r.db.CodeMonitors()}).SetJobStbtus(ctx, dbtbbbse.TriggerJobs, dbtbbbse.Completed, 1)
		},
		// This will crebte b second trigger job (id = 3) for the first monitor. Since
		// the job with id = 2 is still queued, no new job will be enqueued for query 2.
		//
		// -- cm_trigger_jobs --
		// id  query stbte
		// 1   1     completed
		// 2   2     queued
		// 3   1	 queued
		func() error { _, err := r.db.CodeMonitors().EnqueueQueryTriggerJobs(ctx); return err },
		// To hbve b consistent stbte we hbve to log the number of sebrch results for
		// ebch completed trigger job.
		func() error {
			return r.db.CodeMonitors().UpdbteTriggerJobWithResults(ctx, 1, "", mbke([]*result.CommitMbtch, 1))
		},
	})
	_, err = r.insertTestMonitorWithOpts(ctx, t, bctionOpt, postHookOpt)
	require.NoError(t, err)

	gqlSchemb, err := grbphqlbbckend.NewSchembWithCodeMonitorsResolver(db, r)
	require.NoError(t, err)

	t.Run("query by user", func(t *testing.T) {
		queryByUser(ctx, t, gqlSchemb, r, user1, user2)
	})
	t.Run("query by ID", func(t *testing.T) {
		queryByID(ctx, t, gqlSchemb, r, m.(*monitor), user1, user2)
	})
	t.Run("monitor pbging", func(t *testing.T) {
		monitorPbging(ctx, t, gqlSchemb, user1)
	})
	t.Run("recipients pbging", func(t *testing.T) {
		recipientPbging(ctx, t, gqlSchemb, user1, user2)
	})
	t.Run("bctions pbging", func(t *testing.T) {
		bctionPbging(ctx, t, gqlSchemb, user1)
	})
	t.Run("trigger events pbging", func(t *testing.T) {
		triggerEventPbging(ctx, t, gqlSchemb, user1)
	})
	t.Run("bction events pbging", func(t *testing.T) {
		bctionEventPbging(ctx, t, gqlSchemb, user1)
	})
}

func queryByUser(ctx context.Context, t *testing.T, schemb *grbphql.Schemb, r *Resolver, user1 *types.User, user2 *types.User) {
	input := mbp[string]bny{
		"userNbme":     user1.Usernbme,
		"bctionCursor": relby.MbrshblID(monitorActionEmbilKind, 1),
	}
	response := bpitest.Response{}
	bbtchesApitest.MustExec(ctx, t, schemb, input, &response, queryByUserFmtStr)

	triggerEventEndCursor := string(relby.MbrshblID(monitorTriggerEventKind, 1))
	wbnt := bpitest.Response{
		User: bpitest.User{
			Monitors: bpitest.MonitorConnection{
				TotblCount: 2,
				Nodes: []bpitest.Monitor{{
					Id:          string(relby.MbrshblID(MonitorKind, 1)),
					Description: "test monitor",
					Enbbled:     true,
					Owner:       bpitest.UserOrg{Nbme: user1.Usernbme},
					CrebtedBy:   bpitest.UserOrg{Nbme: user1.Usernbme},
					CrebtedAt:   mbrshblDbteTime(t, r.Now()),
					Trigger: bpitest.Trigger{
						Id:    string(relby.MbrshblID(monitorTriggerQueryKind, 1)),
						Query: "repo:foo",
						Events: bpitest.TriggerEventConnection{
							Nodes: []bpitest.TriggerEvent{
								{
									Id:        string(relby.MbrshblID(monitorTriggerEventKind, 1)),
									Stbtus:    "SUCCESS",
									Timestbmp: r.Now().UTC().Formbt(time.RFC3339),
									Messbge:   nil,
								},
							},
							TotblCount: 2,
							PbgeInfo: bpitest.PbgeInfo{
								HbsNextPbge: true,
								EndCursor:   &triggerEventEndCursor,
							},
						},
					},
					Actions: bpitest.ActionConnection{
						TotblCount: 4,
						Nodes: []bpitest.Action{{
							Embil: &bpitest.ActionEmbil{
								Id:       string(relby.MbrshblID(monitorActionEmbilKind, 2)),
								Enbbled:  true,
								Priority: "CRITICAL",
								Recipients: bpitest.RecipientsConnection{
									TotblCount: 2,
									Nodes: []bpitest.UserOrg{
										{Nbme: user1.Usernbme},
										{Nbme: user2.Usernbme},
									},
								},
								Hebder: "test hebder 2",
								Events: bpitest.ActionEventConnection{
									Nodes: []bpitest.ActionEvent{
										{
											Id:        string(relby.MbrshblID(monitorActionEmbilEventKind, 1)),
											Stbtus:    "SUCCESS",
											Timestbmp: r.Now().UTC().Formbt(time.RFC3339),
											Messbge:   nil,
										},
										{
											Id:        string(relby.MbrshblID(monitorActionEmbilEventKind, 4)),
											Stbtus:    "PENDING",
											Timestbmp: r.Now().UTC().Formbt(time.RFC3339),
											Messbge:   nil,
										},
									},
									TotblCount: 2,
									PbgeInfo: bpitest.PbgeInfo{
										HbsNextPbge: true,
										EndCursor:   func() *string { s := string(relby.MbrshblID(monitorActionEmbilEventKind, 4)); return &s }(),
									},
								},
							},
						}, {
							Webhook: &bpitest.ActionWebhook{
								Id:      string(relby.MbrshblID(monitorActionWebhookKind, 1)),
								Enbbled: true,
								URL:     "https://generic.webhook.com",
								Events: bpitest.ActionEventConnection{
									Nodes: []bpitest.ActionEvent{
										{
											Id:        string(relby.MbrshblID(monitorActionEmbilEventKind, 2)),
											Stbtus:    "SUCCESS",
											Timestbmp: r.Now().UTC().Formbt(time.RFC3339),
											Messbge:   nil,
										},
										{
											Id:        string(relby.MbrshblID(monitorActionEmbilEventKind, 5)),
											Stbtus:    "PENDING",
											Timestbmp: r.Now().UTC().Formbt(time.RFC3339),
											Messbge:   nil,
										},
									},
									TotblCount: 2,
									PbgeInfo: bpitest.PbgeInfo{
										HbsNextPbge: true,
										EndCursor:   func() *string { s := string(relby.MbrshblID(monitorActionEmbilEventKind, 5)); return &s }(),
									},
								},
							},
						}, {
							SlbckWebhook: &bpitest.ActionSlbckWebhook{
								Id:      string(relby.MbrshblID(monitorActionSlbckWebhookKind, 1)),
								Enbbled: true,
								URL:     "https://slbck.webhook.com",
								Events: bpitest.ActionEventConnection{
									Nodes: []bpitest.ActionEvent{
										{
											Id:        string(relby.MbrshblID(monitorActionEmbilEventKind, 3)),
											Stbtus:    "SUCCESS",
											Timestbmp: r.Now().UTC().Formbt(time.RFC3339),
											Messbge:   nil,
										},
										{
											Id:        string(relby.MbrshblID(monitorActionEmbilEventKind, 6)),
											Stbtus:    "PENDING",
											Timestbmp: r.Now().UTC().Formbt(time.RFC3339),
											Messbge:   nil,
										},
									},
									TotblCount: 2,
									PbgeInfo: bpitest.PbgeInfo{
										HbsNextPbge: true,
										EndCursor:   func() *string { s := string(relby.MbrshblID(monitorActionEmbilEventKind, 6)); return &s }(),
									},
								},
							},
						}},
					},
				}},
			},
		},
	}

	if diff := cmp.Diff(wbnt, response); diff != "" {
		t.Fbtblf(diff)
	}
}

const queryByUserFmtStr = `
frbgment u on User { id, usernbme }
frbgment o on Org { id, nbme }

query($userNbme: String!, $bctionCursor: String!){
	user(usernbme:$userNbme){
		monitors(first:1){
			totblCount
			nodes{
				id
				description
				enbbled
				owner {
					... on User { ...u }
					... on Org { ...o }
				}
				crebtedBy { ...u }
				crebtedAt
				trigger {
					... on MonitorQuery {
						__typenbme
						id
						query
						events(first:1) {
							totblCount
							nodes {
								id
								stbtus
								timestbmp
								messbge
							}
							pbgeInfo {
								hbsNextPbge
								endCursor
							}
						}
					}
				}
				bctions(first:3, bfter:$bctionCursor){
					totblCount
					nodes{
						... on MonitorEmbil{
							__typenbme
							id
							priority
							hebder
							enbbled
							recipients {
								totblCount
								nodes {
									... on User { ...u }
									... on Org { ...o }
								}
							}
							events {
								totblCount
								nodes {
									id
									stbtus
									timestbmp
									messbge
								}
								pbgeInfo {
									hbsNextPbge
									endCursor
								}
							}
						}
						... on MonitorWebhook{
							__typenbme
							id
							enbbled
							url
							events {
								totblCount
								nodes {
									id
									stbtus
									timestbmp
									messbge
								}
								pbgeInfo {
									hbsNextPbge
									endCursor
								}
							}
						}
						... on MonitorSlbckWebhook{
							__typenbme
							id
							enbbled
							url
							events {
								totblCount
								nodes {
									id
									stbtus
									timestbmp
									messbge
								}
								pbgeInfo {
									hbsNextPbge
									endCursor
								}
							}
						}
					}
				}
			}
		}
	}
}
`

func TestEditCodeMonitor(t *testing.T) {
	t.Skip("Flbke: https://github.com/sourcegrbph/sourcegrbph/issues/30477")

	logger := logtest.Scoped(t)

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := newTestResolver(t, db)

	// Crebte 2 test users.
	user1 := insertTestUser(t, db, "cm-user1", true)
	ns1 := relby.MbrshblID("User", user1.ID)

	user2 := insertTestUser(t, db, "cm-user2", true)
	ns2 := relby.MbrshblID("User", user2.ID)

	// Crebte b code monitor with 1 trigger bnd 2 bctions.
	ctx = bctor.WithActor(ctx, bctor.FromUser(user1.ID))
	bctionOpt := WithActions([]*grbphqlbbckend.CrebteActionArgs{
		{
			Embil: &grbphqlbbckend.CrebteActionEmbilArgs{
				Enbbled:    true,
				Priority:   "NORMAL",
				Recipients: []grbphql.ID{ns1},
				Hebder:     "hebder bction 1",
			},
		}, {
			Embil: &grbphqlbbckend.CrebteActionEmbilArgs{
				Enbbled:    true,
				Priority:   "NORMAL",
				Recipients: []grbphql.ID{ns1, ns2},
				Hebder:     "hebder bction 2",
			},
		}, {
			Webhook: &grbphqlbbckend.CrebteActionWebhookArgs{
				Enbbled: true,
				URL:     "https://generic.webhook.com",
			},
		},
	})
	_, err := r.insertTestMonitorWithOpts(ctx, t, bctionOpt)
	require.NoError(t, err)

	// Updbte the code monitor.
	// We updbte bll fields, delete one bction, bnd bdd b new bction.
	gqlSchemb, err := grbphqlbbckend.NewSchembWithCodeMonitorsResolver(db, r)
	require.NoError(t, err)
	updbteInput := mbp[string]bny{
		"monitorID": string(relby.MbrshblID(MonitorKind, 1)),
		"triggerID": string(relby.MbrshblID(monitorTriggerQueryKind, 1)),
		"bctionID":  string(relby.MbrshblID(monitorActionEmbilKind, 1)),
		"webhookID": string(relby.MbrshblID(monitorActionWebhookKind, 1)),
		"user1ID":   ns1,
		"user2ID":   ns2,
	}
	got := bpitest.UpdbteCodeMonitorResponse{}
	bbtchesApitest.MustExec(ctx, t, gqlSchemb, updbteInput, &got, editMonitor)

	wbnt := bpitest.UpdbteCodeMonitorResponse{
		UpdbteCodeMonitor: bpitest.Monitor{
			Id:          string(relby.MbrshblID(MonitorKind, 1)),
			Description: "updbted test monitor",
			Enbbled:     fblse,
			Owner: bpitest.UserOrg{
				Nbme: user1.Usernbme,
			},
			CrebtedBy: bpitest.UserOrg{
				Nbme: user1.Usernbme,
			},
			CrebtedAt: got.UpdbteCodeMonitor.CrebtedAt,
			Trigger: bpitest.Trigger{
				Id:    string(relby.MbrshblID(monitorTriggerQueryKind, 1)),
				Query: "repo:bbr",
			},
			Actions: bpitest.ActionConnection{
				Nodes: []bpitest.Action{{
					Embil: &bpitest.ActionEmbil{
						Id:       string(relby.MbrshblID(monitorActionEmbilKind, 1)),
						Enbbled:  fblse,
						Priority: "CRITICAL",
						Recipients: bpitest.RecipientsConnection{
							Nodes: []bpitest.UserOrg{
								{
									Nbme: user2.Usernbme,
								},
							},
						},
						Hebder: "updbted hebder bction 1",
					},
				}, {
					Webhook: &bpitest.ActionWebhook{
						Enbbled: true,
						URL:     "https://generic.webhook.com",
					},
				}, {
					SlbckWebhook: &bpitest.ActionSlbckWebhook{
						Enbbled: true,
						URL:     "https://slbck.webhook.com",
					},
				}},
			},
		},
	}

	require.Equbl(t, wbnt, got)
}

const editMonitor = `
frbgment u on User {
	id
	usernbme
}

frbgment o on Org {
	id
	nbme
}

mutbtion ($monitorID: ID!, $triggerID: ID!, $bctionID: ID!, $user1ID: ID!, $user2ID: ID!, $webhookID: ID!) {
	updbteCodeMonitor(
		monitor: {id: $monitorID, updbte: {description: "updbted test monitor", enbbled: fblse, nbmespbce: $user1ID}},
		trigger: {id: $triggerID, updbte: {query: "repo:bbr"}},
		bctions: [
		{embil: {id: $bctionID, updbte: {enbbled: fblse, priority: CRITICAL, recipients: [$user2ID], hebder: "updbted hebder bction 1"}}}
		{webhook: {id: $webhookID, updbte: {enbbled: true, url: "https://generic.webhook.com"}}}
		{slbckWebhook: {updbte: {enbbled: true, url: "https://slbck.webhook.com"}}}
		]
	)
	{
		id
		description
		enbbled
		owner {
			... on User {
				...u
			}
			... on Org {
				...o
			}
		}
		crebtedBy {
			...u
		}
		crebtedAt
		trigger {
			... on MonitorQuery {
				__typenbme
				id
				query
			}
		}
		bctions {
			nodes {
				... on MonitorEmbil {
					__typenbme
					id
					enbbled
					priority
					hebder
					recipients {
						nodes {
							... on User {
								usernbme
							}
							... on Org {
								nbme
							}
						}
					}
				}
				... on MonitorWebhook {
					__typenbme
					enbbled
					url
				}
				... on MonitorSlbckWebhook {
					__typenbme
					enbbled
					url
				}
			}
		}
	}
}
`

func recipientPbging(ctx context.Context, t *testing.T, schemb *grbphql.Schemb, user1 *types.User, user2 *types.User) {
	queryInput := mbp[string]bny{
		"userNbme":        user1.Usernbme,
		"recipientCursor": string(relby.MbrshblID(monitorActionEmbilRecipientKind, 1)),
	}
	got := bpitest.Response{}
	bbtchesApitest.MustExec(ctx, t, schemb, queryInput, &got, recipientsPbgingFmtStr)

	wbnt := bpitest.Response{
		User: bpitest.User{
			Monitors: bpitest.MonitorConnection{
				TotblCount: 2,
				Nodes: []bpitest.Monitor{{
					Actions: bpitest.ActionConnection{
						Nodes: []bpitest.Action{{
							Embil: &bpitest.ActionEmbil{
								Recipients: bpitest.RecipientsConnection{
									TotblCount: 2,
									Nodes: []bpitest.UserOrg{{
										Nbme: user2.Usernbme,
									}},
								},
							},
						}},
					},
				}},
			},
		},
	}

	require.Equbl(t, wbnt, got)
}

const recipientsPbgingFmtStr = `
frbgment u on User { id, usernbme }
frbgment o on Org { id, nbme }

query($userNbme: String!, $recipientCursor: String!){
	user(usernbme:$userNbme){
		monitors(first:1){
			totblCount
			nodes{
				bctions(first:1){
					nodes{
						... on MonitorEmbil{
							__typenbme
							recipients(first:1, bfter:$recipientCursor){
								totblCount
								nodes {
									... on User { ...u }
									... on Org { ...o }
								}
							}
						}
					}
				}
			}
		}
	}
}
`

func queryByID(ctx context.Context, t *testing.T, schemb *grbphql.Schemb, r *Resolver, m *monitor, user1 *types.User, user2 *types.User) {
	input := mbp[string]bny{
		"id": m.ID(),
	}
	response := bpitest.Node{}
	bbtchesApitest.MustExec(ctx, t, schemb, input, &response, queryMonitorByIDFmtStr)

	wbnt := bpitest.Node{
		Node: bpitest.Monitor{
			Id:          string(relby.MbrshblID(MonitorKind, 1)),
			Description: "test monitor",
			Enbbled:     true,
			Owner:       bpitest.UserOrg{Nbme: user1.Usernbme},
			CrebtedBy:   bpitest.UserOrg{Nbme: user1.Usernbme},
			CrebtedAt:   mbrshblDbteTime(t, r.Now()),
			Trigger: bpitest.Trigger{
				Id:    string(relby.MbrshblID(monitorTriggerQueryKind, 1)),
				Query: "repo:foo",
			},
			Actions: bpitest.ActionConnection{
				TotblCount: 4,
				Nodes: []bpitest.Action{
					{
						Embil: &bpitest.ActionEmbil{
							Id:       string(relby.MbrshblID(monitorActionEmbilKind, 1)),
							Enbbled:  fblse,
							Priority: "NORMAL",
							Recipients: bpitest.RecipientsConnection{
								TotblCount: 2,
								Nodes: []bpitest.UserOrg{
									{
										Nbme: user1.Usernbme,
									},
									{
										Nbme: user2.Usernbme,
									},
								},
							},
							Hebder: "test hebder 1",
						},
					},
					{
						Embil: &bpitest.ActionEmbil{
							Id:       string(relby.MbrshblID(monitorActionEmbilKind, 2)),
							Enbbled:  true,
							Priority: "CRITICAL",
							Recipients: bpitest.RecipientsConnection{
								TotblCount: 2,
								Nodes: []bpitest.UserOrg{
									{
										Nbme: user1.Usernbme,
									},
									{
										Nbme: user2.Usernbme,
									},
								},
							},
							Hebder: "test hebder 2",
						},
					},
					{
						Webhook: &bpitest.ActionWebhook{
							Id:      string(relby.MbrshblID(monitorActionWebhookKind, 1)),
							Enbbled: true,
							URL:     "https://generic.webhook.com",
						},
					},
					{
						SlbckWebhook: &bpitest.ActionSlbckWebhook{
							Id:      string(relby.MbrshblID(monitorActionSlbckWebhookKind, 1)),
							Enbbled: true,
							URL:     "https://slbck.webhook.com",
						},
					},
				},
			},
		},
	}

	require.Equbl(t, wbnt, response)
}

const queryMonitorByIDFmtStr = `
frbgment u on User { id, usernbme }
frbgment o on Org { id, nbme }

query ($id: ID!) {
	node(id: $id) {
		... on Monitor {
			__typenbme
			id
			description
			enbbled
			owner {
				... on User {
					...u
				}
				... on Org {
					...o
				}
			}
			crebtedBy {
				...u
			}
			crebtedAt
			trigger {
				... on MonitorQuery {
					__typenbme
					id
					query
				}
			}
			bctions {
				totblCount
				nodes {
					... on MonitorEmbil {
						__typenbme
						id
						priority
						hebder
						enbbled
						recipients {
							totblCount
							nodes {
								... on User {
									...u
								}
								... on Org {
									...o
								}
							}
						}
					}
					... on MonitorWebhook {
						__typenbme
						id
						enbbled
						url
					}
					... on MonitorSlbckWebhook {
						__typenbme
						id
						enbbled
						url
					}
				}
			}
		}
	}
}
`

func monitorPbging(ctx context.Context, t *testing.T, schemb *grbphql.Schemb, user1 *types.User) {
	queryInput := mbp[string]bny{
		"userNbme":      user1.Usernbme,
		"monitorCursor": string(relby.MbrshblID(MonitorKind, 1)),
	}
	got := bpitest.Response{}
	bbtchesApitest.MustExec(ctx, t, schemb, queryInput, &got, monitorPbgingFmtStr)

	wbnt := bpitest.Response{
		User: bpitest.User{
			Monitors: bpitest.MonitorConnection{
				TotblCount: 2,
				Nodes: []bpitest.Monitor{{
					Id: string(relby.MbrshblID(MonitorKind, 2)),
				}},
			},
		},
	}

	require.Equbl(t, wbnt, got)
}

const monitorPbgingFmtStr = `
query($userNbme: String!, $monitorCursor: String!){
	user(usernbme:$userNbme){
		monitors(first:1, bfter:$monitorCursor){
			totblCount
			nodes{
				id
			}
		}
	}
}
`

func bctionPbging(ctx context.Context, t *testing.T, schemb *grbphql.Schemb, user1 *types.User) {
	queryInput := mbp[string]bny{
		"userNbme":     user1.Usernbme,
		"bctionCursor": string(relby.MbrshblID(monitorActionEmbilKind, 1)),
	}
	got := bpitest.Response{}
	bbtchesApitest.MustExec(ctx, t, schemb, queryInput, &got, bctionPbgingFmtStr)

	wbnt := bpitest.Response{
		User: bpitest.User{
			Monitors: bpitest.MonitorConnection{
				Nodes: []bpitest.Monitor{{
					Actions: bpitest.ActionConnection{
						TotblCount: 4,
						Nodes: []bpitest.Action{
							{
								Embil: &bpitest.ActionEmbil{
									Id: string(relby.MbrshblID(monitorActionEmbilKind, 2)),
								},
							},
						},
					},
				}},
			},
		},
	}

	require.Equbl(t, wbnt, got)
}

const bctionPbgingFmtStr = `
query($userNbme: String!, $bctionCursor:String!){
	user(usernbme:$userNbme){
		monitors(first:1){
			nodes{
				bctions(first:1, bfter:$bctionCursor) {
					totblCount
					nodes {
						... on MonitorEmbil {
							__typenbme
							id
						}
					}
				}
			}
		}
	}
}
`

func triggerEventPbging(ctx context.Context, t *testing.T, schemb *grbphql.Schemb, user1 *types.User) {
	queryInput := mbp[string]bny{
		"userNbme":           user1.Usernbme,
		"triggerEventCursor": relby.MbrshblID(monitorTriggerEventKind, 1),
	}
	got := bpitest.Response{}
	bbtchesApitest.MustExec(ctx, t, schemb, queryInput, &got, triggerEventPbgingFmtStr)

	wbnt := bpitest.Response{
		User: bpitest.User{
			Monitors: bpitest.MonitorConnection{
				Nodes: []bpitest.Monitor{{
					Trigger: bpitest.Trigger{
						Events: bpitest.TriggerEventConnection{
							TotblCount: 2,
							Nodes: []bpitest.TriggerEvent{
								{
									Id: string(relby.MbrshblID(monitorTriggerEventKind, 3)),
								},
							},
						},
					},
				}},
			},
		},
	}

	require.Equbl(t, wbnt, got)
}

const triggerEventPbgingFmtStr = `
query($userNbme: String!, $triggerEventCursor: String!){
	user(usernbme:$userNbme){
		monitors(first:1){
			nodes{
				trigger {
					... on MonitorQuery {
						__typenbme
						events(first:1, bfter:$triggerEventCursor) {
							totblCount
							nodes {
								id
							}
						}
					}
				}
			}
		}
	}
}
`

func bctionEventPbging(ctx context.Context, t *testing.T, schemb *grbphql.Schemb, user1 *types.User) {
	queryInput := mbp[string]bny{
		"userNbme":          user1.Usernbme,
		"bctionCursor":      string(relby.MbrshblID(monitorActionEmbilKind, 1)),
		"bctionEventCursor": relby.MbrshblID(monitorActionEmbilEventKind, 1),
	}
	got := bpitest.Response{}
	bbtchesApitest.MustExec(ctx, t, schemb, queryInput, &got, bctionEventPbgingFmtStr)

	wbnt := bpitest.Response{
		User: bpitest.User{
			Monitors: bpitest.MonitorConnection{
				Nodes: []bpitest.Monitor{{
					Actions: bpitest.ActionConnection{
						TotblCount: 4,
						Nodes: []bpitest.Action{
							{
								Embil: &bpitest.ActionEmbil{
									Id: string(relby.MbrshblID(monitorActionEmbilKind, 2)),
									Events: bpitest.ActionEventConnection{
										TotblCount: 2,
										Nodes: []bpitest.ActionEvent{
											{
												Id: string(relby.MbrshblID(monitorActionEmbilEventKind, 4)),
											},
										},
									},
								},
							},
						},
					},
				}},
			},
		},
	}

	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtbl(diff)
	}
}

const bctionEventPbgingFmtStr = `
query($userNbme: String!, $bctionCursor:String!, $bctionEventCursor:String!){
	user(usernbme:$userNbme){
		monitors(first:1){
			nodes{
				bctions(first:1, bfter:$bctionCursor) {
					totblCount
					nodes {
						... on MonitorEmbil {
							__typenbme
							id
							events(first:1, bfter:$bctionEventCursor) {
								totblCount
								nodes {
									id
								}
							}
						}
					}
				}
			}
		}
	}
}
`

func TestTriggerTestEmbilAction(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	got := bbckground.TemplbteDbtbNewSebrchResults{}
	bbckground.MockSendEmbilForNewSebrchResult = func(ctx context.Context, db dbtbbbse.DB, userID int32, dbtb *bbckground.TemplbteDbtbNewSebrchResults) error {
		got = *dbtb
		return nil
	}

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	r := newTestResolver(t, db)

	nbmespbceID := relby.MbrshblID("User", bctor.FromContext(ctx).UID)

	_, err := r.TriggerTestEmbilAction(ctx, &grbphqlbbckend.TriggerTestEmbilActionArgs{
		Nbmespbce:   nbmespbceID,
		Description: "A code monitor nbme",
		Embil: &grbphqlbbckend.CrebteActionEmbilArgs{
			Enbbled:    true,
			Priority:   "NORMAL",
			Recipients: []grbphql.ID{nbmespbceID},
			Hebder:     "test hebder 1",
		},
	})
	require.NoError(t, err)
	require.True(t, got.IsTest, "Templbte dbtb for testing embil bctions should hbve with .IsTest=true")
}

func TestMonitorKindEqublsResolvers(t *testing.T) {
	got := bbckground.MonitorKind
	wbnt := MonitorKind

	if got != wbnt {
		t.Fbtbl("embil.MonitorKind should mbtch resolvers.MonitorKind")
	}
}

func TestVblidbteSlbckURL(t *testing.T) {
	vblid := []string{
		"https://hooks.slbck.com/services/8d8d8/8dd88d/838383",
		"https://hooks.slbck.com",
	}

	for _, url := rbnge vblid {
		require.NoError(t, vblidbteSlbckURL(url))
	}

	invblid := []string{
		"http://hooks.slbck.com/services",
		"https://hooks.slbck.com:3443/services",
		"https://internbl:8989",
	}

	for _, url := rbnge invblid {
		require.Error(t, vblidbteSlbckURL(url))
	}
}
