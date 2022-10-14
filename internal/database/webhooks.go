package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
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

	Create(ctx context.Context, kind, urn string, actorUID int32, secret *types.EncryptableSecret) (*types.Webhook, error)
	GetByID(ctx context.Context, id int32) (*types.Webhook, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (*types.Webhook, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, actorUID int32, newWebhook *types.Webhook) (*types.Webhook, error)
	List(ctx context.Context, opts WebhookListOptions) ([]*types.Webhook, error)
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
func (s *webhookStore) Create(ctx context.Context, kind, urn string, actorUID int32, secret *types.EncryptableSecret) (*types.Webhook, error) {
	var (
		err             error
		encryptedSecret string
		keyID           string
	)

	if secret != nil {
		encryptedSecret, keyID, err = secret.Encrypt(ctx, s.key)
		if err != nil {
			return nil, errors.Wrap(err, "encrypting secret")
		}
		if encryptedSecret == "" && keyID == "" {
			return nil, errors.New("empty secret and key provided")
		}
	}

	q := sqlf.Sprintf(webhookCreateQueryFmtstr,
		kind,
		urn,
		encryptedSecret,
		keyID,
		dbutil.NullInt32Column(actorUID),
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
INSERT INTO
	webhooks (
		code_host_kind,
		code_host_urn,
		secret,
		encryption_key_id,
		created_by_user_id
	)
	VALUES (
		%s,
		%s,
		%s,
		%s,
		%s
	)
	RETURNING %s
`

var webhookColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("uuid"),
	sqlf.Sprintf("code_host_kind"),
	sqlf.Sprintf("code_host_urn"),
	sqlf.Sprintf("secret"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("created_by_user_id"),
	sqlf.Sprintf("updated_by_user_id"),
}

const webhookGetByIDFmtstr = `
SELECT %s FROM webhooks
WHERE id = %d
`

func (s *webhookStore) GetByID(ctx context.Context, id int32) (*types.Webhook, error) {
	q := sqlf.Sprintf(webhookGetByIDFmtstr,
		sqlf.Join(webhookColumns, ", "),
		id,
	)

	webhook, err := scanWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &WebhookNotFoundError{ID: id}
		}
		return nil, errors.Wrap(err, "scanning webhook")
	}

	return webhook, nil
}

const webhookGetByUUIDFmtstr = `
SELECT %s FROM webhooks
WHERE uuid = %s
`

func (s *webhookStore) GetByUUID(ctx context.Context, id uuid.UUID) (*types.Webhook, error) {
	q := sqlf.Sprintf(webhookGetByUUIDFmtstr,
		sqlf.Join(webhookColumns, ", "),
		id,
	)

	webhook, err := scanWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &WebhookNotFoundError{UUID: id}
		}
		return nil, errors.Wrap(err, "scanning webhook")
	}

	return webhook, nil
}

const webhookDeleteQueryFmtstr = `
DELETE FROM webhooks
WHERE uuid = %s
`

// Delete the webhook. Error is returned when provided UUID is invalid, the
// webhook is not found or something went wrong during an SQL query.
func (s *webhookStore) Delete(ctx context.Context, id uuid.UUID) error {
	q := sqlf.Sprintf(webhookDeleteQueryFmtstr, id)
	result, err := s.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "running delete SQL query")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking rows affected after deletion")
	}

	if rowsAffected == 0 {
		return errors.Wrap(&WebhookNotFoundError{UUID: id}, "failed to delete webhook")
	}
	return nil
}

// WebhookNotFoundError occurs when a webhook is not found.
type WebhookNotFoundError struct {
	ID   int32
	UUID uuid.UUID
}

func (w *WebhookNotFoundError) Error() string {
	if w.UUID != uuid.Nil {
		return fmt.Sprintf("webhook with UUID %s not found", w.UUID)
	} else {
		return fmt.Sprintf("webhook with ID %d not found", w.ID)
	}
}

func (w *WebhookNotFoundError) NotFound() bool {
	return true
}

// Update the webhook
func (s *webhookStore) Update(ctx context.Context, actorUID int32, newWebhook *types.Webhook) (*types.Webhook, error) {
	var (
		err             error
		encryptedSecret string
		keyID           string
	)
	if newWebhook.Secret != nil {
		encryptedSecret, keyID, err = newWebhook.Secret.Encrypt(ctx, s.key)
		if err != nil || (encryptedSecret == "" && keyID == "") {
			return nil, errors.Wrap(err, "encrypting secret")
		}
		if encryptedSecret == "" && keyID == "" {
			return nil, errors.New("empty secret and key provided")
		}
	}

	q := sqlf.Sprintf(webhookUpdateQueryFmtstr,
		newWebhook.CodeHostURN, encryptedSecret, keyID, dbutil.NullInt32Column(actorUID), newWebhook.ID,
		sqlf.Join(webhookColumns, ", "))

	updated, err := scanWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &WebhookNotFoundError{ID: newWebhook.ID, UUID: newWebhook.UUID}
		}
		return nil, errors.Wrap(err, "scanning webhook")
	}

	return updated, nil
}

const webhookUpdateQueryFmtstr = `
UPDATE webhooks
SET
	code_host_urn = %s,
	secret = %s,
	encryption_key_id = %s,
	updated_at = NOW(),
	updated_by_user_id = %s
WHERE
	id = %s
RETURNING
	%s
`

// List the webhooks
func (s *webhookStore) List(ctx context.Context, opt WebhookListOptions) ([]*types.Webhook, error) {
	q := sqlf.Sprintf(webhookListQueryFmtstr, sqlf.Join(webhookColumns, ", "))
	if opt.Kind != "" {
		q = sqlf.Sprintf("%s\nWHERE code_host_kind = %s", q, opt.Kind)
	}
	if opt.LimitOffset != nil {
		q = sqlf.Sprintf("%s\n%s", q, opt.LimitOffset.SQL())
	}
	rows, err := s.Query(ctx, q)
	if err != nil {
		return []*types.Webhook{}, errors.Wrap(err, "error running query")
	}
	defer rows.Close()
	res := make([]*types.Webhook, 0, 20)
	for rows.Next() {
		webhook, err := scanWebhook(rows, s.key)
		if err != nil {
			return nil, err
		}
		res = append(res, webhook)
	}
	return res, nil
}

type WebhookListOptions struct {
	Kind string
	*LimitOffset
}

const webhookListQueryFmtstr = `
SELECT
	%s
FROM webhooks
`

func scanWebhook(sc dbutil.Scanner, key encryption.Key) (*types.Webhook, error) {
	var (
		hook      types.Webhook
		keyID     string
		rawSecret string
	)

	if err := sc.Scan(
		&hook.ID,
		&hook.UUID,
		&hook.CodeHostKind,
		&hook.CodeHostURN,
		&rawSecret,
		&hook.CreatedAt,
		&hook.UpdatedAt,
		&keyID,
		&dbutil.NullInt32{N: &hook.CreatedByUserID},
		&dbutil.NullInt32{N: &hook.UpdatedByUserID},
	); err != nil {
		return nil, err
	}

	hook.Secret = types.NewEncryptedSecret(rawSecret, keyID, key)

	return &hook, nil
}
