package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
