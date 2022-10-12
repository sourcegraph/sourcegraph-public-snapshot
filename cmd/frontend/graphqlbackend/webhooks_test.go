package graphqlbackend

import (
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCreateWebhook(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	webhookStore := database.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	assert.NoError(t, err)
	expectedWebhook := types.Webhook{
		ID: 1, UUID: whUUID,
	}
	webhookStore.CreateFunc.SetDefaultReturn(&expectedWebhook, nil)

	//authz := database.NewMockAuthzStore()
	//authz.GrantPendingPermissionsFunc.SetDefaultReturn(nil)

	db := database.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	//db.UsersFunc.SetDefaultReturn(users)
	//db.AuthzFunc.SetDefaultReturn(authz)
	queryStrNoSecret := `mutation CreateWebhook($codeHostKind: String!, $codeHostURN: String!) {
				createWebhook(codeHostKind: $codeHostKind, codeHostURN: $codeHostURN) {
					id
					uuid
				}
			}`
	queryStrWithSecret := `mutation CreateWebhook($codeHostKind: String!, $codeHostURN: String!, $secret: String) {
				createWebhook(codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
					id
					uuid
				}
			}`
	schema := mustParseGraphQLSchema(t, db)

	RunTests(t, []*Test{
		{
			Label:  "basic",
			Schema: schema,
			Query:  queryStrNoSecret,
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
				//"secret": "mysupersecret", // TODO: secret should not be required
			},
		},
		{
			Label:          "invalid code host",
			Schema:         schema,
			Query:          queryStrNoSecret,
			ExpectedResult: "null",
			ExpectedErrors: []*errors.QueryError{
				{
					Message: "Webhooks are not supported for code host kind InvalidKind",
					Path:    []any{"createWebhook"},
				},
			},
			Variables: map[string]any{
				"codeHostKind": "InvalidKind",
				"codeHostURN":  "https://github.com",
				//"secret": "mysupersecret", // TODO: secret should not be required
			},
		},
		{
			Label:          "secrets not supported for code host",
			Schema:         schema,
			Query:          queryStrWithSecret,
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

}
