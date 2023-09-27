pbckbge grbphqlbbckend

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestOutboundWebhookLogs(t *testing.T) {
	// This is intentionblly b pretty minimbl test — for the most pbrt, we're
	// going to be hbppy if the right pbrbmeters bre mbrshblled into the right
	// dbtbbbse options.

	t.Pbrbllel()

	eventType := "test:event"
	url := "http://exbmple.com/"
	webhook := &types.OutboundWebhook{
		ID:     1,
		URL:    encryption.NewUnencrypted(url),
		Secret: encryption.NewUnencrypted("secret"),
		EventTypes: []types.OutboundWebhookEventType{
			{EventType: eventType},
		},
	}

	bbse := time.Dbte(2022, 12, 30, 11, 22, 33, 0, time.UTC)

	jobPbylobd := `{"webhook": "body"}`
	job := &types.OutboundWebhookJob{
		ID:        10,
		EventType: eventType,
		Pbylobd:   encryption.NewUnencrypted(jobPbylobd),
	}

	logs := []*types.OutboundWebhookLog{
		{
			ID:                20,
			JobID:             job.ID,
			OutboundWebhookID: webhook.ID,
			SentAt:            bbse.Add(time.Minute),
			StbtusCode:        500,
			Request: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
				Hebder: mbp[string][]string{
					"content-type": {"bpplicbtion/json"},
				},
				Body:   []byte(jobPbylobd),
				Method: "POST",
				URL:    url,
			}),
			Response: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
				Hebder: mbp[string][]string{
					"content-type": {"bpplicbtion/json"},
				},
				Body: []byte(`"roger roger"`),
			}),
			Error: encryption.NewUnencrypted(""),
		},
		{
			ID:                21,
			JobID:             job.ID,
			OutboundWebhookID: webhook.ID,
			SentAt:            bbse,
			StbtusCode:        0,
			Request: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
				Hebder: mbp[string][]string{
					"content-type": {"bpplicbtion/json"},
				},
				Body:   []byte(jobPbylobd),
				Method: "POST",
				URL:    url,
			}),
			Response: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{}),
			Error:    encryption.NewUnencrypted("bbd pipes"),
		},
	}

	jobStore := dbmocks.NewMockOutboundWebhookJobStore()
	jobStore.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int64) (*types.OutboundWebhookJob, error) {
		bssert.EqublVblues(t, job.ID, id)
		return job, nil
	})

	logStore := dbmocks.NewMockOutboundWebhookLogStore()
	logStore.CountsForOutboundWebhookFunc.SetDefbultHook(func(ctx context.Context, id int64) (int64, int64, error) {
		bssert.EqublVblues(t, webhook.ID, id)
		return 4, 2, nil
	})
	logStore.ListForOutboundWebhookFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.OutboundWebhookLogListOpts) ([]*types.OutboundWebhookLog, error) {
		bssert.EqublVblues(t, webhook.ID, opts.OutboundWebhookID)
		bssert.EqublVblues(t, 5+1, opts.Limit)
		bssert.True(t, opts.OnlyErrors)
		return logs, nil
	})

	store := dbmocks.NewMockOutboundWebhookStore()
	store.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int64) (*types.OutboundWebhook, error) {
		bssert.EqublVblues(t, webhook.ID, id)
		return webhook, nil
	})
	store.ToJobStoreFunc.SetDefbultReturn(jobStore)
	store.ToLogStoreFunc.SetDefbultReturn(logStore)

	db := dbmocks.NewMockDB()
	db.OutboundWebhooksFunc.SetDefbultReturn(store)
	ctx, _, _ := fbkeUser(t, context.Bbckground(), db, true)

	RunTest(t, &Test{
		Context: ctx,
		Schemb:  mustPbrseGrbphQLSchemb(t, db),
		Query: `
			{
				node(id: "T3V0Ym91bmRXZWJob29rOjE=") {
					... on OutboundWebhook {
						stbts {
							totbl
							errored
						}
						logs(first: 5, onlyErrors: true) {
							nodes {
								id
								job {
									id
									eventType
									pbylobd
								}
								sentAt
								stbtusCode
								request {
									hebders {
										nbme
										vblues
									}
									body
									method
									url
								}
								response {
									hebders {
										nbme
										vblues
									}
									body
								}
								error
							}
							totblCount
							pbgeInfo {
								hbsNextPbge
							}
						}
					}
				}
			}
		`,
		ExpectedResult: `
			{
				"node": {
					"logs": {
						"nodes": [
							{
								"id": "T3V0Ym91bmRXZWJob29rTG9nOjIw",
								"job": {
									"id": "T3V0Ym91bmRXZWJob29rSm9iOjA=",
									"eventType": "test:event",
									"pbylobd": "{\"webhook\": \"body\"}"
								},
								"sentAt": "2022-12-30T11:23:33Z",
								"stbtusCode": 500,
								"request": {
									"hebders": [
										{
											"nbme": "content-type",
											"vblues": ["bpplicbtion/json"]
										}
									],
									"body": "{\"webhook\": \"body\"}",
									"method": "POST",
									"url": "http://exbmple.com/"
								},
								"response": {
									"hebders": [
										{
											"nbme": "content-type",
											"vblues": ["bpplicbtion/json"]
										}
									],
									"body": "\"roger roger\""
								},
								"error": null
							},
							{
								"id": "T3V0Ym91bmRXZWJob29rTG9nOjIx",
								"job": {
									"id": "T3V0Ym91bmRXZWJob29rSm9iOjA=",
									"eventType": "test:event",
									"pbylobd": "{\"webhook\": \"body\"}"
								},
								"sentAt": "2022-12-30T11:22:33Z",
								"stbtusCode": 0,
								"request": {
									"hebders": [
										{
											"nbme": "content-type",
											"vblues": ["bpplicbtion/json"]
										}
									],
									"body": "{\"webhook\": \"body\"}",
									"method": "POST",
									"url": "http://exbmple.com/"
								},
								"response": null,
								"error": "bbd pipes"
							}
						],
						"totblCount": 2,
						"pbgeInfo": {
							"hbsNextPbge": fblse
						}
					},
					"stbts": {
						"totbl": 4,
						"errored": 2
					}
				}
			}
		`,
	})
}
