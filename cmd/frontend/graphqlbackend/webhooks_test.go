package graphqlbackend

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	sgerrors "github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

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

func TestDeleteWebhook(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := database.NewMockWebhookStore()
	webhookStore.DeleteFunc.SetDefaultReturn(sgerrors.New("oops"))

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)
	id := marshalWebhookID(42)
	queryStr := `mutation DeleteWebhook($id: ID!) {
				deleteWebhook(id: $id) {
					alwaysNil
				}
			}`
	schema := mustParseGraphQLSchema(t, db)

	// validate error if not site admin
	RunTest(t, &Test{
		Label:          "only site admin can delete webhook",
		Context:        ctx,
		Schema:         schema,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "must be site admin",
				Path:    []any{"deleteWebhook"},
			},
		},
		Variables: map[string]any{
			"id": string(id),
		},
	})

	// User is site admin
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	RunTest(t, &Test{
		Label:          "database error",
		Context:        ctx,
		Schema:         schema,
		Query:          queryStr,
		ExpectedResult: "null",
		ExpectedErrors: []*errors.QueryError{
			{
				Message: "delete webhook: oops",
				Path:    []any{"deleteWebhook"},
			},
		},
		Variables: map[string]any{
			"id": string(id),
		},
	})

	// database layer behaves
	webhookStore.DeleteFunc.SetDefaultReturn(nil)

	RunTest(t, &Test{
		Label:   "webhook successfully deleted",
		Context: ctx,
		Schema:  schema,
		Query:   queryStr,
		ExpectedResult: `
				{
					"deleteWebhook": {
						"alwaysNil": null
					}
				}
			`,
		Variables: map[string]any{
			"id": string(id),
		},
	})
}
