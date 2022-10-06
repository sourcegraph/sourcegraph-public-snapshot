package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestWebhookCreateUnencrypted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.Webhooks(nil)

	kind := extsvc.KindGitHub
	urn := "https://github.com"
	secret := types.NewUnencryptedSecret("very secret (not)")

	created, err := store.Create(ctx, kind, urn, secret)
	assert.NoError(t, err)

	// Check that the calculated fields were correctly calculated.
	assert.NotZero(t, created.ID)
	_, err = uuid.Parse(created.RandomID)
	assert.NoError(t, err)
	assert.Equal(t, kind, created.CodeHostKind)
	assert.Equal(t, urn, created.CodeHostURN)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)

	// getting the secret from the DB as is to verify that it is not encrypted
	row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", created.ID)
	var rawSecret string
	err = row.Scan(&rawSecret)
	assert.NoError(t, err)

	decryptedSecret, err := created.Secret.Decrypt(ctx)
	assert.NoError(t, err)
	assert.Equal(t, rawSecret, decryptedSecret)
}

func TestWebhookCreateEncrypted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	store := db.Webhooks(et.ByteaTestKey{})

	const secret = "don't tell anyone"
	kind := extsvc.KindGitHub
	urn := "https://github.com"
	encryptedSecret := types.NewUnencryptedSecret(secret)

	created, err := store.Create(ctx, kind, urn, encryptedSecret)
	assert.NoError(t, err)

	// Check that the calculated fields were correctly calculated.
	assert.NotZero(t, created.ID)
	_, err = uuid.Parse(created.RandomID)
	assert.NoError(t, err)
	assert.Equal(t, kind, created.CodeHostKind)
	assert.Equal(t, urn, created.CodeHostURN)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)

	// getting the secret from the DB as is to verify that it is encrypted
	row := db.QueryRowContext(ctx, "SELECT secret FROM webhooks where id = $1", created.ID)
	var rawSecret string
	err = row.Scan(&rawSecret)
	assert.NoError(t, err)

	decryptedSecret, err := created.Secret.Decrypt(ctx)
	assert.NoError(t, err)
	assert.NotEqual(t, rawSecret, decryptedSecret)
	assert.Equal(t, secret, decryptedSecret)
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
	_, err = uuid.Parse(created.RandomID)
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

	tests := []struct {
		name              string
		existingWebhookID string
		deleteID          string
		expectedErrorMsg  string
	}{
		{
			name:              "Invalid UUID provided",
			existingWebhookID: "",
			deleteID:          "me-me ID did you get it? me, not you",
			expectedErrorMsg:  "invalid UUID provided: invalid UUID format",
		},
		{
			name:              "Webhook with given ID doesn't exist",
			existingWebhookID: "cafebabe-1337-0420-face-00deadbeef00",
			deleteID:          "cafebabe-1337-0420-face-00baaaaaad00",
			expectedErrorMsg:  "webhook not found: Cannot delete a webhook with id=cafebabe-1337-0420-face-00baaaaaad00: not found.",
		},
		{
			name:              "Webhook with given ID exists and successfully deleted",
			existingWebhookID: "cafebabe-1338-0420-face-00deadbeef00",
			deleteID:          "cafebabe-1338-0420-face-00deadbeef00",
			expectedErrorMsg:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// init everything
			tx, err := db.Transact(ctx)
			assert.NoError(t, err)
			t.Cleanup(func() {
				_ = tx.Done(errors.New("rollback"))
			})

			store := tx.Webhooks(nil)

			// creating a webhook if needed
			if test.existingWebhookID != "" {
				tx.QueryRowContext(ctx, fmt.Sprintf(`
INSERT INTO webhooks (rand_id, code_host_kind, code_host_urn)
VALUES ('%s', 'code_host_kind', 'code_host_urn')`,
					test.existingWebhookID))
			}

			// deleting a webhook
			err = store.Delete(ctx, test.deleteID)

			if test.expectedErrorMsg != "" {
				// checking that the error is expected
				assert.EqualError(t, err, test.expectedErrorMsg)
			} else {
				// double-check that a webhook is deleted
				row := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM webhooks WHERE rand_id='%s')", test.deleteID))
				var exists bool
				err = row.Scan(&exists)
				assert.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}
