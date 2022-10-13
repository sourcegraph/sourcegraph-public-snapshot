package graphqlbackend

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestListWebhooks(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := database.NewMockWebhookStore()
	//webhookStore.ListFunc.SetDefaultReturn(expectedWebhooks, nil)
	webhooks := []*types.Webhook{
		{
			ID:           1,
			CodeHostKind: extsvc.KindGitHub,
		},
		{
			ID:           2,
			CodeHostKind: extsvc.KindGitLab,
		},
		{
			ID:           3,
			CodeHostKind: extsvc.KindGitHub,
		},
		{
			ID:           4,
			CodeHostKind: extsvc.KindGitHub,
		},
	}
	webhookStore.ListFunc.SetDefaultHook(func(ctx2 context.Context, options database.WebhookListOptions) ([]*types.Webhook, error) {
		if options.Kind == extsvc.KindGitHub {
			return append([]*types.Webhook{webhooks[0]}, webhooks[1:3]...), nil
		}
		if options.LimitOffset != nil {
			return webhooks[options.Offset:options.Limit], nil
		}
		return webhooks, nil
	})

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	schema := mustParseGraphQLSchema(t, db)
	RunTests(t, []*Test{
		{
			Label:   "basic",
			Context: ctx,
			Schema:  schema,
			Query: `
				{
					webhooks {
						nodes { id }
						totalCount
						pageInfo { hasNextPage }
					}
				}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2ViaG9vazox"},
						{"id":"V2ViaG9vazoy"},
						{"id":"V2ViaG9vazoz"},
						{"id":"V2ViaG9vazo0"}
					],
					"totalCount":4,
					"pageInfo":{"hasNextPage":false}
				}}`,
		},
		{
			Label:   "specify first",
			Context: ctx,
			Schema:  schema,
			Query: `query Webhooks($first: Int!) {
						webhooks(first: $first) {
							nodes { id }
							totalCount
							pageInfo { hasNextPage }
						}
					}
			`,
			ExpectedResult: `{"webhooks":
				{
					"nodes":[
						{"id":"V2ViaG9vazox"},
						{"id":"V2ViaG9vazoy"}
					],
					"totalCount":2,
					"pageInfo":{"hasNextPage":false}
				}}`,
			Variables: map[string]any{
				"first": 2,
			},
		},
	})
}

func TestCreateWebhook(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := database.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	assert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: 1, UUID: whUUID,
	}
	webhookStore.CreateFunc.SetDefaultReturn(&expectedWebhook, nil)

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	queryStr := `mutation CreateWebhook($codeHostKind: String!, $codeHostURN: String!, $secret: String) {
				createWebhook(codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
					id
					uuid
				}
			}`
	schema := mustParseGraphQLSchema(t, db)

	RunTests(t, []*Test{
		{
			Label:   "basic",
			Context: ctx,
			Schema:  schema,
			Query:   queryStr,
			ExpectedResult: fmt.Sprintf(`
				{
					"createWebhook": {
						"id": "V2ViaG9vazox",
						"uuid": "%s"
					}
				}
			`, whUUID),
			Variables: map[string]any{
				"codeHostKind": "GITHUB",
				"codeHostURN":  "https://github.com",
			},
		},
		{
			Label:          "invalid code host",
			Context:        ctx,
			Schema:         schema,
			Query:          queryStr,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "webhooks are not supported for code host kind InvalidKind",
					Path:    []any{"createWebhook"},
				},
			},
			Variables: map[string]any{
				"codeHostKind": "InvalidKind",
				"codeHostURN":  "https://github.com",
			},
		},
		{
			Label:          "secrets not supported for code host",
			Context:        ctx,
			Schema:         schema,
			Query:          queryStr,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "webhooks do not support secrets for code host kind BITBUCKETCLOUD",
					Path:    []any{"createWebhook"},
				},
			},
			Variables: map[string]any{
				"codeHostKind": "BITBUCKETCLOUD",
				"codeHostURN":  "https://bitbucket.com",
				"secret":       "mysupersecret",
			},
		},
	})

	// validate error if not site admin
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)
	RunTest(t, &Test{
		Label:          "only site admin can create webhook",
		Context:        ctx,
		Schema:         schema,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "must be site admin",
				Path:    []any{"createWebhook"},
			},
		},
		Variables: map[string]any{
			"codeHostKind": "GITHUB",
			"codeHostURN":  "https://github.com",
			"secret":       "mysupersecret",
		},
	})
}

func TestGetWebhookWithURL(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	testURL := "https://testurl.com"
	invalidURL := "https://invalid.com/%+o"
	webhookID := int32(1)
	webhookIDMarshaled := marshalWebhookID(webhookID)
	conf.Mock(
		&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExternalURL: testURL,
			},
		},
	)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := database.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	assert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: webhookID, UUID: whUUID,
	}
	webhookStore.GetByIDFunc.SetDefaultReturn(&expectedWebhook, nil)

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	queryStr := `query GetWebhook($id: ID!) {
                node(id: $id) {
                    ... on Webhook {
                        id
                        uuid
                        url
                    }
                }
			}`
	gqlSchema := mustParseGraphQLSchema(t, db)

	RunTest(t, &Test{
		Label:   "basic",
		Context: ctx,
		Schema:  gqlSchema,
		Query:   queryStr,
		ExpectedResult: fmt.Sprintf(`
				{
					"node": {
						"id": %q,
						"uuid": %q,
                        "url": "%s/webhooks/%s"
					}
				}
			`, webhookIDMarshaled, whUUID.String(), testURL, whUUID.String()),
		Variables: map[string]any{
			"id": "V2ViaG9vazox",
		},
	})

	conf.Mock(
		&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExternalURL: invalidURL,
			},
		},
	)
	RunTest(t, &Test{
		Label:          "error if external URL invalid",
		Context:        ctx,
		Schema:         gqlSchema,
		Query:          queryStr,
		ExpectedResult: `{"node": null}`,
		ExpectedErrors: []*errors.QueryError{
			{
				Message: strings.Join([]string{
					"could not parse site config external URL:",
					` parse "https://invalid.com/%+o": invalid URL escape "%+o"`,
				}, ""),
				Path: []any{"node", "url"},
			},
		},
		Variables: map[string]any{
			"id": "V2ViaG9vazox",
		},
	})
}
