package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestOutboundWebhookLogs(t *testing.T) {
	// This is intentionally a pretty minimal test — for the most part, we're
	// going to be happy if the right parameters are marshalled into the right
	// database options.

	t.Parallel()

	eventType := "test:event"
	url := "http://example.com/"
	webhook := &types.OutboundWebhook{
		ID:     1,
		URL:    encryption.NewUnencrypted(url),
		Secret: encryption.NewUnencrypted("secret"),
		EventTypes: []types.OutboundWebhookEventType{
			{EventType: eventType},
		},
	}

	base := time.Date(2022, 12, 30, 11, 22, 33, 0, time.UTC)

	jobPayload := `{"webhook": "body"}`
	job := &types.OutboundWebhookJob{
		ID:        10,
		EventType: eventType,
		Payload:   encryption.NewUnencrypted(jobPayload),
	}

	logs := []*types.OutboundWebhookLog{
		{
			ID:                20,
			JobID:             job.ID,
			OutboundWebhookID: webhook.ID,
			SentAt:            base.Add(time.Minute),
			StatusCode:        500,
			Request: types.NewUnencryptedWebhookLogMessage(types.WebhookLogMessage{
				Header: map[string][]string{
					"content-type": {"application/json"},
				},
				Body:   []byte(jobPayload),
				Method: "POST",
				URL:    url,
			}),
			Response: types.NewUnencryptedWebhookLogMessage(types.WebhookLogMessage{
				Header: map[string][]string{
					"content-type": {"application/json"},
				},
				Body: []byte(`"roger roger"`),
			}),
			Error: encryption.NewUnencrypted(""),
		},
		{
			ID:                21,
			JobID:             job.ID,
			OutboundWebhookID: webhook.ID,
			SentAt:            base,
			StatusCode:        0,
			Request: types.NewUnencryptedWebhookLogMessage(types.WebhookLogMessage{
				Header: map[string][]string{
					"content-type": {"application/json"},
				},
				Body:   []byte(jobPayload),
				Method: "POST",
				URL:    url,
			}),
			Response: types.NewUnencryptedWebhookLogMessage(types.WebhookLogMessage{}),
			Error:    encryption.NewUnencrypted("bad pipes"),
		},
	}

	jobStore := dbmocks.NewMockOutboundWebhookJobStore()
	jobStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int64) (*types.OutboundWebhookJob, error) {
		assert.EqualValues(t, job.ID, id)
		return job, nil
	})

	logStore := dbmocks.NewMockOutboundWebhookLogStore()
	logStore.CountsForOutboundWebhookFunc.SetDefaultHook(func(ctx context.Context, id int64) (int64, int64, error) {
		assert.EqualValues(t, webhook.ID, id)
		return 4, 2, nil
	})
	logStore.ListForOutboundWebhookFunc.SetDefaultHook(func(ctx context.Context, opts database.OutboundWebhookLogListOpts) ([]*types.OutboundWebhookLog, error) {
		assert.EqualValues(t, webhook.ID, opts.OutboundWebhookID)
		assert.EqualValues(t, 5+1, opts.Limit)
		assert.True(t, opts.OnlyErrors)
		return logs, nil
	})

	store := dbmocks.NewMockOutboundWebhookStore()
	store.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int64) (*types.OutboundWebhook, error) {
		assert.EqualValues(t, webhook.ID, id)
		return webhook, nil
	})
	store.ToJobStoreFunc.SetDefaultReturn(jobStore)
	store.ToLogStoreFunc.SetDefaultReturn(logStore)

	db := dbmocks.NewMockDB()
	db.OutboundWebhooksFunc.SetDefaultReturn(store)
	ctx, _, _ := fakeUser(t, context.Background(), db, true)

	RunTest(t, &Test{
		Context: ctx,
		Schema:  mustParseGraphQLSchema(t, db),
		Query: `
			{
				node(id: "T3V0Ym91bmRXZWJob29rOjE=") {
					... on OutboundWebhook {
						stats {
							total
							errored
						}
						logs(first: 5, onlyErrors: true) {
							nodes {
								id
								job {
									id
									eventType
									payload
								}
								sentAt
								statusCode
								request {
									headers {
										name
										values
									}
									body
									method
									url
								}
								response {
									headers {
										name
										values
									}
									body
								}
								error
							}
							totalCount
							pageInfo {
								hasNextPage
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
									"payload": "{\"webhook\": \"body\"}"
								},
								"sentAt": "2022-12-30T11:23:33Z",
								"statusCode": 500,
								"request": {
									"headers": [
										{
											"name": "content-type",
											"values": ["application/json"]
										}
									],
									"body": "{\"webhook\": \"body\"}",
									"method": "POST",
									"url": "http://example.com/"
								},
								"response": {
									"headers": [
										{
											"name": "content-type",
											"values": ["application/json"]
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
									"payload": "{\"webhook\": \"body\"}"
								},
								"sentAt": "2022-12-30T11:22:33Z",
								"statusCode": 0,
								"request": {
									"headers": [
										{
											"name": "content-type",
											"values": ["application/json"]
										}
									],
									"body": "{\"webhook\": \"body\"}",
									"method": "POST",
									"url": "http://example.com/"
								},
								"response": null,
								"error": "bad pipes"
							}
						],
						"totalCount": 2,
						"pageInfo": {
							"hasNextPage": false
						}
					},
					"stats": {
						"total": 4,
						"errored": 2
					}
				}
			}
		`,
	})
}
