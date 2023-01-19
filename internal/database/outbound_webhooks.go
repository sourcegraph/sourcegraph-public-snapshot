package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OutboundWebhookStore interface {
	basestore.ShareableStore
	Transact(context.Context) (OutboundWebhookStore, error)
	With(basestore.ShareableStore) OutboundWebhookStore
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	Done(error) error

	// Convenience methods to construct job and log stores with the same
	// encryption key.
	ToJobStore() OutboundWebhookJobStore
	ToLogStore() OutboundWebhookLogStore

	Count(context.Context, OutboundWebhookCountOpts) (int64, error)
	Create(context.Context, *types.OutboundWebhook) error
	GetByID(context.Context, int64) (*types.OutboundWebhook, error)
	List(context.Context, OutboundWebhookListOpts) ([]*types.OutboundWebhook, error)
	Delete(context.Context, int64) error
	Update(context.Context, *types.OutboundWebhook) error
}

type OutboundWebhookNotFoundErr struct{ args []any }

func (err OutboundWebhookNotFoundErr) Error() string {
	return fmt.Sprintf("outbound webhook not found: %v", err.args)
}

func (OutboundWebhookNotFoundErr) NotFound() bool { return true }

type outboundWebhookStore struct {
	*basestore.Store
	key encryption.Key
}

func OutboundWebhooksWith(other basestore.ShareableStore, key encryption.Key) OutboundWebhookStore {
	return &outboundWebhookStore{
		Store: basestore.NewWithHandle(other.Handle()),
		key:   key,
	}
}

func (s *outboundWebhookStore) With(other basestore.ShareableStore) OutboundWebhookStore {
	return &outboundWebhookStore{
		Store: s.Store.With(other),
		key:   s.key,
	}
}

func (s *outboundWebhookStore) Transact(ctx context.Context) (OutboundWebhookStore, error) {
	tx, err := s.Store.Transact(ctx)
	return &outboundWebhookStore{
		Store: tx,
		key:   s.key,
	}, err
}

func (s *outboundWebhookStore) ToJobStore() OutboundWebhookJobStore {
	return &outboundWebhookJobStore{
		Store: s.Store,
		key:   s.key,
	}
}

func (s *outboundWebhookStore) ToLogStore() OutboundWebhookLogStore {
	return &outboundWebhookLogStore{
		Store: s.Store,
		key:   s.key,
	}
}

const FilterEventTypeNoScope string = "RESERVED_KEYWORD_MATCH_NULL_SCOPE"

type FilterEventType struct {
	EventType string
	// "foo" matches "foo", NoScope matches NULL, omit to match any scope
	Scope *string
}

type OutboundWebhookCountOpts struct {
	EventTypes []FilterEventType
}

func (opts *OutboundWebhookCountOpts) where() *sqlf.Query {
	// We're going to build up predicates for the subquery on
	// outbound_webhook_event_types, if any.
	preds := []*sqlf.Query{}
	if len(opts.EventTypes) > 0 {
		for _, opt := range opts.EventTypes {
			if opt.Scope == nil {
				preds = append(preds, sqlf.Sprintf(
					// Filter to ones that match the event type, ignoring scope
					"(event_type = %s)",
					opt.EventType,
				))
			} else if *opt.Scope == FilterEventTypeNoScope {
				preds = append(preds, sqlf.Sprintf(
					// Filter to ones that match the event type and have a NULL scope
					"(event_type = %s AND scope IS NULL)",
					opt.EventType,
				))
			} else {
				preds = append(preds, sqlf.Sprintf(
					// Filter to ones that match the event type and scope
					"(event_type = %s AND scope = %s)",
					opt.EventType,
					*opt.Scope,
				))
			}
		}
	}

	var whereClause *sqlf.Query
	if len(preds) > 0 {
		whereClause = sqlf.Sprintf(
			"id IN (%s)",
			sqlf.Sprintf(
				outboundWebhookListSubqueryFmtstr,
				sqlf.Join(preds, "OR"),
			),
		)
	} else {
		whereClause = sqlf.Sprintf("TRUE")
	}

	return whereClause
}

func (s *outboundWebhookStore) Count(ctx context.Context, opts OutboundWebhookCountOpts) (int64, error) {
	q := sqlf.Sprintf(
		outboundWebhookCountQueryFmtstr,
		opts.where(),
	)

	var count int64
	err := s.QueryRow(ctx, q).Scan(&count)

	return count, err
}

func (s *outboundWebhookStore) Create(ctx context.Context, webhook *types.OutboundWebhook) error {
	enc, err := s.encryptFields(ctx, webhook.URL, webhook.Secret)
	if err != nil {
		return errors.Wrap(err, "encrypting fields")
	}

	eventTypes, err := eventTypesToInsertableRows(webhook.EventTypes)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		outboundWebhookCreateQueryFmtstr,
		webhook.CreatedBy,
		webhook.UpdatedBy,
		dbutil.NullStringColumn(enc.keyID),
		[]byte(enc.url),
		[]byte(enc.secret),
		eventTypes,
		sqlf.Join(outboundWebhookColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scanOutboundWebhook(webhook, row); err != nil {
		return errors.Wrap(err, "scanning outbound webhook")
	}

	return nil
}

func (s *outboundWebhookStore) GetByID(ctx context.Context, id int64) (*types.OutboundWebhook, error) {
	q := sqlf.Sprintf(
		outboundWebhookGetByIDQueryFmtstr,
		sqlf.Join(outboundWebhookWithEventTypesColumns, ","),
		id,
	)

	webhook := types.OutboundWebhook{}
	if err := s.scanOutboundWebhook(&webhook, s.QueryRow(ctx, q)); err == sql.ErrNoRows {
		return nil, OutboundWebhookNotFoundErr{args: []any{id}}
	} else if err != nil {
		return nil, err
	}

	return &webhook, nil
}

type ScopedEventType struct {
	EventType string
	Scope     *string
}

func (s *outboundWebhookStore) Delete(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(outboundWebhookDeleteQueryFmtstr, id)
	_, err := s.Query(ctx, q)

	return err
}

type OutboundWebhookListOpts struct {
	*LimitOffset
	OutboundWebhookCountOpts
}

func (s *outboundWebhookStore) List(ctx context.Context, opts OutboundWebhookListOpts) ([]*types.OutboundWebhook, error) {
	q := sqlf.Sprintf(
		outboundWebhookListQueryFmtstr,
		sqlf.Join(outboundWebhookWithEventTypesColumns, ","),
		opts.where(),
		opts.LimitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	webhooks := []*types.OutboundWebhook{}
	for rows.Next() {
		var webhook types.OutboundWebhook
		if err := s.scanOutboundWebhook(&webhook, rows); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, &webhook)
	}

	return webhooks, nil
}

func (s *outboundWebhookStore) Update(ctx context.Context, webhook *types.OutboundWebhook) error {
	enc, err := s.encryptFields(ctx, webhook.URL, webhook.Secret)
	if err != nil {
		return errors.Wrap(err, "encrypting fields")
	}

	eventTypes, err := eventTypesToInsertableRows(webhook.EventTypes)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		outboundWebhookUpdateQueryFmtstr,
		webhook.UpdatedBy,
		dbutil.NullStringColumn(enc.keyID),
		[]byte(enc.url),
		[]byte(enc.secret),
		webhook.ID,
		eventTypes,
		sqlf.Join(outboundWebhookColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scanOutboundWebhook(webhook, row); err != nil {
		return errors.Wrap(err, "scanning outbound webhook")
	}

	return nil
}

var outboundWebhookColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("created_by"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_by"),
	sqlf.Sprintf("updated_at"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("url"),
	sqlf.Sprintf("secret"),
}

var outboundWebhookWithEventTypesColumns = append(
	outboundWebhookColumns,
	sqlf.Sprintf("event_types"),
)

// In the GetByID and List methods, we use the
// outbound_webhooks_with_event_types to retrieve the event types associated
// with a webhook in an atomic single query. When we're using CTEs to insert,
// delete, or update the webhook and/or its associated event types, however, we
// have to recalculate the value of the view's event_types column in place,
// since the view isn't updated until the query is actually committed.
//
// This blob of PostgreSQL ick does the same as the view, allowing us to use the
// same scanner and column definitions for the mutators as we do for the
// accessors.
const outboundWebhookEventTypeColumnFmtstr = `
array_to_json(
	array(
		SELECT
			json_build_object(
				'id', id,
				'outbound_webhook_id', outbound_webhook_id,
				'event_type', event_type,
				'scope', scope
			)
		FROM
			event_types
	)
) AS event_types
`

const outboundWebhookCountQueryFmtstr = `
-- source: internal/database/outbound_webhooks.go:Count
SELECT
	COUNT(*)
FROM
	outbound_webhooks_with_event_types
WHERE
	%s
`

const outboundWebhookCreateQueryFmtstr = `
-- source: internal/database/outbound_webhooks.go:Create
WITH
	outbound_webhook AS (
		INSERT INTO
			outbound_webhooks (
				created_by,
				updated_by,
				encryption_key_id,
				url,
				secret
			)
			VALUES (
				%s,
				%s,
				%s,
				%s,
				%s
			)
			RETURNING
				*
	),
	data (event_type, scope) AS (
		VALUES %s
	),
	event_types AS (
		INSERT INTO
			outbound_webhook_event_types (
				outbound_webhook_id,
				event_type,
				scope
			)
		SELECT
			outbound_webhook.id,
			data.event_type,
			data.scope
		FROM
			outbound_webhook
		CROSS JOIN
			data
		RETURNING *
	)
SELECT
	%s,
	` + outboundWebhookEventTypeColumnFmtstr + `
FROM
	outbound_webhook
`

const outboundWebhookDeleteQueryFmtstr = `
-- source: internal/database/outbound_webhooks.go:Delete
DELETE FROM
	outbound_webhooks
WHERE
	id = %s
`

const outboundWebhookGetByIDQueryFmtstr = `
-- source: internal/database/outbound_webhooks.go:GetByID
SELECT
	%s
FROM
	outbound_webhooks_with_event_types
WHERE
	id = %s
`

const outboundWebhookListQueryFmtstr = `
-- source: internal/database/outbound_webhooks.go:List
SELECT
	%s
FROM
	outbound_webhooks_with_event_types
WHERE
	%s
ORDER BY
	id ASC
%s -- LIMIT
`

const outboundWebhookListSubqueryFmtstr = `
SELECT
	outbound_webhook_id
FROM
	outbound_webhook_event_types
WHERE
	%s
`

const outboundWebhookUpdateQueryFmtstr = `
-- source: internal/database/outbound_webhooks.go:Update
WITH
	outbound_webhook AS (
		UPDATE
			outbound_webhooks
		SET
			updated_at = NOW(),
			updated_by = %s,
			encryption_key_id = %s,
			url = %s,
			secret = %s
		WHERE
			id = %s
		RETURNING
			*
	),
	delete_all_event_types AS (
		DELETE FROM
			outbound_webhook_event_types
		WHERE
			outbound_webhook_id IN (
				SELECT
					id
				FROM
					outbound_webhook
			)
	),
	data (event_type, scope) AS (
		VALUES %s
	),
	event_types AS (
		INSERT INTO
			outbound_webhook_event_types (
				outbound_webhook_id,
				event_type,
				scope
			)
		SELECT
			outbound_webhook.id,
			data.event_type,
			data.scope
		FROM
			outbound_webhook
		CROSS JOIN
			data
		RETURNING
			*
	)
SELECT
	%s,
	` + outboundWebhookEventTypeColumnFmtstr + `
FROM
	outbound_webhook
`

func (s *outboundWebhookStore) scanOutboundWebhook(webhook *types.OutboundWebhook, sc dbutil.Scanner) error {
	var (
		rawURL, rawSecret []byte
		keyID             string
		rawEventTypes     string
	)

	if err := sc.Scan(
		&webhook.ID,
		&dbutil.NullInt32{N: &webhook.CreatedBy},
		&webhook.CreatedAt,
		&dbutil.NullInt32{N: &webhook.UpdatedBy},
		&webhook.UpdatedAt,
		&dbutil.NullString{S: &keyID},
		&rawURL,
		&rawSecret,
		&rawEventTypes,
	); err != nil {
		return err
	}

	webhook.URL = encryption.NewEncrypted(string(rawURL), keyID, s.key)
	webhook.Secret = encryption.NewEncrypted(string(rawSecret), keyID, s.key)

	if err := json.Unmarshal([]byte(rawEventTypes), &webhook.EventTypes); err != nil {
		return errors.Wrap(err, "unmarshalling event types")
	}

	return nil
}

type encryptedFields struct {
	url    string
	secret string
	keyID  string
}

func (s *outboundWebhookStore) encryptFields(ctx context.Context, url, secret *encryption.Encryptable) (ef encryptedFields, err error) {
	var urlKey, secretKey string

	ef.url, urlKey, err = url.Encrypt(ctx, s.key)
	if err != nil {
		return
	}
	ef.secret, secretKey, err = secret.Encrypt(ctx, s.key)
	if err != nil {
		return
	}

	// These should always match, whether we're encrypting or not.
	if urlKey != secretKey {
		err = errors.New("different key IDs returned when using the same key")
	}

	ef.keyID = urlKey
	return
}

var errOutboundWebhookHasNoEventTypes = errors.New("an outbound webhook must have at least one event type")

func eventTypesToInsertableRows(eventTypes []types.OutboundWebhookEventType) (*sqlf.Query, error) {
	if len(eventTypes) == 0 {
		return nil, errOutboundWebhookHasNoEventTypes
	}

	rows := make([]*sqlf.Query, len(eventTypes))
	for i, eventType := range eventTypes {
		rows[i] = sqlf.Sprintf("(%s, %s)", eventType.EventType, eventType.Scope)
	}

	return sqlf.Join(rows, ","), nil
}
