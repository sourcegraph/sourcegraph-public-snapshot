package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type WebhookStore interface {
	basestore.ShareableStore

	Create(ctx context.Context, name, kind, urn string, actorUID int32, secret *types.EncryptableSecret) (*types.Webhook, error)
	GetByID(ctx context.Context, id int32) (*types.Webhook, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (*types.Webhook, error)
	Delete(ctx context.Context, opts DeleteWebhookOpts) error
	Update(ctx context.Context, webhook *types.Webhook) (*types.Webhook, error)
	List(ctx context.Context, opts WebhookListOptions) ([]*types.Webhook, error)
	Count(ctx context.Context, opts WebhookListOptions) (int, error)
}

type webhookStore struct {
	*basestore.Store

	key encryption.Key
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

type (
	DeleteWebhookOpts WebhookOpts
	GetWebhookOpts    WebhookOpts
)

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
func (s *webhookStore) Create(ctx context.Context, name, kind, urn string, actorUID int32, secret *types.EncryptableSecret) (*types.Webhook, error) {
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
		name,
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
        name,
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
	sqlf.Sprintf("name"),
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
func (s *webhookStore) Update(ctx context.Context, webhook *types.Webhook) (*types.Webhook, error) {
	var (
		err             error
		encryptedSecret string
		keyID           string
	)

	if webhook.Secret != nil {
		encryptedSecret, keyID, err = webhook.Secret.Encrypt(ctx, s.key)
		if err != nil {
			return nil, errors.Wrap(err, "encrypting secret")
		}
		if encryptedSecret == "" && keyID == "" {
			return nil, errors.New("empty secret and key provided")
		}
	}

	q := sqlf.Sprintf(webhookUpdateQueryFmtstr,
		webhook.Name, webhook.CodeHostURN.String(), webhook.CodeHostKind, encryptedSecret, keyID, dbutil.NullInt32Column(actor.FromContext(ctx).UID), webhook.ID,
		sqlf.Join(webhookColumns, ", "))

	updated, err := scanWebhook(s.QueryRow(ctx, q), s.key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &WebhookNotFoundError{ID: webhook.ID, UUID: webhook.UUID}
		}
		return nil, errors.Wrap(err, "scanning webhook")
	}

	return updated, nil
}

const webhookUpdateQueryFmtstr = `
UPDATE webhooks
SET
    name = %s,
	code_host_urn = %s,
    code_host_kind = %s,
	secret = %s,
	encryption_key_id = %s,
	updated_at = NOW(),
	updated_by_user_id = %s
WHERE
	id = %s
RETURNING
	%s
`

func (s *webhookStore) list(ctx context.Context, opt WebhookListOptions, selects *sqlf.Query, scanWebhook func(rows *sql.Rows) error) error {
	q := sqlf.Sprintf(webhookListQueryFmtstr, selects)
	wheres := make([]*sqlf.Query, 0, 2)
	if opt.Kind != "" {
		wheres = append(wheres, sqlf.Sprintf("code_host_kind = %s", opt.Kind))
	}
	cond, err := parseWebhookCursorCond(opt.Cursor)
	if err != nil {
		return errors.Wrap(err, "parsing webhook cursor")
	}
	if cond != nil {
		wheres = append(wheres, cond)
	}
	if len(wheres) != 0 {
		where := sqlf.Join(wheres, "AND")
		q = sqlf.Sprintf("%s\nWHERE %s", q, where)
	}
	if opt.LimitOffset != nil {
		q = sqlf.Sprintf("%s\n%s", q, opt.LimitOffset.SQL())
	}
	rows, err := s.Query(ctx, q)
	if err != nil {
		return errors.Wrap(err, "error running query")
	}
	defer rows.Close()
	for rows.Next() {
		if err := scanWebhook(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

// List the webhooks
func (s *webhookStore) List(ctx context.Context, opt WebhookListOptions) ([]*types.Webhook, error) {
	res := make([]*types.Webhook, 0, 20)

	scanFunc := func(rows *sql.Rows) error {
		webhook, err := scanWebhook(rows, s.key)
		if err != nil {
			return err
		}
		res = append(res, webhook)
		return nil
	}

	err := s.list(ctx, opt, sqlf.Join(webhookColumns, ", "), scanFunc)
	return res, err
}

type WebhookListOptions struct {
	Kind   string
	Cursor *types.Cursor
	*LimitOffset
}

// parseWebhookCursorCond returns the WHERE conditions for the given cursor
func parseWebhookCursorCond(cursor *types.Cursor) (cond *sqlf.Query, err error) {
	if cursor == nil || cursor.Column == "" || cursor.Value == "" {
		return nil, nil
	}

	var operator string
	switch cursor.Direction {
	case "next":
		operator = ">="
	case "prev":
		operator = "<="
	default:
		return nil, errors.Errorf("missing or invalid cursor direction: %q", cursor.Direction)
	}

	if cursor.Column != "id" {
		return nil, errors.Errorf("missing or invalid cursor: %q %q", cursor.Column, cursor.Value)
	}

	return sqlf.Sprintf(fmt.Sprintf("(%s) %s (%%s)", cursor.Column, operator), cursor.Value), nil
}

const webhookListQueryFmtstr = `
SELECT
	%s
FROM webhooks
`

func (s *webhookStore) Count(ctx context.Context, opts WebhookListOptions) (ct int, err error) {
	opts.LimitOffset = nil
	err = s.list(ctx, opts, sqlf.Sprintf("COUNT(*)"), func(rows *sql.Rows) error {
		return rows.Scan(&ct)
	})
	return ct, err
}

func scanWebhook(sc dbutil.Scanner, key encryption.Key) (*types.Webhook, error) {
	var (
		hook      types.Webhook
		keyID     string
		rawSecret string
	)

	var codeHostURL string
	if err := sc.Scan(
		&hook.ID,
		&hook.UUID,
		&hook.CodeHostKind,
		&codeHostURL,
		&dbutil.NullString{S: &rawSecret},
		&hook.CreatedAt,
		&hook.UpdatedAt,
		&dbutil.NullString{S: &keyID},
		&dbutil.NullInt32{N: &hook.CreatedByUserID},
		&dbutil.NullInt32{N: &hook.UpdatedByUserID},
		&hook.Name,
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

	codeHostURN, err := extsvc.NewCodeHostBaseURL(codeHostURL)
	if err != nil {
		return nil, err
	}
	hook.CodeHostURN = codeHostURN

	return &hook, nil
}
