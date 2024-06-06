package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	testSecret        = "my secret"
	testURN           = "https://github.com"
	githubWebhookName = "GitHub webhook"
	gitlabWebhookName = "GitLab webhook"
)

func TestWebhookCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	for _, encrypted := range []bool{true, false} {
		t.Run(fmt.Sprintf("encrypted=%t", encrypted), func(t *testing.T) {
			store := db.Webhooks(nil)
			if encrypted {
				store = db.Webhooks(et.ByteaTestKey{})
			}

			kind := extsvc.KindGitHub
			codeHostURL := "https://github.com/"
			encryptedSecret := types.NewUnencryptedSecret(testSecret)

			created, err := store.Create(ctx, githubWebhookName, kind, codeHostURL, 0, encryptedSecret)
			assert.NoError(t, err)

			// Check that the calculated fields were correctly calculated.
			assert.NotZero(t, created.ID)
			assert.NotZero(t, created.UUID)
			assert.Equal(t, githubWebhookName, created.Name)
			assert.Equal(t, kind, created.CodeHostKind)
			assert.Equal(t, codeHostURL, created.CodeHostURN.String())
			assert.Equal(t, int32(0), created.CreatedByUserID)
			assert.NotZero(t, created.CreatedAt)
			assert.NotZero(t, created.UpdatedAt)

			// getting the secret from the DB as is to verify its encryption
			row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", created.ID)
			var rawSecret string
			err = row.Scan(&rawSecret)
			assert.NoError(t, err)

			decryptedSecret, err := created.Secret.Decrypt(ctx)
			assert.NoError(t, err)

			if !encrypted {
				// if no encryption, raw secret stored in the db and decrypted secret should be the same
				assert.Equal(t, rawSecret, decryptedSecret)
			} else {
				// if encryption is specified, decrypted secret and raw secret should not match
				assert.NotEqual(t, rawSecret, decryptedSecret)
				assert.Equal(t, testSecret, decryptedSecret)
			}
		})
	}

	t.Run("no secret", func(t *testing.T) {
		store := db.Webhooks(et.ByteaTestKey{})

		kind := extsvc.KindGitHub
		codeHostURL := "https://github.com/"

		created, err := store.Create(ctx, githubWebhookName, kind, codeHostURL, 0, nil)
		assert.NoError(t, err)

		// Check that the calculated fields were correctly calculated.
		assert.NotZero(t, created.ID)
		assert.NotZero(t, created.UUID)
		assert.NoError(t, err)
		assert.Equal(t, githubWebhookName, created.Name)
		assert.Equal(t, kind, created.CodeHostKind)
		assert.Equal(t, codeHostURL, created.CodeHostURN.String())
		assert.Equal(t, int32(0), created.CreatedByUserID)
		assert.NotZero(t, created.CreatedAt)
		assert.NotZero(t, created.UpdatedAt)

		// secret in the DB should be null
		row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", created.ID)
		var rawSecret string
		err = row.Scan(&dbutil.NullString{S: &rawSecret})
		assert.NoError(t, err)
		assert.Zero(t, rawSecret)
	})
	t.Run("created by, updated by", func(t *testing.T) {
		webhooksStore := db.Webhooks(et.ByteaTestKey{})
		usersStore := db.Users()

		// First we need to create users, so they can be referenced from webhooks table
		user1, err := usersStore.Create(ctx, NewUser{Username: "user-1", Password: "user-1"})
		assert.NoError(t, err)
		UID1 := user1.ID
		user2, err := usersStore.Create(ctx, NewUser{Username: "user-2", Password: "user-2"})
		assert.NoError(t, err)
		UID2 := user2.ID

		// Creating two webhooks (one per each created user)
		webhook1 := createWebhookWithActorUID(ctx, t, UID1, webhooksStore)
		webhook2 := createWebhookWithActorUID(ctx, t, UID2, webhooksStore)

		// Check that created_by_user_id is correctly set and updated_by_user_id is
		// defaulted to NULL
		assert.Equal(t, UID1, webhook1.CreatedByUserID)
		assert.Equal(t, int32(0), webhook1.UpdatedByUserID)
		assert.Equal(t, UID2, webhook2.CreatedByUserID)
		assert.Equal(t, int32(0), webhook2.UpdatedByUserID)

		// Updating webhook1 by user2 and checking that updated_by_user_id is updated
		ctx = actor.WithActor(ctx, &actor.Actor{UID: UID2})
		webhook1, err = webhooksStore.Update(ctx, webhook1)
		assert.NoError(t, err)
		assert.Equal(t, UID2, webhook1.UpdatedByUserID)
	})
	t.Run("with bad key", func(t *testing.T) {
		store := db.Webhooks(&et.BadKey{Err: errors.New("some error occurred, sorry")})

		_, err := store.Create(ctx, "name", extsvc.KindGitHub, "https://github.com", 0, types.NewUnencryptedSecret("very secret (not)"))
		assert.Error(t, err)
		assert.Equal(t, "encrypting secret: some error occurred, sorry", err.Error())
	})
}

func TestWebhookDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Webhooks(nil)

	// Test that delete with wrong UUID returns an error
	nonExistentUUID := uuid.New()
	err := store.Delete(ctx, DeleteWebhookOpts{UUID: nonExistentUUID})
	if !errors.HasType[*WebhookNotFoundError](err) {
		t.Fatalf("want WebhookNotFoundError, got: %s", err)
	}
	assert.EqualError(t, err, fmt.Sprintf("failed to delete webhook: webhook with UUID %s not found", nonExistentUUID))

	// Test that delete with wrong ID returns an error
	nonExistentID := int32(123)
	err = store.Delete(ctx, DeleteWebhookOpts{ID: nonExistentID})
	if !errors.HasType[*WebhookNotFoundError](err) {
		t.Fatalf("want WebhookNotFoundError, got: %s", err)
	}
	assert.EqualError(t, err, fmt.Sprintf("failed to delete webhook: webhook with ID %d not found", nonExistentID))

	// Test that delete with empty options returns an error
	err = store.Delete(ctx, DeleteWebhookOpts{})
	assert.EqualError(t, err, "not enough conditions to build query to delete webhook")

	// Creating something to be deleted
	createdWebhook1 := createWebhook(ctx, t, store)
	createdWebhook2 := createWebhook(ctx, t, store)

	// Test that delete with right UUID deletes the webhook
	err = store.Delete(ctx, DeleteWebhookOpts{UUID: createdWebhook1.UUID})
	assert.NoError(t, err)

	// Test that delete with both ID and UUID deletes the webhook by ID
	err = store.Delete(ctx, DeleteWebhookOpts{UUID: uuid.New(), ID: createdWebhook2.ID})
	assert.NoError(t, err)

	exists, _, err := basestore.ScanFirstBool(db.QueryContext(ctx, "SELECT EXISTS(SELECT 1 FROM webhooks WHERE id IN ($1, $2))", createdWebhook1.ID, createdWebhook2.ID))
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestWebhookUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	newCodeHostURN, err := extsvc.NewCodeHostBaseURL("https://new.github.com")
	require.NoError(t, err)
	const updatedSecret = "my new secret"

	t.Run("updating w/ unencrypted secret", func(t *testing.T) {
		store := db.Webhooks(nil)
		created := createWebhook(ctx, t, store)

		created.CodeHostURN = newCodeHostURN
		created.Secret = types.NewUnencryptedSecret(updatedSecret)
		updated, err := store.Update(ctx, created)
		if err != nil {
			t.Fatalf("error updating webhook: %s", err)
		}
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, created.UUID, updated.UUID)
		assert.Equal(t, created.CodeHostKind, updated.CodeHostKind)
		assert.Equal(t, newCodeHostURN.String(), updated.CodeHostURN.String())
		assert.NotZero(t, created.CreatedAt, updated.CreatedAt)
		assert.NotZero(t, created.UpdatedAt)
		assert.Greater(t, updated.UpdatedAt, created.UpdatedAt)
	})

	t.Run("updating w/ encrypted secret", func(t *testing.T) {
		store := db.Webhooks(et.ByteaTestKey{})
		created := createWebhook(ctx, t, store)

		created.CodeHostURN = newCodeHostURN
		created.Secret = types.NewUnencryptedSecret(updatedSecret)
		updated, err := store.Update(ctx, created)
		if err != nil {
			t.Fatalf("error updating webhook: %s", err)
		}
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, created.UUID, updated.UUID)
		assert.Equal(t, created.CodeHostKind, updated.CodeHostKind)
		assert.Equal(t, newCodeHostURN.String(), updated.CodeHostURN.String())
		assert.NotZero(t, created.CreatedAt, updated.CreatedAt)
		assert.NotZero(t, created.UpdatedAt)
		assert.Greater(t, updated.UpdatedAt, created.UpdatedAt)

		row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", created.ID)
		var rawSecret string
		err = row.Scan(&rawSecret)
		assert.NoError(t, err)

		decryptedSecret, err := updated.Secret.Decrypt(ctx)
		assert.NoError(t, err)
		assert.NotEqual(t, rawSecret, decryptedSecret)
		assert.Equal(t, decryptedSecret, updatedSecret)
	})

	t.Run("updating webhook to have nil secret", func(t *testing.T) {
		store := db.Webhooks(nil)
		created := createWebhook(ctx, t, store)
		created.Secret = nil
		updated, err := store.Update(ctx, created)
		if err != nil {
			t.Fatalf("unexpected error updating webhook: %s", err)
		}
		assert.Nil(t, updated.Secret)

		// Also assert that the values in the DB are nil
		row := db.QueryRowContext(ctx, "SELECT secret, encryption_key_id FROM webhooks where id = $1", updated.ID)
		var rawSecret string
		var rawEncryptionKey string
		err = row.Scan(&dbutil.NullString{S: &rawSecret}, &dbutil.NullString{S: &rawEncryptionKey})
		assert.NoError(t, err)
		assert.Empty(t, rawSecret)
		assert.Empty(t, rawEncryptionKey)
	})

	t.Run("updating webhook that doesn't exist", func(t *testing.T) {
		nonExistentUUID := uuid.New()
		webhook := types.Webhook{ID: 100, UUID: nonExistentUUID}

		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(t))

		store := db.Webhooks(nil)
		_, err := store.Update(ctx, &webhook)
		if err == nil {
			t.Fatal("attempting to update a non-existent webhook should return an error")
		}
		assert.Equal(t, err, &WebhookNotFoundError{ID: 100, UUID: nonExistentUUID})
	})
}

func createWebhookWithActorUID(ctx context.Context, t *testing.T, actorUID int32, store WebhookStore) *types.Webhook {
	t.Helper()
	kind := extsvc.KindGitHub
	encryptedSecret := types.NewUnencryptedSecret(testSecret)

	created, err := store.Create(ctx, githubWebhookName, kind, testURN, actorUID, encryptedSecret)
	assert.NoError(t, err)
	return created
}

func TestWebhookCount(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Webhooks(et.ByteaTestKey{})
	ctx := context.Background()

	totalWebhooks, totalGitlabHooks := createTestWebhooks(ctx, t, store)

	t.Run("basic, no opts", func(t *testing.T) {
		count, err := store.Count(ctx, WebhookListOptions{})
		assert.NoError(t, err)
		assert.Equal(t, totalWebhooks, count)
	})

	t.Run("with filtering by kind", func(t *testing.T) {
		count, err := store.Count(ctx, WebhookListOptions{Kind: extsvc.KindGitLab})
		assert.NoError(t, err)
		assert.Equal(t, totalGitlabHooks, count)
	})
}

func TestWebhookList(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Webhooks(et.ByteaTestKey{})
	ctx := context.Background()

	totalWebhooks, numGitlabHooks := createTestWebhooks(ctx, t, store)

	t.Run("basic, no opts", func(t *testing.T) {
		allWebhooks, err := store.List(ctx, WebhookListOptions{})
		assert.NoError(t, err)
		assert.Len(t, allWebhooks, totalWebhooks)
	})

	t.Run("specify code host kind", func(t *testing.T) {
		gitlabWebhooks, err := store.List(ctx, WebhookListOptions{Kind: extsvc.KindGitLab})
		assert.NoError(t, err)
		assert.Len(t, gitlabWebhooks, numGitlabHooks)
	})

	t.Run("with pagination", func(t *testing.T) {
		webhooks, err := store.List(ctx, WebhookListOptions{LimitOffset: &LimitOffset{Limit: 2, Offset: 1}})
		assert.NoError(t, err)
		assert.Len(t, webhooks, 2)
		assert.Equal(t, webhooks[0].ID, int32(2))
		assert.Equal(t, webhooks[0].CodeHostKind, extsvc.KindGitHub)
		assert.Equal(t, webhooks[0].Name, githubWebhookName)
		assert.Equal(t, webhooks[1].ID, int32(3))
		assert.Equal(t, webhooks[1].CodeHostKind, extsvc.KindGitLab)
		assert.Equal(t, webhooks[1].Name, gitlabWebhookName)
	})

	t.Run("with pagination and filtering by code host kind", func(t *testing.T) {
		webhooks, err := store.List(ctx, WebhookListOptions{Kind: extsvc.KindGitHub, LimitOffset: &LimitOffset{Limit: 3, Offset: 2}})
		assert.NoError(t, err)
		assert.Len(t, webhooks, 3)
		for _, wh := range webhooks {
			assert.Equal(t, wh.CodeHostKind, extsvc.KindGitHub)
		}
	})

	t.Run("with cursor", func(t *testing.T) {
		t.Run("with invalid direction", func(t *testing.T) {
			cursor := types.Cursor{
				Column:    "id",
				Direction: "foo",
				Value:     "2",
			}
			_, err := store.List(ctx, WebhookListOptions{Cursor: &cursor})
			assert.Equal(t, err.Error(), `parsing webhook cursor: missing or invalid cursor direction: "foo"`)
		})
		t.Run("with invalid column", func(t *testing.T) {
			cursor := types.Cursor{
				Column:    "uuid",
				Direction: "next",
				Value:     "2",
			}
			_, err := store.List(ctx, WebhookListOptions{Cursor: &cursor})
			assert.Equal(t, err.Error(), `parsing webhook cursor: missing or invalid cursor: "uuid" "2"`)
		})
		t.Run("valid", func(t *testing.T) {
			cursor := types.Cursor{
				Column:    "id",
				Direction: "next",
				Value:     "4",
			}
			webhooks, err := store.List(ctx, WebhookListOptions{Cursor: &cursor})
			assert.NoError(t, err)
			assert.Len(t, webhooks, 7)
			assert.Equal(t, webhooks[0].ID, int32(4))
		})
	})
}

func createTestWebhooks(ctx context.Context, t *testing.T, store WebhookStore) (int, int) {
	t.Helper()
	encryptedSecret := types.NewUnencryptedSecret(testSecret)
	numGitlabHooks := 0
	totalWebhooks := 10
	for i := 1; i <= totalWebhooks; i++ {
		var err error
		if i%3 == 0 {
			numGitlabHooks++
			_, err = store.Create(ctx, gitlabWebhookName, extsvc.KindGitLab, fmt.Sprintf("http://instance-%d.github.com", i), 0, encryptedSecret)
		} else {
			_, err = store.Create(ctx, githubWebhookName, extsvc.KindGitHub, fmt.Sprintf("http://instance-%d.gitlab.com", i), 0, encryptedSecret)
		}
		assert.NoError(t, err)
	}
	return totalWebhooks, numGitlabHooks
}

func createWebhook(ctx context.Context, t *testing.T, store WebhookStore) *types.Webhook {
	return createWebhookWithActorUID(ctx, t, 0, store)
}

func TestGetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Webhooks(nil)

	// Test that non-existent webhook cannot be found
	webhook, err := store.GetByID(ctx, 1)
	assert.Error(t, err)
	assert.EqualError(t, err, "webhook with ID 1 not found")
	assert.Nil(t, webhook)

	// Test that existent webhook cannot be found
	createdWebhook, err := store.Create(ctx, githubWebhookName, extsvc.KindGitHub, "https://github.com", 0, types.NewUnencryptedSecret("very secret (not)"))
	assert.NoError(t, err)

	webhook, err = store.GetByID(ctx, createdWebhook.ID)
	assert.NoError(t, err)
	assert.NotNil(t, webhook)
	assert.Equal(t, webhook.ID, createdWebhook.ID)
	assert.Equal(t, webhook.UUID, createdWebhook.UUID)
	assert.Equal(t, webhook.Secret, createdWebhook.Secret)
	assert.Equal(t, webhook.Name, createdWebhook.Name)
	assert.Equal(t, webhook.CodeHostKind, createdWebhook.CodeHostKind)
	assert.Equal(t, webhook.CodeHostURN.String(), createdWebhook.CodeHostURN.String())
	assert.Equal(t, webhook.CreatedAt, createdWebhook.CreatedAt)
	assert.Equal(t, webhook.UpdatedAt, createdWebhook.UpdatedAt)
}

func TestGetByUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Webhooks(nil)

	// Test that non-existent webhook cannot be found
	randomUUID := uuid.New()
	webhook, err := store.GetByUUID(ctx, randomUUID)
	assert.EqualError(t, err, fmt.Sprintf("webhook with UUID %s not found", randomUUID))
	assert.Nil(t, webhook)

	// Test that existent webhook cannot be found
	createdWebhook, err := store.Create(ctx, githubWebhookName, extsvc.KindGitHub, "https://github.com", 0, types.NewUnencryptedSecret("very secret (not)"))
	assert.NoError(t, err)

	webhook, err = store.GetByUUID(ctx, createdWebhook.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, webhook)
	assert.Equal(t, webhook.ID, createdWebhook.ID)
	assert.Equal(t, webhook.UUID, createdWebhook.UUID)
	assert.Equal(t, webhook.Secret, createdWebhook.Secret)
	assert.Equal(t, webhook.Name, createdWebhook.Name)
	assert.Equal(t, webhook.CodeHostKind, createdWebhook.CodeHostKind)
	assert.Equal(t, webhook.CodeHostURN.String(), createdWebhook.CodeHostURN.String())
	assert.Equal(t, webhook.CreatedAt, createdWebhook.CreatedAt)
	assert.Equal(t, webhook.UpdatedAt, createdWebhook.UpdatedAt)
}
