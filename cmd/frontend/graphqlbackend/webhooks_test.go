package graphqlbackend

import (
	"fmt"
	"testing"

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

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `mutation CreateWebhook($codeHostKind: String!, $codeHostURN: String!) {
				createWebhook(codeHostKind: $codeHostKind, codeHostURN: $codeHostURN) {
					id
					uuid
				}
			}`,
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
	})

}
