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
	Delete(ctx context.Context, opts DeleteWebhookOpts) error
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

type WebhookOpts struct {
	ID   int32
	UUID uuid.UUID
}

type DeleteWebhookOpts WebhookOpts
type GetWebhookOpts WebhookOpts

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
		dbutil.NullStringColumn(encryptedSecret),
		dbutil.NullStringColumn(keyID),
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

const webhookGetFmtstr = `
SELECT %s FROM webhooks
WHERE %s
`

func (s *webhookStore) GetByID(ctx context.Context, id int32) (*types.Webhook, error) {
	return s.getBy(ctx, GetWebhookOpts{ID: id})
}

func (s *webhookStore) GetByUUID(ctx context.Context, id uuid.UUID) (*types.Webhook, error) {
	return s.getBy(ctx, GetWebhookOpts{UUID: id})
}

func (s *webhookStore) getBy(ctx context.Context, opts GetWebhookOpts) (*types.Webhook, error) {
	var whereClause *sqlf.Query
	if opts.ID > 0 {
		whereClause = sqlf.Sprintf("ID = %d", opts.ID)
	}

	if opts.UUID != uuid.Nil {
		whereClause = sqlf.Sprintf("UUID = %s", opts.UUID)
	}

	if whereClause == nil {
		return nil, errors.New("not enough conditions to build query to delete webhook")
	}

	q := sqlf.Sprintf(webhookGetFmtstr,
		sqlf.Join(webhookColumns, ", "),
		whereClause,
	)

	webhook, err := scanWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &WebhookNotFoundError{UUID: opts.UUID, ID: opts.ID}
		}
		return nil, errors.Wrap(err, "scanning webhook")
	}

	return webhook, nil
}

const webhookDeleteByQueryFmtstr = `
DELETE FROM webhooks
WHERE %s
`

// Delete the webhook with given options.
//
// Either ID or UUID can be provided.
//
// No error is returned if both ID and UUID are provided, ID is used in this
// case. Error is returned when the webhook is not found or something went wrong
// during an SQL query.
func (s *webhookStore) Delete(ctx context.Context, opts DeleteWebhookOpts) error {
	query, err := buildDeleteWebhookQuery(opts)
	if err != nil {
		return err
	}

	result, err := s.ExecResult(ctx, query)
	if err != nil {
		return errors.Wrap(err, "running delete SQL query")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking rows affected after deletion")
	}
	if rowsAffected == 0 {
		return errors.Wrap(NewWebhookNotFoundErrorFromOpts(opts), "failed to delete webhook")
	}
	return nil
}

func buildDeleteWebhookQuery(opts DeleteWebhookOpts) (*sqlf.Query, error) {
	if opts.ID > 0 {
		return sqlf.Sprintf(webhookDeleteByQueryFmtstr, sqlf.Sprintf("ID = %d", opts.ID)), nil
	}

	if opts.UUID != uuid.Nil {
		return sqlf.Sprintf(webhookDeleteByQueryFmtstr, sqlf.Sprintf("UUID = %s", opts.UUID)), nil
	}

	return nil, errors.New("not enough conditions to build query to delete webhook")
}

// WebhookNotFoundError occurs when a webhook is not found.
type WebhookNotFoundError struct {
	ID   int32
	UUID uuid.UUID
}

func (w *WebhookNotFoundError) Error() string {
	if w.ID > 0 {
		return fmt.Sprintf("webhook with ID %d not found", w.ID)
	} else {
		return fmt.Sprintf("webhook with UUID %s not found", w.UUID)
	}
}

func (w *WebhookNotFoundError) NotFound() bool {
	return true
}

func NewWebhookNotFoundErrorFromOpts(opts DeleteWebhookOpts) *WebhookNotFoundError {
	return &WebhookNotFoundError{
		ID:   opts.ID,
		UUID: opts.UUID,
	}
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
		if err != nil {
			return nil, errors.Wrap(err, "encrypting secret")
		}
		if encryptedSecret == "" && keyID == "" {
			return nil, errors.New("empty secret and key provided")
		}
	}

	q := sqlf.Sprintf(webhookUpdateQueryFmtstr,
		newWebhook.CodeHostURN,
		dbutil.NullStringColumn(encryptedSecret),
		dbutil.NullStringColumn(keyID),
		dbutil.NullInt32Column(actorUID),
		newWebhook.ID,
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
	predicates := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt.Kind != "" {
		predicates = append(predicates, sqlf.Sprintf("code_host_kind = %s", opt.Kind))
	}

	q := sqlf.Sprintf(
		webhookListQueryFmtstr,
		sqlf.Join(webhookColumns, ","),
		sqlf.Join(predicates, "AND"),
		opt.LimitOffset.SQL())
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
WHERE %s -- Predicates
%s -- Limit/offset
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
		&dbutil.NullString{S: &rawSecret},
		&hook.CreatedAt,
		&hook.UpdatedAt,
		&dbutil.NullString{S: &keyID},
		&dbutil.NullInt32{N: &hook.CreatedByUserID},
		&dbutil.NullInt32{N: &hook.UpdatedByUserID},
	); err != nil {
		return nil, err
	}

	if keyID == "" && rawSecret != "" {
		// We have an unencrypted secret
		hook.Secret = types.NewUnencryptedSecret(rawSecret)
	} else if keyID != "" && rawSecret != "" {
		// We have an encrypted secret
		hook.Secret = types.NewEncryptedSecret(rawSecret, keyID, key)
	}
	// If both keyID and rawSecret are empty then we didn't set a secret and we leave
	// hook.Secret as nil

	return &hook, nil
}
