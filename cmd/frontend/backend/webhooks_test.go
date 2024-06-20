package backend

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCreateWebhook(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	webhookStore := dbmocks.NewMockWebhookStore()
	whUUID, err := uuid.NewUUID()
	assert.NoError(t, err)

	db := dbmocks.NewMockDB()
	db.WebhooksFunc.SetDefaultReturn(webhookStore)
	db.UsersFunc.SetDefaultReturn(users)

	ws := NewWebhookService(db, keyring.Default())

	rawURN := "https://github.com"
	// ghURN should be normalized and now include the trailing slash
	ghURN, err := extsvc.NewCodeHostBaseURL(rawURN)
	require.NoError(t, err)
	testSecret := "mysecret"
	tests := []struct {
		label        string
		name         string
		codeHostKind string
		codeHostURN  string
		secret       *string
		expected     types.Webhook
		expectedErr  error
	}{
		{
			label:        "basic",
			name:         "webhook name",
			codeHostKind: extsvc.KindGitHub,
			codeHostURN:  rawURN,
			secret:       &testSecret,
			expected: types.Webhook{
				ID:              1,
				Name:            "webhook name",
				UUID:            whUUID,
				CodeHostKind:    extsvc.KindGitHub,
				CodeHostURN:     ghURN,
				Secret:          nil,
				CreatedByUserID: 0,
			},
		},
		{
			label:        "basic with trailing slash",
			name:         "webhook name 2",
			codeHostKind: extsvc.KindGitHub,
			codeHostURN:  rawURN + "/",
			secret:       &testSecret,
			expected: types.Webhook{
				ID:              2,
				Name:            "webhook name 2",
				UUID:            whUUID,
				CodeHostKind:    extsvc.KindGitHub,
				CodeHostURN:     ghURN,
				Secret:          nil,
				CreatedByUserID: 0,
			},
		},
		{
			label:        "invalid code host",
			codeHostKind: "InvalidKind",
			codeHostURN:  rawURN,
			expectedErr:  errors.New("webhooks are not supported for code host kind InvalidKind"),
		},
		{
			label:        "secrets are not supported for code host",
			codeHostKind: extsvc.KindAzureDevOps,
			secret:       &testSecret,
			expectedErr:  errors.New("webhooks do not support secrets for code host kind AZUREDEVOPS"),
		},
	}

	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			webhookStore.CreateFunc.SetDefaultReturn(&test.expected, nil)
			webhookStore.CreateFunc.SetDefaultHook(func(_ context.Context, _, _, _ string, _ int32, secret *encryption.Encryptable) (*types.Webhook, error) {
				if test.secret != nil {
					assert.NotZero(t, secret)
				}
				return &test.expected, nil
			})
			wh, err := ws.CreateWebhook(ctx, test.name, test.codeHostKind, test.codeHostURN, test.secret)
			if test.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, test.expected.ID, wh.ID)
				assert.Equal(t, test.expected.Name, wh.Name)
				assert.Equal(t, test.expected.CodeHostKind, wh.CodeHostKind)
				assert.Equal(t, test.expected.UUID, wh.UUID)
				assert.Equal(t, test.expected.Secret, wh.Secret)
				assert.Equal(t, test.expected.CreatedByUserID, wh.CreatedByUserID)
			} else {
				assert.Equal(t, err.Error(), test.expectedErr.Error())
			}
		})
	}
}

func TestCreateUpdateDeleteWebhook(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	newUser := database.NewUser{Username: "testUser"}
	createdUser, err := db.Users().Create(ctx, newUser)
	require.NoError(t, err)
	require.NoError(t, db.Users().SetIsSiteAdmin(ctx, createdUser.ID, true))

	ctx = actor.WithActor(ctx, &actor.Actor{UID: createdUser.ID})

	ws := NewWebhookService(db, keyring.Default())

	// Create webhook
	secret := "12345"
	webhook, err := ws.CreateWebhook(ctx, "github", extsvc.KindGitHub, "https://github.com", &secret)
	require.NoError(t, err)
	assert.Equal(t, "github", webhook.Name)
	assert.Equal(t, extsvc.KindGitHub, webhook.CodeHostKind)
	assert.Equal(t, "https://github.com/", webhook.CodeHostURN.String())
	whSecret, err := webhook.Secret.Decrypt(ctx)
	require.NoError(t, err)
	assert.Equal(t, secret, whSecret)

	// Update webhook
	newSecret := "54321"
	newName := "new name"
	newCodeHostKind := extsvc.KindGitLab
	newCodeHostURL := "https://gitlab.com/"
	newWebhook, err := ws.UpdateWebhook(ctx, webhook.ID, newName, newCodeHostKind, newCodeHostURL, &newSecret)
	require.NoError(t, err)
	// assert that it's still the same webhook
	assert.Equal(t, webhook.ID, newWebhook.ID)
	assert.Equal(t, webhook.UUID, newWebhook.UUID)
	// assert values updated correctly
	assert.Equal(t, newName, newWebhook.Name)
	assert.Equal(t, newCodeHostKind, newWebhook.CodeHostKind)
	assert.Equal(t, newCodeHostURL, newWebhook.CodeHostURN.String())
	newWHSecret, err := newWebhook.Secret.Decrypt(ctx)
	require.NoError(t, err)
	assert.Equal(t, newSecret, newWHSecret)

	// Delete webhook
	err = ws.DeleteWebhook(ctx, webhook.ID)
	require.NoError(t, err)
	// assert webhook no longer exists
	deletedWH, err := db.Webhooks(keyring.Default().WebhookKey).GetByID(ctx, webhook.ID)
	assert.Nil(t, deletedWH)
	assert.Error(t, err)
}
