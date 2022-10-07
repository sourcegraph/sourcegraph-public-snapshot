package database

import (
	"context"
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

	tx, err := db.Transact(ctx)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = tx.Done(errors.New("rollback")) })

	store := tx.Webhooks(nil)

	hook := &types.Webhook{
		ID:           "",
		CodeHostKind: extsvc.KindGitHub,
		CodeHostURN:  "https://github.com",
		Secret:       types.NewUnencryptedSecret("very secret (not)"),
	}
	created, err := store.Create(ctx, hook)
	assert.NoError(t, err)

	// Check that the calculated fields were correctly calculated.
	id := created.ID
	assert.NotZero(t, id)
	_ = uuid.MustParse(id)
	assert.Equal(t, hook.CodeHostKind, created.CodeHostKind)
	assert.Equal(t, hook.CodeHostURN, created.CodeHostURN)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)

	// getting the secret from the DB as is to verify that it is not encrypted
	row := tx.QueryRowContext(ctx, "SELECT secret FROM webhooks")
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

	tx, err := db.Transact(ctx)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = tx.Done(errors.New("rollback")) })

	store := tx.Webhooks(et.ByteaTestKey{})

	const secret = "don't tell anyone"
	hook := &types.Webhook{
		ID:           "",
		CodeHostKind: extsvc.KindGitHub,
		CodeHostURN:  "https://github.com",
		Secret:       types.NewUnencryptedSecret(secret),
	}
	created, err := store.Create(ctx, hook)
	assert.NoError(t, err)

	// Check that the calculated fields were correctly calculated.
	id := created.ID
	assert.NotZero(t, id)
	_ = uuid.MustParse(id)
	assert.Equal(t, hook.CodeHostKind, created.CodeHostKind)
	assert.Equal(t, hook.CodeHostURN, created.CodeHostURN)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)

	// getting the secret from the DB as is to verify that it is encrypted
	row := tx.QueryRowContext(ctx, "SELECT secret FROM webhooks")
	var rawSecret string
	err = row.Scan(&rawSecret)
	assert.NoError(t, err)

	decryptedSecret, err := created.Secret.Decrypt(ctx)
	assert.NoError(t, err)
	assert.NotEqual(t, rawSecret, decryptedSecret)
	assert.Equal(t, secret, decryptedSecret)
}

func TestWebhookCreateWithBadKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	tx, err := db.Transact(ctx)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = tx.Done(errors.New("rollback")) })

	store := tx.Webhooks(&et.BadKey{Err: errors.New("some error occurred, sorry")})

	hook := &types.Webhook{
		ID:           "",
		CodeHostKind: extsvc.KindGitHub,
		CodeHostURN:  "https://github.com",
		Secret:       types.NewUnencryptedSecret("very secret (not)"),
	}
	_, err = store.Create(ctx, hook)
	assert.Error(t, err)
	assert.Equal(t, "encrypting secret: some error occurred, sorry", err.Error())
}
