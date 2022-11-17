package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type WebhookLogStore interface {
	basestore.ShareableStore

	Create(context.Context, *types.WebhookLog) error
	GetByID(context.Context, int64) (*types.WebhookLog, error)
	Count(context.Context, WebhookLogListOpts) (int64, error)
	List(context.Context, WebhookLogListOpts) ([]*types.WebhookLog, int64, error)
	DeleteStale(context.Context, time.Duration) error
}

type webhookLogStore struct {
	*basestore.Store
	key encryption.Key
}

var _ WebhookLogStore = &webhookLogStore{}

func WebhookLogsWith(other basestore.ShareableStore, key encryption.Key) *webhookLogStore {
	return &webhookLogStore{
		Store: basestore.NewWithHandle(other.Handle()),
		key:   key,
	}
}

func (s *webhookLogStore) Create(ctx context.Context, log *types.WebhookLog) error {
	var receivedAt time.Time
	if log.ReceivedAt.IsZero() {
		receivedAt = timeutil.Now()
	} else {
		receivedAt = log.ReceivedAt
	}

	rawRequest, _, err := log.Request.Encrypt(ctx, s.key)
	if err != nil {
		return err
	}
	rawResponse, keyID, err := log.Response.Encrypt(ctx, s.key)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		webhookLogCreateQueryFmtstr,
		receivedAt,
		dbutil.NullInt64{N: log.ExternalServiceID},
		dbutil.NullInt32{N: log.WebhookID},
		log.StatusCode,
		[]byte(rawRequest),
		[]byte(rawResponse),
		keyID,
		sqlf.Join(webhookLogColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scanWebhookLog(log, row); err != nil {
		return errors.Wrap(err, "scanning webhook log")
	}

	return nil
}

func (s *webhookLogStore) GetByID(ctx context.Context, id int64) (*types.WebhookLog, error) {
	q := sqlf.Sprintf(
		webhookLogGetByIDQueryFmtstr,
		sqlf.Join(webhookLogColumns, ", "),
		id,
	)

	row := s.QueryRow(ctx, q)
	log := types.WebhookLog{}
	if err := s.scanWebhookLog(&log, row); err != nil {
		return nil, errors.Wrap(err, "scanning webhook log")
	}

	return &log, nil
}

type WebhookLogListOpts struct {
	// The maximum number of entries to return, and the cursor, if any. This
	// doesn't use LimitOffset because we're paging down a potentially changing
	// result set, so our cursor needs to be based on the ID and not the row
	// number.
	Limit  int
	Cursor int64

	// If set and non-zero, this limits the webhook logs to those matched to
	// that external service. If set and zero, this limits the webhook logs to
	// those that did not match an external service. If nil, then all webhook
	// logs will be returned.
	ExternalServiceID *int64

	// If set and non-zero, this limits the webhook logs to those matched to
	// that configured webhook. If set and zero, this limits the webhook logs to
	// those that did not match any webhook. If nil, then all webhook
	// logs will be returned.
	WebhookID *int32

	// If set, only webhook logs that resulted in errors will be returned.
	OnlyErrors bool

	Since *time.Time
	Until *time.Time
}

func (opts *WebhookLogListOpts) predicates() []*sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if id := opts.ExternalServiceID; id != nil {
		if *id == 0 {
			preds = append(preds, sqlf.Sprintf("external_service_id IS NULL"))
		} else {
			preds = append(preds, sqlf.Sprintf("external_service_id = %s", *id))
		}
	}
	if id := opts.WebhookID; id != nil {
		if *id == 0 {
			preds = append(preds, sqlf.Sprintf("webhook_id IS NULL"))
		} else {
			preds = append(preds, sqlf.Sprintf("webhook_id = %s", *id))
		}
	}
	if opts.OnlyErrors {
		preds = append(preds, sqlf.Sprintf("status_code NOT BETWEEN 100 AND 399"))
	}
	if since := opts.Since; since != nil {
		preds = append(preds, sqlf.Sprintf("received_at >= %s", *since))
	}
	if until := opts.Until; until != nil {
		preds = append(preds, sqlf.Sprintf("received_at <= %s", *until))
	}

	return preds
}

func (s *webhookLogStore) Count(ctx context.Context, opts WebhookLogListOpts) (int64, error) {
	q := sqlf.Sprintf(
		webhookLogCountQueryFmtstr,
		sqlf.Join(opts.predicates(), " AND "),
	)

	row := s.QueryRow(ctx, q)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *webhookLogStore) List(ctx context.Context, opts WebhookLogListOpts) ([]*types.WebhookLog, int64, error) {
	preds := opts.predicates()
	if cursor := opts.Cursor; cursor != 0 {
		preds = append(preds, sqlf.Sprintf("id <= %s", cursor))
	}

	var limit *sqlf.Query
	if opts.Limit != 0 {
		limit = sqlf.Sprintf("LIMIT %s", opts.Limit+1)
	} else {
		limit = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(
		webhookLogListQueryFmtstr,
		sqlf.Join(webhookLogColumns, ", "),
		sqlf.Join(preds, " AND "),
		limit,
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	defer func() { basestore.CloseRows(rows, err) }()

	logs := []*types.WebhookLog{}
	for rows.Next() {
		log := types.WebhookLog{}
		if err := s.scanWebhookLog(&log, rows); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &log)
	}

	var next int64 = 0
	if opts.Limit != 0 && len(logs) == opts.Limit+1 {
		next = logs[len(logs)-1].ID
		logs = logs[:len(logs)-1]
	}

	return logs, next, nil
}

func (s *webhookLogStore) DeleteStale(ctx context.Context, retention time.Duration) error {
	before := timeutil.Now().Add(-retention)

	q := sqlf.Sprintf(
		webhookLogDeleteStaleQueryFmtstr,
		before,
	)

	return s.Exec(ctx, q)
}

var webhookLogColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("received_at"),
	sqlf.Sprintf("external_service_id"),
	sqlf.Sprintf("webhook_id"),
	sqlf.Sprintf("status_code"),
	sqlf.Sprintf("request"),
	sqlf.Sprintf("response"),
	sqlf.Sprintf("encryption_key_id"),
}

const webhookLogCreateQueryFmtstr = `
INSERT INTO
	webhook_logs (
		received_at,
		external_service_id,
		webhook_id,
		status_code,
		request,
		response,
		encryption_key_id
	)
	VALUES (
		%s,
		%s,
		%s,
		%s,
		%s,
		%s,
		%s
	)
	RETURNING %s
`

const webhookLogGetByIDQueryFmtstr = `
SELECT
	%s
FROM
	webhook_logs
WHERE
	id = %s
`

const webhookLogCountQueryFmtstr = `
SELECT
	COUNT(id)
FROM
	webhook_logs
WHERE
	%s
`

const webhookLogListQueryFmtstr = `
SELECT
	%s
FROM
	webhook_logs
WHERE
	%s
ORDER BY
	id DESC
%s -- LIMIT
`

const webhookLogDeleteStaleQueryFmtstr = `
DELETE FROM
	webhook_logs
WHERE
	received_at <= %s
`

func (s *webhookLogStore) scanWebhookLog(log *types.WebhookLog, sc dbutil.Scanner) error {
	var (
		externalServiceID int64 = -1
		webhookID         int32 = -1
		request, response []byte
		keyID             string
	)

	if err := sc.Scan(
		&log.ID,
		&log.ReceivedAt,
		&dbutil.NullInt64{N: &externalServiceID},
		&dbutil.NullInt32{N: &webhookID},
		&log.StatusCode,
		&request,
		&response,
		&keyID,
	); err != nil {
		return err
	}

	if externalServiceID != -1 {
		log.ExternalServiceID = &externalServiceID
	}
	if webhookID != -1 {
		log.WebhookID = &webhookID
	}

	log.Request = types.NewEncryptedWebhookLogMessage(string(request), keyID, s.key)
	log.Response = types.NewEncryptedWebhookLogMessage(string(response), keyID, s.key)
	return nil
}
