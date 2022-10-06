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

	Create(ctx context.Context, newWebhook *types.Webhook) (*types.Webhook, error)
	Delete(ctx context.Context, id string) error
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
func (s *webhookStore) Create(ctx context.Context, hook *types.Webhook) (*types.Webhook, error) {
	// TODO: Test what happens we secret is empty as some code hosts don't require it
	encryptedSecret, keyID, err := hook.Secret.Encrypt(ctx, s.key)
	if err != nil {
		return nil, errors.Wrap(err, "encrypting secret")
	}

	q := sqlf.Sprintf(webhookCreateQueryFmtstr,
		hook.CodeHostKind,
		hook.CodeHostURN,
		encryptedSecret,
		keyID,
		sqlf.Join(webhookColumns, ", "),
	)

	created := new(types.Webhook)
	row := s.QueryRow(ctx, q)

	if err := s.scanWebhook(created, row); err != nil {
		return nil, errors.Wrap(err, "scanning webhook")
	}

	return created, nil
}

func (s *webhookStore) scanWebhook(hook *types.Webhook, sc dbutil.Scanner) error {
	var keyID string
	var rawSecret string

	if err := sc.Scan(
		&hook.ID,
		&hook.CodeHostKind,
		&hook.CodeHostURN,
		&rawSecret,
		&hook.CreatedAt,
		&hook.UpdatedAt,
		&keyID,
	); err != nil {
		return err
	}

	hook.Secret = types.NewEncryptedSecret(rawSecret, keyID, s.key)

	return nil
}

var webhookColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("code_host_kind"),
	sqlf.Sprintf("code_host_urn"),
	sqlf.Sprintf("secret"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("encryption_key_id"),
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

// Delete the webhook
func (s *webhookStore) Delete(ctx context.Context, id string) error {
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
