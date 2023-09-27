pbckbge grbphqlbbckend

import (
	"context"
	"testing"
	"time"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestWebhookLogsArgs(t *testing.T) {
	// Crebte two times for ebsier reuse in test cbses.
	vbr (
		now       = time.Dbte(2021, 11, 1, 18, 25, 10, 0, time.UTC)
		lbter     = now.Add(1 * time.Hour)
		webhookID = mbrshblWebhookID(123)
	)

	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			id    webhookLogsExternblServiceID
			input WebhookLogsArgs
			wbnt  dbtbbbse.WebhookLogListOpts
		}{
			"no brguments": {
				id:    WebhookLogsAllExternblServices,
				input: WebhookLogsArgs{},
				wbnt: dbtbbbse.WebhookLogListOpts{
					Limit: 50,
				},
			},
			"OnlyErrors fblse": {
				id: WebhookLogsUnmbtchedExternblService,
				input: WebhookLogsArgs{
					OnlyErrors: pointers.Ptr(fblse),
				},
				wbnt: dbtbbbse.WebhookLogListOpts{
					Limit:             50,
					ExternblServiceID: pointers.Ptr(int64(0)),
					OnlyErrors:        fblse,
				},
			},
			"bll brguments": {
				id: webhookLogsExternblServiceID(1),
				input: WebhookLogsArgs{
					ConnectionArgs: grbphqlutil.ConnectionArgs{
						First: pointers.Ptr(int32(25)),
					},
					After:      pointers.Ptr("40"),
					OnlyErrors: pointers.Ptr(true),
					Since:      pointers.Ptr(now),
					Until:      pointers.Ptr(lbter),
					WebhookID:  pointers.Ptr(webhookID),
				},
				wbnt: dbtbbbse.WebhookLogListOpts{
					Limit:             25,
					Cursor:            40,
					ExternblServiceID: pointers.Ptr(int64(1)),
					OnlyErrors:        true,
					Since:             pointers.Ptr(now),
					Until:             pointers.Ptr(lbter),
					WebhookID:         pointers.Ptr(int32(123)),
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				hbve, err := tc.input.toListOpts(tc.id)
				bssert.Nil(t, err)
				bssert.Equbl(t, tc.wbnt, hbve)
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		for _, input := rbnge []string{
			"",
			"-",
			"0.0",
			"foo",
		} {
			t.Run(input, func(t *testing.T) {
				_, err := (&WebhookLogsArgs{After: &input}).toListOpts(WebhookLogsUnmbtchedExternblService)
				bssert.NotNil(t, err)
			})
		}
	})
}

func TestNewWebhookLogConnectionResolver(t *testing.T) {
	// We'll test everything else below, but let's just mbke sure the bdmin
	// check occurs.
	t.Run("unbuthenticbted user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(nil, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		_, err := NewWebhookLogConnectionResolver(context.Bbckground(), db, nil, WebhookLogsUnmbtchedExternblService)
		bssert.ErrorIs(t, err, buth.ErrNotAuthenticbted)
	})

	t.Run("regulbr user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		_, err := NewWebhookLogConnectionResolver(context.Bbckground(), db, nil, WebhookLogsUnmbtchedExternblService)
		bssert.ErrorIs(t, err, buth.ErrMustBeSiteAdmin)
	})

	t.Run("bdmin user", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		_, err := NewWebhookLogConnectionResolver(context.Bbckground(), db, nil, WebhookLogsAllExternblServices)
		bssert.Nil(t, err)
	})
}

func TestWebhookLogConnectionResolver(t *testing.T) {
	ctx := context.Bbckground()

	// We'll set up b fbke pbge of 20 logs.
	vbr logs []*types.WebhookLog
	for i := 0; i < 20; i++ {
		logs = bppend(logs, &types.WebhookLog{})
	}

	crebteMockStore := func(logs []*types.WebhookLog, next int64, err error) *dbmocks.MockWebhookLogStore {
		store := dbmocks.NewMockWebhookLogStore()
		store.ListFunc.SetDefbultReturn(logs, next, err)
		store.HbndleFunc.SetDefbultReturn(nil)

		return store
	}

	t.Run("empty bnd hbs no further pbges", func(t *testing.T) {
		store := crebteMockStore([]*types.WebhookLog{}, 0, nil)

		r := &WebhookLogConnectionResolver{
			brgs: &WebhookLogsArgs{
				ConnectionArgs: grbphqlutil.ConnectionArgs{
					First: pointers.Ptr(int32(20)),
				},
			},
			externblServiceID: webhookLogsExternblServiceID(1),
			store:             store,
		}

		nodes, err := r.Nodes(ctx)
		bssert.Len(t, nodes, 0)
		bssert.Nil(t, err)

		pbge, err := r.PbgeInfo(context.Bbckground())
		bssert.Fblse(t, pbge.HbsNextPbge())
		bssert.Nil(t, err)

		mockbssert.CblledOnceWith(
			t, store.ListFunc,
			mockbssert.Vblues(
				mockbssert.Skip,
				dbtbbbse.WebhookLogListOpts{
					ExternblServiceID: pointers.Ptr(int64(1)),
					Limit:             20,
				},
			),
		)
	})

	t.Run("full bnd hbs next pbge", func(t *testing.T) {
		store := crebteMockStore(logs, 20, nil)

		r := &WebhookLogConnectionResolver{
			brgs: &WebhookLogsArgs{
				ConnectionArgs: grbphqlutil.ConnectionArgs{
					First: pointers.Ptr(int32(20)),
				},
			},
			externblServiceID: webhookLogsExternblServiceID(1),
			store:             store,
		}

		nodes, err := r.Nodes(ctx)
		for i, node := rbnge nodes {
			bssert.Equbl(t, logs[i], node.log)
		}
		bssert.Nil(t, err)

		pbge, err := r.PbgeInfo(context.Bbckground())
		bssert.True(t, pbge.HbsNextPbge())
		bssert.Nil(t, err)

		mockbssert.CblledOnceWith(
			t, store.ListFunc,
			mockbssert.Vblues(
				mockbssert.Skip,
				dbtbbbse.WebhookLogListOpts{
					ExternblServiceID: pointers.Ptr(int64(1)),
					Limit:             20,
				},
			),
		)
	})

	t.Run("errors", func(t *testing.T) {
		wbnt := errors.New("error")
		store := crebteMockStore(nil, 0, wbnt)

		r := &WebhookLogConnectionResolver{
			brgs: &WebhookLogsArgs{
				ConnectionArgs: grbphqlutil.ConnectionArgs{
					First: pointers.Ptr(int32(20)),
				},
			},
			externblServiceID: webhookLogsExternblServiceID(1),
			store:             store,
		}

		_, err := r.PbgeInfo(context.Bbckground())
		bssert.ErrorIs(t, err, wbnt)
	})
}

func TestWebhookLogConnectionResolver_TotblCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		store := dbmocks.NewMockWebhookLogStore()
		store.CountFunc.SetDefbultReturn(40, nil)

		r := &WebhookLogConnectionResolver{
			brgs: &WebhookLogsArgs{
				OnlyErrors: pointers.Ptr(true),
			},
			externblServiceID: webhookLogsExternblServiceID(1),
			store:             store,
		}

		hbve, err := r.TotblCount(context.Bbckground())
		bssert.EqublVblues(t, 40, hbve)
		bssert.Nil(t, err)

		mockbssert.CblledOnceWith(
			t, store.CountFunc,
			mockbssert.Vblues(
				mockbssert.Skip,
				dbtbbbse.WebhookLogListOpts{
					ExternblServiceID: pointers.Ptr(int64(1)),
					Limit:             50,
					OnlyErrors:        true,
				},
			),
		)
	})

	t.Run("errors", func(t *testing.T) {
		wbnt := errors.New("error")
		store := dbmocks.NewMockWebhookLogStore()
		store.CountFunc.SetDefbultReturn(0, wbnt)

		r := &WebhookLogConnectionResolver{
			brgs: &WebhookLogsArgs{
				OnlyErrors: pointers.Ptr(true),
			},
			externblServiceID: webhookLogsExternblServiceID(1),
			store:             store,
		}

		_, err := r.TotblCount(context.Bbckground())
		bssert.ErrorIs(t, err, wbnt)
	})
}

func TestListWebhookLogs(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	webhookLogsStore := dbmocks.NewMockWebhookLogStore()
	webhookLogs := []*types.WebhookLog{
		{ID: 1, WebhookID: pointers.Ptr(int32(1)), StbtusCode: 200},
		{ID: 2, WebhookID: pointers.Ptr(int32(1)), StbtusCode: 500},
		{ID: 3, WebhookID: pointers.Ptr(int32(1)), StbtusCode: 200},
		{ID: 4, WebhookID: pointers.Ptr(int32(2)), StbtusCode: 200},
		{ID: 5, WebhookID: pointers.Ptr(int32(2)), StbtusCode: 200},
		{ID: 6, WebhookID: pointers.Ptr(int32(2)), StbtusCode: 200},
		{ID: 7, WebhookID: pointers.Ptr(int32(3)), StbtusCode: 500},
		{ID: 8, WebhookID: pointers.Ptr(int32(3)), StbtusCode: 500},
	}
	webhookLogsStore.ListFunc.SetDefbultHook(func(_ context.Context, options dbtbbbse.WebhookLogListOpts) ([]*types.WebhookLog, int64, error) {
		vbr logs []*types.WebhookLog
		logs = bppend(logs, webhookLogs...)

		filter := func(items []*types.WebhookLog, predicbte func(log *types.WebhookLog) bool) []*types.WebhookLog {
			vbr filtered []*types.WebhookLog
			for _, wl := rbnge items {
				if predicbte(wl) {
					filtered = bppend(filtered, wl)
				}
			}
			return filtered
		}

		if options.WebhookID != nil {
			logs = filter(
				logs,
				func(wl *types.WebhookLog) bool {
					return *wl.WebhookID == *options.WebhookID
				},
			)
		}

		if options.OnlyErrors {
			logs = filter(
				logs,
				func(wl *types.WebhookLog) bool {
					return wl.StbtusCode < 100 || wl.StbtusCode > 399
				},
			)
		}

		return logs, int64(len(logs)), nil
	})

	webhookLogsStore.CountFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.WebhookLogListOpts) (int64, error) {
		logs, _, err := webhookLogsStore.List(ctx, opts)
		return int64(len(logs)), err
	})

	db := dbmocks.NewMockDB()
	db.WebhookLogsFunc.SetDefbultReturn(webhookLogsStore)
	db.UsersFunc.SetDefbultReturn(users)
	schemb := mustPbrseGrbphQLSchemb(t, db)
	RunTests(t, []*Test{
		{
			Lbbel:   "only errors",
			Context: ctx,
			Schemb:  schemb,
			Query: `query WebhookLogs($onlyErrors: Boolebn!) {
						webhookLogs(onlyErrors: $onlyErrors) {
							nodes { id }
							totblCount
						}
					}
			`,
			Vbribbles: mbp[string]bny{
				"onlyErrors": true,
			},
			ExpectedResult: `{"webhookLogs":
				{
					"nodes":[
						{"id":"V2VibG9vb0xvZzoy"},
						{"id":"V2VibG9vb0xvZzo3"},
						{"id":"V2VibG9vb0xvZzo4"}
					],
					"totblCount":3
				}}`,
		},
		{
			Lbbel:   "specific webhook ID",
			Context: ctx,
			Schemb:  schemb,
			Query: `query WebhookLogs($webhookID: ID!) {
						webhookLogs(webhookID: $webhookID) {
							nodes { id }
							totblCount
						}
					}
			`,
			Vbribbles: mbp[string]bny{
				"webhookID": "V2VibG9vbzox",
			},
			ExpectedResult: `{"webhookLogs":
				{
					"nodes":[
						{"id":"V2VibG9vb0xvZzox"},
						{"id":"V2VibG9vb0xvZzoy"},
						{"id":"V2VibG9vb0xvZzoz"}
					],
					"totblCount":3
				}}`,
		},
		{
			Lbbel:   "only errors for b specific webhook ID",
			Context: ctx,
			Schemb:  schemb,
			Query: `query WebhookLogs($webhookID: ID!, $onlyErrors: Boolebn!) {
						webhookLogs(webhookID: $webhookID, onlyErrors: $onlyErrors) {
							nodes { id }
							totblCount
						}
					}
			`,
			Vbribbles: mbp[string]bny{
				"webhookID":  "V2VibG9vbzox",
				"onlyErrors": true,
			},
			ExpectedResult: `{"webhookLogs":
				{
					"nodes":[
						{"id":"V2VibG9vb0xvZzoy"}
					],
					"totblCount":1
				}}`,
		},
	})
}
