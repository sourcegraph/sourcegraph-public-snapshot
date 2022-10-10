package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var testSecret = "my secret"

func TestWebhookCreateUnencrypted(t *testing.T) {
	t.Parallel()

	runCreateWebhookTest(context.Background(), t, false)
}

func TestWebhookCreateEncrypted(t *testing.T) {
	t.Parallel()

	runCreateWebhookTest(context.Background(), t, true)
}

func TestWebhookCreateNoSecret(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.Webhooks(et.ByteaTestKey{})

	kind := extsvc.KindGitHub
	urn := "https://github.com"

	created, err := store.Create(ctx, kind, urn, nil)
	assert.NoError(t, err)

	// Check that the calculated fields were correctly calculated.
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.UUID)
	assert.NoError(t, err)
	assert.Equal(t, kind, created.CodeHostKind)
	assert.Equal(t, urn, created.CodeHostURN)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)

	// secret in the DB should be null
	row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", created.ID)
	var rawSecret string
	err = row.Scan(&rawSecret)
	assert.NoError(t, err)
	assert.Zero(t, rawSecret)
}

func TestWebhookCreateWithBadKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.Webhooks(&et.BadKey{Err: errors.New("some error occurred, sorry")})

	_, err := store.Create(ctx, extsvc.KindGitHub, "https://github.com", types.NewUnencryptedSecret("very secret (not)"))
	assert.Error(t, err)
	assert.Equal(t, "encrypting secret: some error occurred, sorry", err.Error())
}

func TestWebhookDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Webhooks(nil)

	// Test that delete with wrong ID returns an error
	nonExistentUUID := uuid.New()
	err := store.Delete(ctx, nonExistentUUID)
	if !errors.HasType(err, &WebhookNotFoundError{}) {
		t.Fatalf("want WebhookNotFoundError, got: %s", err)
	}
	assert.EqualError(t, err, fmt.Sprintf("failed to delete webhook: webhook with UUID %s not found", nonExistentUUID))

	// Test that delete with right ID deletes the webhook
	createdWebhook, err := store.Create(ctx, extsvc.KindGitHub, "https://github.com", types.NewUnencryptedSecret("very secret (not)"))
	assert.NoError(t, err)
	err = store.Delete(ctx, createdWebhook.UUID)
	assert.NoError(t, err)

	exists, _, err := basestore.ScanFirstBool(db.QueryContext(ctx, "SELECT EXISTS(SELECT 1 FROM webhooks WHERE uuid=$1)", createdWebhook.UUID))
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestWebhookUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	const newCodeHostURN = "https://new.github.com"
	const updatedSecret = "my new secret"

	t.Run("updating w/ unencrypted secret", func(t *testing.T) {
		store, _, created := runCreateWebhookTest(ctx, t, false)

		created.CodeHostURN = newCodeHostURN
		created.Secret = types.NewUnencryptedSecret(updatedSecret)
		updated, err := store.Update(ctx, created)
		if err != nil {
			t.Fatalf("error updating webhook: %s", err)
		}
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, created.UUID, updated.UUID)
		assert.Equal(t, created.CodeHostKind, updated.CodeHostKind)
		assert.Equal(t, newCodeHostURN, updated.CodeHostURN)
		assert.NotZero(t, created.CreatedAt, updated.CreatedAt)
		assert.NotZero(t, created.UpdatedAt)
		assert.Greater(t, updated.UpdatedAt, created.UpdatedAt)
	})

	t.Run("updating w/ encrypted secret", func(t *testing.T) {
		store, db, created := runCreateWebhookTest(ctx, t, true)
		created.CodeHostURN = newCodeHostURN
		created.Secret = types.NewUnencryptedSecret(updatedSecret)
		updated, err := store.Update(ctx, created)
		if err != nil {
			t.Fatalf("error updating webhook: %s", err)
		}
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, created.UUID, updated.UUID)
		assert.Equal(t, created.CodeHostKind, updated.CodeHostKind)
		assert.Equal(t, newCodeHostURN, updated.CodeHostURN)
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

	t.Run("updating webhook that doesn't exist", func(t *testing.T) {
		webhook := types.Webhook{ID: 100}

		logger := logtest.Scoped(t)
		db := NewDB(logger, dbtest.NewDB(logger, t))

		store := db.Webhooks(nil)
		_, err := store.Update(context.Background(), &webhook)
		if err == nil {
			t.Fatal("attempting to update a non-existent webhook should return an error")
		}
		assert.Equal(t, err, webhookNotFoundErr{id: 100})
	})
}

func runCreateWebhookTest(ctx context.Context, t *testing.T, encrypt bool) (WebhookStore, DB, *types.Webhook) {
	t.Helper()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.Webhooks(nil)
	if encrypt {
		store = db.Webhooks(et.ByteaTestKey{})
	}

	kind := extsvc.KindGitHub
	urn := "https://github.com"
	encryptedSecret := types.NewUnencryptedSecret(testSecret)

	created, err := store.Create(ctx, kind, urn, encryptedSecret)
	assert.NoError(t, err)

	// Check that the calculated fields were correctly calculated.
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.UUID)
	assert.Equal(t, kind, created.CodeHostKind)
	assert.Equal(t, urn, created.CodeHostURN)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)

	// getting the secret from the DB as is to verify its encryption
	row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", created.ID)
	var rawSecret string
	err = row.Scan(&rawSecret)
	assert.NoError(t, err)

	decryptedSecret, err := created.Secret.Decrypt(ctx)
	assert.NoError(t, err)

	if !encrypt {
		// if no encryption, raw secret stored in the db and decrypted secret should be the same
		assert.Equal(t, rawSecret, decryptedSecret)
	} else {
		// if encryption is specified, decrypted secret and raw secret should not match
		assert.NotEqual(t, rawSecret, decryptedSecret)
		assert.Equal(t, testSecret, decryptedSecret)
	}
	return store, db, created
}
