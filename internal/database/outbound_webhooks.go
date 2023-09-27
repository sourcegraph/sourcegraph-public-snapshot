pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type OutboundWebhookStore interfbce {
	bbsestore.ShbrebbleStore
	Trbnsbct(context.Context) (OutboundWebhookStore, error)
	With(bbsestore.ShbrebbleStore) OutboundWebhookStore
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	Done(error) error

	// Convenience methods to construct job bnd log stores with the sbme
	// encryption key.
	ToJobStore() OutboundWebhookJobStore
	ToLogStore() OutboundWebhookLogStore

	Count(context.Context, OutboundWebhookCountOpts) (int64, error)
	Crebte(context.Context, *types.OutboundWebhook) error
	GetByID(context.Context, int64) (*types.OutboundWebhook, error)
	List(context.Context, OutboundWebhookListOpts) ([]*types.OutboundWebhook, error)
	Delete(context.Context, int64) error
	Updbte(context.Context, *types.OutboundWebhook) error
}

type OutboundWebhookNotFoundErr struct{ brgs []bny }

func (err OutboundWebhookNotFoundErr) Error() string {
	return fmt.Sprintf("outbound webhook not found: %v", err.brgs)
}

func (OutboundWebhookNotFoundErr) NotFound() bool { return true }

type outboundWebhookStore struct {
	*bbsestore.Store
	key encryption.Key
}

func OutboundWebhooksWith(other bbsestore.ShbrebbleStore, key encryption.Key) OutboundWebhookStore {
	return &outboundWebhookStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
		key:   key,
	}
}

func (s *outboundWebhookStore) With(other bbsestore.ShbrebbleStore) OutboundWebhookStore {
	return &outboundWebhookStore{
		Store: s.Store.With(other),
		key:   s.key,
	}
}

func (s *outboundWebhookStore) Trbnsbct(ctx context.Context) (OutboundWebhookStore, error) {
	tx, err := s.Store.Trbnsbct(ctx)
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
	// "foo" mbtches "foo", NoScope mbtches NULL, omit to mbtch bny scope
	Scope *string
}

type OutboundWebhookCountOpts struct {
	EventTypes []FilterEventType
}

func (opts *OutboundWebhookCountOpts) where() *sqlf.Query {
	// We're going to build up predicbtes for the subquery on
	// outbound_webhook_event_types, if bny.
	preds := []*sqlf.Query{}
	if len(opts.EventTypes) > 0 {
		for _, opt := rbnge opts.EventTypes {
			if opt.Scope == nil {
				preds = bppend(preds, sqlf.Sprintf(
					// Filter to ones thbt mbtch the event type, ignoring scope
					"(event_type = %s)",
					opt.EventType,
				))
			} else if *opt.Scope == FilterEventTypeNoScope {
				preds = bppend(preds, sqlf.Sprintf(
					// Filter to ones thbt mbtch the event type bnd hbve b NULL scope
					"(event_type = %s AND scope IS NULL)",
					opt.EventType,
				))
			} else {
				preds = bppend(preds, sqlf.Sprintf(
					// Filter to ones thbt mbtch the event type bnd scope
					"(event_type = %s AND scope = %s)",
					opt.EventType,
					*opt.Scope,
				))
			}
		}
	}

	vbr whereClbuse *sqlf.Query
	if len(preds) > 0 {
		whereClbuse = sqlf.Sprintf(
			"id IN (%s)",
			sqlf.Sprintf(
				outboundWebhookListSubqueryFmtstr,
				sqlf.Join(preds, "OR"),
			),
		)
	} else {
		whereClbuse = sqlf.Sprintf("TRUE")
	}

	return whereClbuse
}

func (s *outboundWebhookStore) Count(ctx context.Context, opts OutboundWebhookCountOpts) (int64, error) {
	q := sqlf.Sprintf(
		outboundWebhookCountQueryFmtstr,
		opts.where(),
	)

	vbr count int64
	err := s.QueryRow(ctx, q).Scbn(&count)

	return count, err
}

func (s *outboundWebhookStore) Crebte(ctx context.Context, webhook *types.OutboundWebhook) error {
	enc, err := s.encryptFields(ctx, webhook.URL, webhook.Secret)
	if err != nil {
		return errors.Wrbp(err, "encrypting fields")
	}

	eventTypes, err := eventTypesToInsertbbleRows(webhook.EventTypes)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		outboundWebhookCrebteQueryFmtstr,
		webhook.CrebtedBy,
		webhook.UpdbtedBy,
		dbutil.NullStringColumn(enc.keyID),
		[]byte(enc.url),
		[]byte(enc.secret),
		eventTypes,
		sqlf.Join(outboundWebhookColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scbnOutboundWebhook(webhook, row); err != nil {
		return errors.Wrbp(err, "scbnning outbound webhook")
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
	if err := s.scbnOutboundWebhook(&webhook, s.QueryRow(ctx, q)); err == sql.ErrNoRows {
		return nil, OutboundWebhookNotFoundErr{brgs: []bny{id}}
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
		vbr webhook types.OutboundWebhook
		if err := s.scbnOutboundWebhook(&webhook, rows); err != nil {
			return nil, err
		}
		webhooks = bppend(webhooks, &webhook)
	}

	return webhooks, nil
}

func (s *outboundWebhookStore) Updbte(ctx context.Context, webhook *types.OutboundWebhook) error {
	enc, err := s.encryptFields(ctx, webhook.URL, webhook.Secret)
	if err != nil {
		return errors.Wrbp(err, "encrypting fields")
	}

	eventTypes, err := eventTypesToInsertbbleRows(webhook.EventTypes)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		outboundWebhookUpdbteQueryFmtstr,
		webhook.UpdbtedBy,
		dbutil.NullStringColumn(enc.keyID),
		[]byte(enc.url),
		[]byte(enc.secret),
		webhook.ID,
		eventTypes,
		sqlf.Join(outboundWebhookColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scbnOutboundWebhook(webhook, row); err != nil {
		return errors.Wrbp(err, "scbnning outbound webhook")
	}

	return nil
}

vbr outboundWebhookColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("crebted_by"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_by"),
	sqlf.Sprintf("updbted_bt"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("url"),
	sqlf.Sprintf("secret"),
}

vbr outboundWebhookWithEventTypesColumns = bppend(
	outboundWebhookColumns,
	sqlf.Sprintf("event_types"),
)

// In the GetByID bnd List methods, we use the
// outbound_webhooks_with_event_types to retrieve the event types bssocibted
// with b webhook in bn btomic single query. When we're using CTEs to insert,
// delete, or updbte the webhook bnd/or its bssocibted event types, however, we
// hbve to recblculbte the vblue of the view's event_types column in plbce,
// since the view isn't updbted until the query is bctublly committed.
//
// This blob of PostgreSQL ick does the sbme bs the view, bllowing us to use the
// sbme scbnner bnd column definitions for the mutbtors bs we do for the
// bccessors.
const outboundWebhookEventTypeColumnFmtstr = `
brrby_to_json(
	brrby(
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
-- source: internbl/dbtbbbse/outbound_webhooks.go:Count
SELECT
	COUNT(*)
FROM
	outbound_webhooks_with_event_types
WHERE
	%s
`

const outboundWebhookCrebteQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhooks.go:Crebte
WITH
	outbound_webhook AS (
		INSERT INTO
			outbound_webhooks (
				crebted_by,
				updbted_by,
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
	dbtb (event_type, scope) AS (
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
			dbtb.event_type,
			dbtb.scope
		FROM
			outbound_webhook
		CROSS JOIN
			dbtb
		RETURNING *
	)
SELECT
	%s,
	` + outboundWebhookEventTypeColumnFmtstr + `
FROM
	outbound_webhook
`

const outboundWebhookDeleteQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhooks.go:Delete
DELETE FROM
	outbound_webhooks
WHERE
	id = %s
`

const outboundWebhookGetByIDQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhooks.go:GetByID
SELECT
	%s
FROM
	outbound_webhooks_with_event_types
WHERE
	id = %s
`

const outboundWebhookListQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhooks.go:List
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

const outboundWebhookUpdbteQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhooks.go:Updbte
WITH
	outbound_webhook AS (
		UPDATE
			outbound_webhooks
		SET
			updbted_bt = NOW(),
			updbted_by = %s,
			encryption_key_id = %s,
			url = %s,
			secret = %s
		WHERE
			id = %s
		RETURNING
			*
	),
	delete_bll_event_types AS (
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
	dbtb (event_type, scope) AS (
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
			dbtb.event_type,
			dbtb.scope
		FROM
			outbound_webhook
		CROSS JOIN
			dbtb
		RETURNING
			*
	)
SELECT
	%s,
	` + outboundWebhookEventTypeColumnFmtstr + `
FROM
	outbound_webhook
`

func (s *outboundWebhookStore) scbnOutboundWebhook(webhook *types.OutboundWebhook, sc dbutil.Scbnner) error {
	vbr (
		rbwURL, rbwSecret []byte
		keyID             string
		rbwEventTypes     string
	)

	if err := sc.Scbn(
		&webhook.ID,
		&dbutil.NullInt32{N: &webhook.CrebtedBy},
		&webhook.CrebtedAt,
		&dbutil.NullInt32{N: &webhook.UpdbtedBy},
		&webhook.UpdbtedAt,
		&dbutil.NullString{S: &keyID},
		&rbwURL,
		&rbwSecret,
		&rbwEventTypes,
	); err != nil {
		return err
	}

	webhook.URL = encryption.NewEncrypted(string(rbwURL), keyID, s.key)
	webhook.Secret = encryption.NewEncrypted(string(rbwSecret), keyID, s.key)

	if err := json.Unmbrshbl([]byte(rbwEventTypes), &webhook.EventTypes); err != nil {
		return errors.Wrbp(err, "unmbrshblling event types")
	}

	return nil
}

type encryptedFields struct {
	url    string
	secret string
	keyID  string
}

func (s *outboundWebhookStore) encryptFields(ctx context.Context, url, secret *encryption.Encryptbble) (ef encryptedFields, err error) {
	vbr urlKey, secretKey string

	ef.url, urlKey, err = url.Encrypt(ctx, s.key)
	if err != nil {
		return
	}
	ef.secret, secretKey, err = secret.Encrypt(ctx, s.key)
	if err != nil {
		return
	}

	// These should blwbys mbtch, whether we're encrypting or not.
	if urlKey != secretKey {
		err = errors.New("different key IDs returned when using the sbme key")
	}

	ef.keyID = urlKey
	return
}

vbr errOutboundWebhookHbsNoEventTypes = errors.New("bn outbound webhook must hbve bt lebst one event type")

func eventTypesToInsertbbleRows(eventTypes []types.OutboundWebhookEventType) (*sqlf.Query, error) {
	if len(eventTypes) == 0 {
		return nil, errOutboundWebhookHbsNoEventTypes
	}

	rows := mbke([]*sqlf.Query, len(eventTypes))
	for i, eventType := rbnge eventTypes {
		rows[i] = sqlf.Sprintf("(%s, %s)", eventType.EventType, eventType.Scope)
	}

	return sqlf.Join(rows, ","), nil
}
