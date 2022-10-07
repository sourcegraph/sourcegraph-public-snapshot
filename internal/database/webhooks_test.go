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
