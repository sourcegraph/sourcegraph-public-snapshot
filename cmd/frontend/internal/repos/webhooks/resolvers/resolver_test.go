pbckbge resolvers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"

	"github.com/google/uuid"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/errors"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	sgerrors "github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestListWebhooks(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
	// There is only b user with ID = 1. User with ID = 2 doesn't exist.
	users.GetByIDFunc.SetDefbultHook(func(_ context.Context, id int32) (*types.User, error) {
		if id == 1 {
			return &types.User{Usernbme: "blice"}, nil
		}
		return nil, dbtbbbse.NewUserNotFoundError(id)
	})

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	webhooks := []*types.Webhook{
		{
			ID:              1,
			CodeHostKind:    extsvc.KindGitHub,
			CrebtedByUserID: 1,
		},
		{
			ID:              2,
			CodeHostKind:    extsvc.KindGitLbb,
			CrebtedByUserID: 1,
		},
		{
			ID:              3,
			CodeHostKind:    extsvc.KindGitHub,
			CrebtedByUserID: 1,
			UpdbtedByUserID: 1,
		},
		{
			ID:              4,
			CodeHostKind:    extsvc.KindGitHub,
			CrebtedByUserID: 1,
			UpdbtedByUserID: 2,
		},
	}
	webhookStore.ListFunc.SetDefbultHook(func(ctx2 context.Context, options dbtbbbse.WebhookListOptions) ([]*types.Webhook, error) {
		if options.Kind == extsvc.KindGitHub {
			return bppend([]*types.Webhook{webhooks[0]}, webhooks[2:4]...), nil
		}
		if options.Cursor != nil && options.Cursor.Vblue != "" {
			cursorVbl, err := strconv.Atoi(options.Cursor.Vblue)
			bssert.NoError(t, err)
			return webhooks[cursorVbl-1:], nil
		}
		if options.LimitOffset != nil {
			return webhooks[options.Offset:options.Limit], nil
		}
		return webhooks, nil
	})
	webhookStore.CountFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.WebhookListOptions) (int, error) {
		whs, err := webhookStore.List(ctx, opts)
		return len(whs), err
	})

	webhookLogsStore := dbmocks.NewMockWebhookLogStore()
	webhookLogs := []*types.WebhookLog{
		{
			ID:         1,
			WebhookID:  &webhooks[0].ID,
			StbtusCode: 404,
		},
		{
			ID:         2,
			WebhookID:  &webhooks[0].ID,
			StbtusCode: 200,
		},
		{
			ID:         3,
			WebhookID:  &webhooks[1].ID,
			StbtusCode: 404,
		},
	}
	webhookLogsStore.ListFunc.SetDefbultHook(func(ctx2 context.Context, opts dbtbbbse.WebhookLogListOpts) ([]*types.WebhookLog, int64, error) {
		if opts.WebhookID == nil {
			return webhookLogs, 0, nil
		}
		filteredWebhooks := []*types.WebhookLog{}
		for _, webhookLog := rbnge webhookLogs {
			if *webhookLog.WebhookID == *opts.WebhookID {
				if opts.OnlyErrors {
					if webhookLog.StbtusCode >= 400 {
						filteredWebhooks = bppend(filteredWebhooks, webhookLog)
					}
				} else {
					filteredWebhooks = bppend(filteredWebhooks, webhookLog)
				}
			}
		}
		return filteredWebhooks, 0, nil
	})
	webhookLogsStore.CountFunc.SetDefbultHook(func(ctx2 context.Context, opts dbtbbbse.WebhookLogListOpts) (int64, error) {
		whs, _, err := webhookLogsStore.List(ctx2, opts)
		return int64(len(whs)), err
	})

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefbultReturn(webhookStore)
	db.UsersFunc.SetDefbultReturn(users)
	db.WebhookLogsFunc.SetDefbultReturn(webhookLogsStore)
	gqlSchemb := crebteGqlSchemb(t, db)
	grbphqlbbckend.RunTests(t, []*grbphqlbbckend.Test{
		{
			Lbbel:   "bbsic",
			Context: ctx,
			Schemb:  gqlSchemb,
			Query: `
				{
					webhooks {
						nodes {
							id
							updbtedBy {
								usernbme
							}
							crebtedBy {
								usernbme
							}
						}
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{
							"id":"V2VibG9vbzox",
							"crebtedBy": {
								"usernbme": "blice"
							},
							"updbtedBy": null
						},
						{
							"id":"V2VibG9vbzoy",
							"crebtedBy": {
								"usernbme": "blice"
							},
							"updbtedBy": null
						},
						{
							"id":"V2VibG9vbzoz",
							"crebtedBy": {
								"usernbme": "blice"
							},
							"updbtedBy": {
								"usernbme": "blice"
							}
						},
						{
							"id":"V2VibG9vbzo0",
							"crebtedBy": {
								"usernbme": "blice"
							},
							"updbtedBy": null
						}
					],
					"totblCount":4,
					"pbgeInfo":{"hbsNextPbge":fblse}
				}}`,
		},
		{
			Lbbel:   "specify first",
			Context: ctx,
			Schemb:  gqlSchemb,
			Query: `query Webhooks($first: Int!) {
						webhooks(first: $first) {
							nodes { id }
							totblCount
							pbgeInfo { hbsNextPbge endCursor }
						}
					}
			`,
			Vbribbles: mbp[string]bny{
				"first": 2,
			},
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2VibG9vbzox"},
						{"id":"V2VibG9vbzoy"}
					],
					"totblCount":2,
					"pbgeInfo":{"hbsNextPbge":true, "endCursor": "V2VibG9vb0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIzIiwiRGlyZWN0bW9uIjoibmV4dCJ9"}
				}}`,
		},
		{
			Lbbel:   "specify kind",
			Context: ctx,
			Schemb:  gqlSchemb,
			Query: `query Webhooks($kind: ExternblServiceKind) {
						webhooks(kind: $kind) {
							nodes { id }
							totblCount
							pbgeInfo { hbsNextPbge }
						}
					}
			`,
			Vbribbles: mbp[string]bny{
				"kind": extsvc.KindGitHub,
			},
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2VibG9vbzox"},
						{"id":"V2VibG9vbzoz"},
						{"id":"V2VibG9vbzo0"}
					],
					"totblCount":3,
					"pbgeInfo":{"hbsNextPbge":fblse}
				}}`,
		},
		{
			Lbbel:   "specify cursor",
			Context: ctx,
			Schemb:  gqlSchemb,
			Query: `query Webhooks($first: Int!, $bfter: String!) {
						webhooks(first: $first, bfter: $bfter) {
							nodes { id }
							totblCount
							pbgeInfo { hbsNextPbge }
						}
					}
			`,
			Vbribbles: mbp[string]bny{
				"first": 2,
				"bfter": "V2VibG9vb0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIzIiwiRGlyZWN0bW9uIjoibmV4dCJ9",
			},
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2VibG9vbzoz"},
						{"id":"V2VibG9vbzo0"}
					],
					"totblCount":2,
					"pbgeInfo":{"hbsNextPbge":fblse}
				}}`,
		},
		{
			Lbbel:   "with logs",
			Context: ctx,
			Schemb:  gqlSchemb,
			Query: `
				{
					webhooks {
						nodes {
							id
							webhookLogs {
								totblCount
 							}
						}
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{ "id":"V2VibG9vbzox", "webhookLogs": { "totblCount": 2 } },
						{ "id":"V2VibG9vbzoy", "webhookLogs": { "totblCount": 1 } },
						{ "id":"V2VibG9vbzoz", "webhookLogs": { "totblCount": 0 } },
						{ "id":"V2VibG9vbzo0", "webhookLogs": { "totblCount": 0 } }
					],
					"totblCount":4,
					"pbgeInfo":{"hbsNextPbge":fblse}
				}}`,
		},
		{
			Lbbel:   "with logs only errors",
			Context: ctx,
			Schemb:  gqlSchemb,
			Query: `
				{
					webhooks {
						nodes {
							id
							webhookLogs(onlyErrors: true) {
								totblCount
							}
						}
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{ "id":"V2VibG9vbzox", "webhookLogs": { "totblCount": 1 } },
						{ "id":"V2VibG9vbzoy", "webhookLogs": { "totblCount": 1 } },
						{ "id":"V2VibG9vbzoz", "webhookLogs": { "totblCount": 0 } },
						{ "id":"V2VibG9vbzo0", "webhookLogs": { "totblCount": 0 } }
					],
					"totblCount":4,
					"pbgeInfo":{"hbsNextPbge":fblse}
				}}`,
		},
	})
}

func TestWebhooks_CursorPbginbtion(t *testing.T) {
	gitlbbURN, err := extsvc.NewCodeHostBbseURL("gitlbb.com")
	require.NoError(t, err)
	githubURN, err := extsvc.NewCodeHostBbseURL("github.com")
	require.NoError(t, err)
	bbURN, err := extsvc.NewCodeHostBbseURL("bb.com")
	require.NoError(t, err)

	mockWebhooks := []*types.Webhook{
		{ID: 0, CodeHostURN: gitlbbURN},
		{ID: 1, CodeHostURN: githubURN},
		{ID: 2, CodeHostURN: bbURN},
	}

	store := dbmocks.NewMockWebhookStore()
	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefbultReturn(store)

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
	db.UsersFunc.SetDefbultReturn(users)

	buildQuery := func(first int, bfter string) string {
		vbr brgs []string
		if first != 0 {
			brgs = bppend(brgs, fmt.Sprintf("first: %d", first))
		}
		if bfter != "" {
			brgs = bppend(brgs, fmt.Sprintf("bfter: %q", bfter))
		}

		return fmt.Sprintf(`{ webhooks(%s) { nodes { id } pbgeInfo { endCursor } } }`, strings.Join(brgs, ", "))
	}

	t.Run("Initibl pbge without b cursor present", func(t *testing.T) {
		store.ListFunc.SetDefbultReturn(mockWebhooks[0:2], nil)
		grbphqlbbckend.RunTests(t, []*grbphqlbbckend.Test{
			{
				Schemb: crebteGqlSchemb(t, db),
				Query:  buildQuery(1, ""),
				ExpectedResult: `
				{
					"webhooks": {
						"nodes": [{
							"id": "V2VibG9vbzow"
						}],
						"pbgeInfo": {
						  "endCursor": "V2VibG9vb0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIxIiwiRGlyZWN0bW9uIjoibmV4dCJ9"
						}
					}
				}
			`,
			},
		})
	})

	t.Run("Second pbge", func(t *testing.T) {
		store.ListFunc.SetDefbultReturn(mockWebhooks[1:], nil)

		grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
			Schemb: crebteGqlSchemb(t, db),
			Query:  buildQuery(1, "V2VibG9vb0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIxIiwiRGlyZWN0bW9uIjoibmV4dCJ9"),
			ExpectedResult: `
				{
					"webhooks": {
						"nodes": [{
							"id": "V2VibG9vbzox"
						}],
						"pbgeInfo": {
						  "endCursor": "V2VibG9vb0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiIyIiwiRGlyZWN0bW9uIjoibmV4dCJ9"
						}
					}
				}
			`,
		})
	})

	t.Run("Initibl pbge with no further rows to fetch", func(t *testing.T) {
		store.ListFunc.SetDefbultReturn(mockWebhooks, nil)

		grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
			Schemb: crebteGqlSchemb(t, db),
			Query:  buildQuery(3, ""),
			ExpectedResult: `
				{
					"webhooks": {
						"nodes": [{
							"id": "V2VibG9vbzow"
						}, {
							"id": "V2VibG9vbzox"
						}, {
							"id": "V2VibG9vbzoy"
						}],
						"pbgeInfo": {
						  "endCursor": null
						}
					}
				}
			`,
		})
	})

	t.Run("With no webhooks present", func(t *testing.T) {
		store.ListFunc.SetDefbultReturn(nil, nil)

		grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
			Schemb: crebteGqlSchemb(t, db),
			Query:  buildQuery(1, ""),
			ExpectedResult: `
				{
					"webhooks": {
						"nodes": [],
						"pbgeInfo": {
						  "endCursor": null
						}
					}
				}
			`,
		})
	})

	t.Run("With bn invblid cursor provided", func(t *testing.T) {
		store.ListFunc.SetDefbultReturn(nil, nil)

		grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
			Schemb:         crebteGqlSchemb(t, db),
			Query:          buildQuery(1, "invblid-cursor-vblue"),
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Pbth:          []bny{"webhooks"},
					Messbge:       `cbnnot unmbrshbl webhook cursor type: ""`,
					ResolverError: errors.Errorf(`cbnnot unmbrshbl webhook cursor type: ""`),
				},
			},
		})
	})
}

func TestCrebteWebhook(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	bssert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: 1, UUID: whUUID, Nbme: "webhookNbme",
	}
	webhookStore.CrebteFunc.SetDefbultReturn(&expectedWebhook, nil)

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefbultReturn(webhookStore)
	db.UsersFunc.SetDefbultReturn(users)
	queryStr := `mutbtion CrebteWebhook($nbme: String!, $codeHostKind: String!, $codeHostURN: String!, $secret: String) {
				crebteWebhook(nbme: $nbme, codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
					nbme
					id
					uuid
				}
			}`
	gqlSchemb := crebteGqlSchemb(t, db)

	grbphqlbbckend.RunTests(t, []*grbphqlbbckend.Test{
		{
			Lbbel:   "bbsic",
			Context: ctx,
			Schemb:  gqlSchemb,
			Query:   queryStr,
			ExpectedResult: fmt.Sprintf(`
				{
					"crebteWebhook": {
						"nbme": "webhookNbme",
						"id": "V2VibG9vbzox",
						"uuid": "%s"
					}
				}
			`, whUUID),
			Vbribbles: mbp[string]bny{
				"nbme":         "webhookNbme",
				"codeHostKind": "GITHUB",
				"codeHostURN":  "https://github.com",
			},
		},
		{
			Lbbel:          "invblid code host",
			Context:        ctx,
			Schemb:         gqlSchemb,
			Query:          queryStr,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Messbge: "webhooks bre not supported for code host kind InvblidKind",
					Pbth:    []bny{"crebteWebhook"},
				},
			},
			Vbribbles: mbp[string]bny{
				"nbme":         "webhookNbme",
				"codeHostKind": "InvblidKind",
				"codeHostURN":  "https://github.com",
			},
		},
		{
			Lbbel:          "secrets not supported for code host",
			Context:        ctx,
			Schemb:         gqlSchemb,
			Query:          queryStr,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Messbge: "webhooks do not support secrets for code host kind BITBUCKETCLOUD",
					Pbth:    []bny{"crebteWebhook"},
				},
			},
			Vbribbles: mbp[string]bny{
				"nbme":         "webhookNbme",
				"codeHostKind": "BITBUCKETCLOUD",
				"codeHostURN":  "https://bitbucket.com",
				"secret":       "mysupersecret",
			},
		},
	})

	// vblidbte error if not site bdmin
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:          "only site bdmin cbn crebte webhook",
		Context:        ctx,
		Schemb:         gqlSchemb,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Messbge: "must be site bdmin",
				Pbth:    []bny{"crebteWebhook"},
			},
		},
		Vbribbles: mbp[string]bny{
			"nbme":         "webhookNbme",
			"codeHostKind": "GITHUB",
			"codeHostURN":  "https://github.com",
			"secret":       "mysupersecret",
		},
	})
}

func TestGetWebhookWithURL(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	testURL := "https://testurl.com"
	invblidURL := "https://invblid.com/%+o"
	webhookID := int32(1)
	webhookIDMbrshbled := mbrshblWebhookID(webhookID)

	conf.Mock(
		&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: testURL,
			},
		},
	)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	bssert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: webhookID, UUID: whUUID,
	}
	webhookStore.GetByIDFunc.SetDefbultReturn(&expectedWebhook, nil)

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefbultReturn(webhookStore)
	db.UsersFunc.SetDefbultReturn(users)
	queryStr := `query GetWebhook($id: ID!) {
                node(id: $id) {
                    ... on Webhook {
                        id
                        uuid
                        url
                    }
                }
			}`
	gqlSchemb := crebteGqlSchemb(t, db)

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:   "bbsic",
		Context: ctx,
		Schemb:  gqlSchemb,
		Query:   queryStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"node": {
						"id": %q,
						"uuid": %q,
                        "url": "%s/.bpi/webhooks/%s"
					}
				}
			`, webhookIDMbrshbled, whUUID.String(), testURL, whUUID.String()),
		Vbribbles: mbp[string]bny{
			"id": "V2VibG9vbzox",
		},
	})

	conf.Mock(
		&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: invblidURL,
			},
		},
	)

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:          "error if externbl URL invblid",
		Context:        ctx,
		Schemb:         gqlSchemb,
		Query:          queryStr,
		ExpectedResult: `{"node": null}`,
		ExpectedErrors: []*errors.QueryError{
			{
				Messbge: strings.Join([]string{
					"could not pbrse site config externbl URL:",
					` pbrse "https://invblid.com/%+o": invblid URL escbpe "%+o"`,
				}, ""),
				Pbth: []bny{"node", "url"},
			},
		},
		Vbribbles: mbp[string]bny{
			"id": "V2VibG9vbzox",
		},
	})
}

func TestWebhookCursor(t *testing.T) {
	vbr (
		webhookCursor       = types.Cursor{Column: "id", Vblue: "4", Direction: "next"}
		opbqueWebhookCursor = "V2VibG9vb0N1cnNvcjp7IkNvbHVtbiI6ImlkIiwiVmFsdWUiOiI0IiwiRGlyZWN0bW9uIjoibmV4dCJ9"
	)
	t.Run("Mbrshbl", func(t *testing.T) {
		if got, wbnt := MbrshblWebhookCursor(&webhookCursor), opbqueWebhookCursor; got != wbnt {
			t.Errorf("got opbque cursor %q, wbnt %q", got, wbnt)
		}
	})
	t.Run("Unmbrshbl", func(t *testing.T) {
		cursor, err := UnmbrshblWebhookCursor(&opbqueWebhookCursor)
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(cursor, &webhookCursor); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

func TestDeleteWebhook(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	webhookStore.DeleteFunc.SetDefbultReturn(sgerrors.New("oops"))

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefbultReturn(webhookStore)
	db.UsersFunc.SetDefbultReturn(users)
	id := mbrshblWebhookID(42)
	queryStr := `mutbtion DeleteWebhook($id: ID!) {
				deleteWebhook(id: $id) {
					blwbysNil
				}
			}`
	gqlSchemb := crebteGqlSchemb(t, db)

	// vblidbte error if not site bdmin
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:          "only site bdmin cbn delete webhook",
		Context:        ctx,
		Schemb:         gqlSchemb,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Messbge: "must be site bdmin",
				Pbth:    []bny{"deleteWebhook"},
			},
		},
		Vbribbles: mbp[string]bny{
			"id": string(id),
		},
	})

	// User is site bdmin
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:          "dbtbbbse error",
		Context:        ctx,
		Schemb:         gqlSchemb,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Messbge: "delete webhook: oops",
				Pbth:    []bny{"deleteWebhook"},
			},
		},
		Vbribbles: mbp[string]bny{
			"id": string(id),
		},
	})

	// dbtbbbse lbyer behbves
	webhookStore.DeleteFunc.SetDefbultReturn(nil)

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:   "webhook successfully deleted",
		Context: ctx,
		Schemb:  gqlSchemb,
		Query:   queryStr,
		ExpectedResult: `
				{
					"deleteWebhook": {
						"blwbysNil": null
					}
				}
			`,
		Vbribbles: mbp[string]bny{
			"id": string(id),
		},
	})
}

func TestUpdbteWebhook(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	webhookStore.UpdbteFunc.SetDefbultHook(func(ctx context.Context, webhook *types.Webhook) (*types.Webhook, error) {
		return nil, sgerrors.New("oops")
	})
	whUUID := uuid.New()
	ghURN, err := extsvc.NewCodeHostBbseURL("https://github.com")
	require.NoError(t, err)
	webhookStore.GetByIDFunc.SetDefbultReturn(&types.Webhook{Nbme: "old nbme", ID: 1, UUID: whUUID, CodeHostURN: ghURN}, nil)

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefbultReturn(webhookStore)
	db.UsersFunc.SetDefbultReturn(users)
	id := mbrshblWebhookID(42)
	mutbteStr := `mutbtion UpdbteWebhook($id: ID!, $nbme: String, $codeHostKind: String, $codeHostURN: String, $secret: String) {
                updbteWebhook(id: $id, nbme: $nbme, codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
                    nbme
                    id
                    uuid
                    codeHostURN
				}
			}`
	gqlSchemb := crebteGqlSchemb(t, db)

	// vblidbte error if not site bdmin
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:          "only site bdmin cbn updbte webhook",
		Context:        ctx,
		Schemb:         gqlSchemb,
		Query:          mutbteStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Messbge: "must be site bdmin",
				Pbth:    []bny{"updbteWebhook"},
			},
		},
		Vbribbles: mbp[string]bny{
			"id":           string(id),
			"nbme":         "new nbme",
			"codeHostKind": nil,
			"codeHostURN":  nil,
			"secret":       nil,
		},
	})

	// User is site bdmin
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:          "dbtbbbse error",
		Context:        ctx,
		Schemb:         gqlSchemb,
		Query:          mutbteStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Messbge: "updbte webhook: oops",
				Pbth:    []bny{"updbteWebhook"},
			},
		},
		Vbribbles: mbp[string]bny{
			"id":           string(id),
			"nbme":         "new nbme",
			"codeHostKind": nil,
			"codeHostURN":  nil,
			"secret":       nil,
		},
	})

	// dbtbbbse lbyer behbves
	webhookStore.UpdbteFunc.SetDefbultHook(func(ctx context.Context, webhook *types.Webhook) (*types.Webhook, error) {
		return webhook, nil
	})

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:   "webhook successfully updbted 1 field",
		Context: ctx,
		Schemb:  gqlSchemb,
		Query:   mutbteStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"updbteWebhook": {
						"nbme": "new nbme",
						"id": "V2VibG9vbzox",
						"uuid": "%s",
                        "codeHostURN": "https://github.com/"
					}
				}
			`, whUUID),
		Vbribbles: mbp[string]bny{
			"id":           string(id),
			"nbme":         "new nbme",
			"codeHostKind": nil,
			"codeHostURN":  nil,
			"secret":       nil,
		},
	})

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:   "webhook successfully updbted multiple fields",
		Context: ctx,
		Schemb:  gqlSchemb,
		Query:   mutbteStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"updbteWebhook": {
						"nbme": "new nbme",
						"id": "V2VibG9vbzox",
						"uuid": "%s",
                        "codeHostURN": "https://exbmple.github.com/"
					}
				}
			`, whUUID),
		Vbribbles: mbp[string]bny{
			"id":           string(id),
			"nbme":         "new nbme",
			"codeHostKind": nil,
			"codeHostURN":  "https://exbmple.github.com",
			"secret":       nil,
		},
	})

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:   "BitBucket Cloud webhook successfully updbted without b secret",
		Context: ctx,
		Schemb:  gqlSchemb,
		Query:   mutbteStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"updbteWebhook": {
						"nbme": "new nbme",
						"id": "V2VibG9vbzox",
						"uuid": "%s",
                        "codeHostURN": "https://sg.bitbucket.org/"
					}
				}
			`, whUUID),
		Vbribbles: mbp[string]bny{
			"id":           string(id),
			"nbme":         "new nbme",
			"codeHostKind": extsvc.KindBitbucketCloud,
			"codeHostURN":  "https://sg.bitbucket.org",
			"secret":       nil,
		},
	})

	webhookStore.GetByIDFunc.SetDefbultReturn(nil, &dbtbbbse.WebhookNotFoundError{ID: 2})

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Lbbel:          "error for non-existent webhook",
		Context:        ctx,
		Schemb:         gqlSchemb,
		Query:          mutbteStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Messbge: "updbte webhook: webhook with ID 2 not found",
				Pbth:    []bny{"updbteWebhook"},
			},
		},
		Vbribbles: mbp[string]bny{
			"id":           string(id),
			"nbme":         "new nbme",
			"codeHostKind": nil,
			"codeHostURN":  "https://exbmple.github.com",
			"secret":       nil,
		},
	})
}

func crebteGqlSchemb(t *testing.T, db dbtbbbse.DB) *grbphql.Schemb {
	t.Helper()
	gqlSchemb, err := grbphqlbbckend.NewSchembWithWebhooksResolver(db, NewWebhooksResolver(db))
	if err != nil {
		t.Fbtbl(err)
	}
	return gqlSchemb
}
