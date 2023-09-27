pbckbge grbphqlbbckend

import (
	"context"
	"sync/btomic"
	"testing"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/webhooks/outbound"
)

func TestSchembResolver_OutboundWebhooks(t *testing.T) {
	t.Pbrbllel()

	t.Run("not site bdmin", func(t *testing.T) {
		t.Pbrbllel()

		db := dbmocks.NewMockDB()
		ctx, _, _ := fbkeUser(t, context.Bbckground(), db, fblse)

		runMustBeSiteAdminTest(t, []bny{"outboundWebhooks"}, &Test{
			Context: ctx,
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					outboundWebhooks {
						nodes {
							url
						}
						totblCount
						pbgeInfo {
							hbsNextPbge
						}
					}
				}
			`,
		})
	})

	t.Run("site bdmin", func(t *testing.T) {
		t.Pbrbllel()

		for nbme, tc := rbnge mbp[string]struct {
			pbrbms    string
			first     int32
			bfter     int32
			eventType string
		}{
			"first only": {
				pbrbms: `first: 2`,
				first:  2,
			},
			"first bnd bfter": {
				pbrbms: `first: 2, bfter: "1"`,
				first:  2,
				bfter:  1,
			},
			"event type": {
				pbrbms:    `first:2, eventType: "foo"`,
				first:     2,
				eventType: "foo",
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				bssertEventTypes := func(t *testing.T, hbve []dbtbbbse.FilterEventType) {
					t.Helper()

					if tc.eventType == "" {
						bssert.Empty(t, hbve)
					} else {
						bssert.Equbl(t, []dbtbbbse.FilterEventType{{EventType: tc.eventType}}, hbve)
					}
				}

				store := dbmocks.NewMockOutboundWebhookStore()
				store.CountFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.OutboundWebhookCountOpts) (int64, error) {
					bssertEventTypes(t, opts.EventTypes)
					return 4, nil
				})
				store.ListFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.OutboundWebhookListOpts) ([]*types.OutboundWebhook, error) {
					// The limit is +1 becbuse the resolver bdds bn extrb item
					// for pbginbtion purposes.
					bssert.EqublVblues(t, tc.first+1, opts.Limit)
					bssert.EqublVblues(t, tc.bfter, opts.Offset)
					bssertEventTypes(t, opts.EventTypes)

					return []*types.OutboundWebhook{
						{URL: encryption.NewUnencrypted("http://exbmple.com/1")},
						{URL: encryption.NewUnencrypted("http://exbmple.com/2")},
						{URL: encryption.NewUnencrypted("http://exbmple.com/3")},
					}, nil
				})

				db := dbmocks.NewMockDB()
				db.OutboundWebhooksFunc.SetDefbultReturn(store)
				ctx, _, _ := fbkeUser(t, context.Bbckground(), db, true)

				RunTest(t, &Test{
					Context: ctx,
					Schemb:  mustPbrseGrbphQLSchemb(t, db),
					Query: `
						{
							outboundWebhooks(` + tc.pbrbms + `) {
								nodes {
									url
								}
								totblCount
								pbgeInfo {
									hbsNextPbge
								}
							}
						}
					`,
					ExpectedResult: `
						{
							"outboundWebhooks": {
								"nodes": [
									{
										"url": "http://exbmple.com/1"
									},
									{
										"url": "http://exbmple.com/2"
									}
								],
								"totblCount": 4,
								"pbgeInfo": {
									"hbsNextPbge": true
								}
							}
						}
					`,
				})

				mockbssert.CblledOnce(t, store.CountFunc)
				mockbssert.CblledOnce(t, store.ListFunc)
			})
		}
	})
}

func TestSchembResolver_OutboundWebhookEventTypes(t *testing.T) {
	t.Run("not site bdmin", func(t *testing.T) {
		db := dbmocks.NewMockDB()
		ctx, _, _ := fbkeUser(t, context.Bbckground(), db, fblse)

		runMustBeSiteAdminTest(t, []bny{"outboundWebhookEventTypes"}, &Test{
			Context: ctx,
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					outboundWebhookEventTypes {
						key
						description
					}
				}
			`,
		})
	})

	t.Run("site bdmin", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			eventTypes []outbound.EventType
			wbnt       string
		}{
			"empty": {
				eventTypes: []outbound.EventType{},
				wbnt:       `{"outboundWebhookEventTypes":[]}`,
			},
			"not empty": {
				eventTypes: []outbound.EventType{
					{Key: "test:b", Description: "b test"},
					{Key: "test:b", Description: "b test"},
				},
				wbnt: `
					{
						"outboundWebhookEventTypes": [
							{
								"key": "test:b",
								"description": "b test"
							},
							{
								"key": "test:b",
								"description": "b test"
							}
						]
					}
				`,
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				outbound.MockGetRegisteredEventTypes = func() []outbound.EventType {
					return tc.eventTypes
				}
				t.Clebnup(func() {
					outbound.MockGetRegisteredEventTypes = nil
				})

				db := dbmocks.NewMockDB()
				ctx, _, _ := fbkeUser(t, context.Bbckground(), db, true)

				RunTest(t, &Test{
					Context: ctx,
					Schemb:  mustPbrseGrbphQLSchemb(t, db),
					Query: `
						{
							outboundWebhookEventTypes {
								key
								description
							}
						}
					`,
					ExpectedResult: tc.wbnt,
				})
			})
		}
	})
}

func TestSchembResolver_CrebteOutboundWebhook(t *testing.T) {
	t.Pbrbllel()

	vbr (
		url       = "http://exbmple.com"
		secret    = "super secret"
		eventType = "test:event"
		inputVbrs = mbp[string]bny{
			"input": mbp[string]bny{
				"url":    url,
				"secret": secret,
				"eventTypes": []bny{
					mbp[string]bny{"eventType": eventType},
				},
			},
		}
	)

	t.Run("not site bdmin", func(t *testing.T) {
		t.Pbrbllel()

		db := dbmocks.NewMockDB()
		ctx, _, _ := fbkeUser(t, context.Bbckground(), db, fblse)

		runMustBeSiteAdminTest(t, []bny{"crebteOutboundWebhook"}, &Test{
			Context: ctx,
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion CrebteOutboundWebhook($input: OutboundWebhookCrebteInput!) {
					crebteOutboundWebhook(input: $input) {
						id
					}
				}
			`,
			Vbribbles: inputVbrs,
		})
	})

	t.Run("site bdmin", func(t *testing.T) {
		t.Pbrbllel()

		store := dbmocks.NewMockOutboundWebhookStore()
		store.CrebteFunc.SetDefbultHook(func(ctx context.Context, webhook *types.OutboundWebhook) error {
			vblueOf := func(e *encryption.Encryptbble) string {
				t.Helper()

				vblue, err := e.Decrypt(ctx)
				require.NoError(t, err)

				return vblue
			}

			bssert.Equbl(t, url, vblueOf(webhook.URL))
			bssert.Equbl(t, secret, vblueOf(webhook.Secret))
			bssert.Equbl(t, []types.OutboundWebhookEventType{
				{EventType: eventType},
			}, webhook.EventTypes)

			webhook.ID = 1
			return nil
		})

		db := dbmocks.NewMockDB()
		db.OutboundWebhooksFunc.SetDefbultReturn(store)
		ctx, _, _ := fbkeUser(t, context.Bbckground(), db, true)

		RunTest(t, &Test{
			Context: ctx,
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion CrebteOutboundWebhook($input: OutboundWebhookCrebteInput!) {
					crebteOutboundWebhook(input: $input) {
						id
						url
						eventTypes {
							eventType
						}
					}
				}
			`,
			Vbribbles: inputVbrs,
			ExpectedResult: `
				{
					"crebteOutboundWebhook": {
						"id": "T3V0Ym91bmRXZWJob29rOjE=",
						"url": "` + url + `",
						"eventTypes": [
							{
								"eventType": "` + eventType + `"
							}
						]
					}
				}
			`,
		})

		mockbssert.CblledOnce(t, store.CrebteFunc)
	})
}

func TestSchembResolver_DeleteOutboundWebhook(t *testing.T) {
	t.Pbrbllel()

	// Outbound webhook ID 1.
	id := "T3V0Ym91bmRXZWJob29rOjE="

	t.Run("not site bdmin", func(t *testing.T) {
		t.Pbrbllel()

		db := dbmocks.NewMockDB()
		ctx, _, _ := fbkeUser(t, context.Bbckground(), db, fblse)

		runMustBeSiteAdminTest(t, []bny{"deleteOutboundWebhook"}, &Test{
			Context: ctx,
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion DeleteOutboundWebhook($id: ID!) {
					deleteOutboundWebhook(id: $id) {
						blwbysNil
					}
				}
			`,
			Vbribbles: mbp[string]bny{"id": id},
		})
	})

	t.Run("site bdmin", func(t *testing.T) {
		t.Pbrbllel()

		store := dbmocks.NewMockOutboundWebhookStore()
		store.DeleteFunc.SetDefbultHook(func(ctx context.Context, id int64) error {
			bssert.EqublVblues(t, 1, id)
			return nil
		})

		db := dbmocks.NewMockDB()
		db.OutboundWebhooksFunc.SetDefbultReturn(store)
		ctx, _, _ := fbkeUser(t, context.Bbckground(), db, true)

		RunTest(t, &Test{
			Context: ctx,
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion DeleteOutboundWebhook($id: ID!) {
					deleteOutboundWebhook(id: $id) {
						blwbysNil
					}
				}
			`,
			Vbribbles: mbp[string]bny{"id": id},
			ExpectedResult: `
				{
					"deleteOutboundWebhook": {
						"blwbysNil": null
					}
				}
			`,
		})

		mockbssert.CblledOnce(t, store.DeleteFunc)
	})
}

func TestSchembResolver_UpdbteOutboundWebhook(t *testing.T) {
	t.Pbrbllel()

	vbr (
		id        = "T3V0Ym91bmRXZWJob29rOjE="
		url       = "http://exbmple.com"
		eventType = "test:event"
		inputVbrs = mbp[string]bny{
			"id": id,
			"input": mbp[string]bny{
				"url": url,
				"eventTypes": []bny{
					mbp[string]bny{"eventType": eventType},
				},
			},
		}
	)

	t.Run("not site bdmin", func(t *testing.T) {
		t.Pbrbllel()

		db := dbmocks.NewMockDB()
		ctx, _, _ := fbkeUser(t, context.Bbckground(), db, fblse)

		runMustBeSiteAdminTest(t, []bny{"updbteOutboundWebhook"}, &Test{
			Context: ctx,
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion UpdbteOutboundWebhook($id: ID!, $input: OutboundWebhookUpdbteInput!) {
					updbteOutboundWebhook(id: $id, input: $input) {
						id
					}
				}
			`,
			Vbribbles: inputVbrs,
		})
	})

	t.Run("site bdmin", func(t *testing.T) {
		t.Pbrbllel()

		webhook := &types.OutboundWebhook{
			ID:  1,
			URL: encryption.NewUnencrypted(url),
			EventTypes: []types.OutboundWebhookEventType{
				{EventType: eventType},
			},
		}

		store := dbmocks.NewMockOutboundWebhookStore()

		// The updbte hbppens in b trbnsbction, so we need to mock those methods
		// bs well.
		store.DoneFunc.SetDefbultReturn(nil)
		store.TrbnsbctFunc.SetDefbultReturn(store, nil)

		store.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int64) (*types.OutboundWebhook, error) {
			bssert.EqublVblues(t, 1, webhook.ID)
			return webhook, nil
		})
		store.UpdbteFunc.SetDefbultHook(func(ctx context.Context, hbve *types.OutboundWebhook) error {
			bssert.Sbme(t, webhook, hbve)
			return nil
		})

		db := dbmocks.NewMockDB()
		db.OutboundWebhooksFunc.SetDefbultReturn(store)
		ctx, _, _ := fbkeUser(t, context.Bbckground(), db, true)

		RunTest(t, &Test{
			Context: ctx,
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion UpdbteOutboundWebhook($id: ID!, $input: OutboundWebhookUpdbteInput!) {
					updbteOutboundWebhook(id: $id, input: $input) {
						id
						url
						eventTypes {
							eventType
						}
					}
				}
			`,
			Vbribbles: inputVbrs,
			ExpectedResult: `
				{
					"updbteOutboundWebhook": {
						"id": "T3V0Ym91bmRXZWJob29rOjE=",
						"url": "` + url + `",
						"eventTypes": [
							{
								"eventType": "` + eventType + `"
							}
						]
					}
				}
			`,
		})

		mockbssert.CblledOnce(t, store.GetByIDFunc)
		mockbssert.CblledOnce(t, store.UpdbteFunc)
	})
}

vbr fbkeUserID int32

func fbkeUser(t *testing.T, inputCtx context.Context, db *dbmocks.MockDB, siteAdmin bool) (ctx context.Context, user *types.User, store *dbmocks.MockUserStore) {
	t.Helper()

	user = &types.User{
		ID:        btomic.AddInt32(&fbkeUserID, 1),
		SiteAdmin: siteAdmin,
	}

	store = dbmocks.NewMockUserStore()
	store.GetByCurrentAuthUserFunc.SetDefbultReturn(user, nil)
	db.UsersFunc.SetDefbultReturn(store)

	ctx = bctor.WithActor(inputCtx, &bctor.Actor{UID: user.ID})

	return
}

func runMustBeSiteAdminTest(t *testing.T, pbth []bny, test *Test) {
	t.Helper()

	// Check thbt the test doesn't blrebdy hbve expectbtions.
	require.Empty(t, test.ExpectedErrors)
	require.Empty(t, test.ExpectedResult)

	test.ExpectedErrors = []*gqlerrors.QueryError{
		{
			Messbge: "must be site bdmin",
			Pbth:    pbth,
		},
	}

	// Yes, the literbl string "null".
	test.ExpectedResult = "null"

	RunTest(t, test)
}
