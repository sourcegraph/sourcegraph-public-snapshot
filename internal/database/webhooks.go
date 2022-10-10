package database

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type WebhookStore interface {
	basestore.ShareableStore

	Create(ctx context.Context, kind, urn string, secret *types.EncryptableSecret) (*types.Webhook, error)
	GetByID(ctx context.Context, id int32) (*types.Webhook, error)
	GetByRandomID(ctx context.Context, id string) (*types.Webhook, error)
	Delete(ctx context.Context, id int32) error
	Update(ctx context.Context, newWebhook *types.Webhook) (*types.Webhook, error)
	List(ctx context.Context) ([]*types.Webhook, error)
}

type webhookStore struct {
	*basestore.Store

	logger log.Logger
	key    encryption.Key
}

var _ WebhookStore = &webhookStore{}

func WebhooksWith(other basestore.ShareableStore, key encryption.Key) WebhookStore {
	return &webhookStore{
		Store: basestore.NewWithHandle(other.Handle()),
		key:   key,
	}
}

// Create the webhook
//
// secret is optional since some code hosts do not support signing payloads.
// Also, encryption at the instance level is also optional. If encryption is
// disabled then the secret value will be stored in plain text in the secret
// column and encryption_key_id will be blank.
//
// If encryption IS enabled then the encrypted value will be stored in secret and
// the encryption_key_id field will also be populated so that we can decrypt the
// value later.
func (s *webhookStore) Create(ctx context.Context, kind, urn string, secret *types.EncryptableSecret) (*types.Webhook, error) {
	var (
		err             error
		encryptedSecret string
		keyID           string
	)

	if secret != nil {
		encryptedSecret, keyID, err = secret.Encrypt(ctx, s.key)
		if err != nil || (encryptedSecret == "" && keyID == "") {
			return nil, errors.Wrap(err, "encrypting secret")
		}
	}

	q := sqlf.Sprintf(webhookCreateQueryFmtstr,
		kind,
		urn,
		encryptedSecret,
		keyID,
		// Returning
		sqlf.Join(webhookColumns, ", "),
	)

	created, err := scanWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		return nil, errors.Wrap(err, "scanning webhook")
	}

	return created, nil
}

const webhookCreateQueryFmtstr = `
-- source: internal/database/webhooks.go:Create
INSERT INTO
	webhooks (
		code_host_kind,
		code_host_urn,
		secret,
		encryption_key_id
	)
	VALUES (
		%s,
		%s,
		%s,
		%s
	)
	RETURNING %s
`

var webhookColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("rand_id"),
	sqlf.Sprintf("code_host_kind"),
	sqlf.Sprintf("code_host_urn"),
	sqlf.Sprintf("secret"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("encryption_key_id"),
}

func (s *webhookStore) GetByID(ctx context.Context, id int32) (*types.Webhook, error) {
	panic("Implement this method")
}

func (s *webhookStore) GetByRandomID(ctx context.Context, id string) (*types.Webhook, error) {
	panic("Implement this method")
}

// Delete the webhook
func (s *webhookStore) Delete(ctx context.Context, id int32) error {
	// TODO(sashaostrikov) implement this method
	panic("implement this method")
}

// Update the webhook
func (s *webhookStore) Update(ctx context.Context, newWebhook *types.Webhook) (*types.Webhook, error) {
	// TODO(sashaostrikov) implement this method
	panic("implement this method")
}

// List the webhooks
func (s *webhookStore) List(ctx context.Context) ([]*types.Webhook, error) {
	// TODO(sashaostrikov) implement this method
	panic("implement this method")
}

func scanWebhook(sc dbutil.Scanner, key encryption.Key) (*types.Webhook, error) {
	var hook types.Webhook
	var keyID string
	var rawSecret string

	if err := sc.Scan(
		&hook.ID,
		&hook.RandomID,
		&hook.CodeHostKind,
		&hook.CodeHostURN,
		&rawSecret,
		&hook.CreatedAt,
		&hook.UpdatedAt,
		&keyID,
	); err != nil {
		return nil, err
	}

	hook.Secret = types.NewEncryptedSecret(rawSecret, keyID, key)

	return &hook, nil
}
