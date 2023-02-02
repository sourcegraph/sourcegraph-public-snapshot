package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OutboundWebhookLogStore interface {
	basestore.ShareableStore
	WithTransact(context.Context, func(OutboundWebhookLogStore) error) error
	With(basestore.ShareableStore) OutboundWebhookLogStore
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	Done(error) error

	CountsForOutboundWebhook(ctx context.Context, outboundWebhookID int64) (total, errored int64, err error)
	Create(context.Context, *types.OutboundWebhookLog) error
	ListForOutboundWebhook(ctx context.Context, opts OutboundWebhookLogListOpts) ([]*types.OutboundWebhookLog, error)
}

type OutboundWebhookLogListOpts struct {
	*LimitOffset
	OnlyErrors        bool
	OutboundWebhookID int64
}

func (opts *OutboundWebhookLogListOpts) where() *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("outbound_webhook_id = %s", opts.OutboundWebhookID),
	}

	if opts.OnlyErrors {
		preds = append(
			preds,
			sqlf.Sprintf("status_code NOT BETWEEN 100 AND 399"),
		)
	}

	return sqlf.Join(preds, "AND")
}

type outboundWebhookLogStore struct {
	*basestore.Store
	key encryption.Key
}

func OutboundWebhookLogsWith(other basestore.ShareableStore, key encryption.Key) OutboundWebhookLogStore {
	return &outboundWebhookLogStore{
		Store: basestore.NewWithHandle(other.Handle()),
		key:   key,
	}
}

func (s *outboundWebhookLogStore) With(other basestore.ShareableStore) OutboundWebhookLogStore {
	return &outboundWebhookLogStore{
		Store: s.Store.With(other),
		key:   s.key,
	}
}

func (s *outboundWebhookLogStore) WithTransact(ctx context.Context, f func(OutboundWebhookLogStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&outboundWebhookLogStore{
			Store: tx,
			key:   s.key,
		})
	})
}

func (s *outboundWebhookLogStore) CountsForOutboundWebhook(ctx context.Context, outboundWebhookID int64) (total, errored int64, err error) {
	q := sqlf.Sprintf(
		outboundWebhookCountsForOutboundWebhookQueryFmtstr,
		outboundWebhookID,
	)

	err = s.QueryRow(ctx, q).Scan(&total, &errored)
	return
}

func (s *outboundWebhookLogStore) Create(ctx context.Context, log *types.OutboundWebhookLog) error {
	rawRequest, _, err := log.Request.Encrypt(ctx, s.key)
	if err != nil {
		return errors.Wrap(err, "encrypting request")
	}

	rawResponse, _, err := log.Response.Encrypt(ctx, s.key)
	if err != nil {
		return errors.Wrap(err, "encrypting response")
	}

	rawError, keyID, err := log.Error.Encrypt(ctx, s.key)
	if err != nil {
		return errors.Wrap(err, "encrypting error")
	}

	q := sqlf.Sprintf(
		outboundWebhookLogCreateQueryFmtstr,
		log.JobID,
		log.OutboundWebhookID,
		log.StatusCode,
		dbutil.NullStringColumn(keyID),
		[]byte(rawRequest),
		[]byte(rawResponse),
		[]byte(rawError),
		sqlf.Join(outboundWebhookLogColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scanOutboundWebhookLog(log, row); err != nil {
		return errors.Wrap(err, "scanning outbound webhook log")
	}

	return nil
}

func (s *outboundWebhookLogStore) ListForOutboundWebhook(ctx context.Context, opts OutboundWebhookLogListOpts) ([]*types.OutboundWebhookLog, error) {
	q := sqlf.Sprintf(
		outboundWebhookLogListForOutboundWebhookQueryFmtstr,
		sqlf.Join(outboundWebhookLogColumns, ","),
		opts.where(),
		opts.LimitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []*types.OutboundWebhookLog{}
	for rows.Next() {
		var log types.OutboundWebhookLog
		if err := s.scanOutboundWebhookLog(&log, rows); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

func (s *outboundWebhookLogStore) scanOutboundWebhookLog(log *types.OutboundWebhookLog, sc dbutil.Scanner) error {
	var (
		keyID       string
		rawRequest  []byte
		rawResponse []byte
		rawError    []byte
	)

	if err := sc.Scan(
		&log.ID,
		&log.JobID,
		&log.OutboundWebhookID,
		&log.SentAt,
		&log.StatusCode,
		&dbutil.NullString{S: &keyID},
		&rawRequest,
		&rawResponse,
		&rawError,
	); err != nil {
		return err
	}

	log.Request = types.NewEncryptedWebhookLogMessage(string(rawRequest), keyID, s.key)
	log.Response = types.NewEncryptedWebhookLogMessage(string(rawResponse), keyID, s.key)
	log.Error = encryption.NewEncrypted(string(rawError), keyID, s.key)

	return nil
}

var outboundWebhookLogColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("job_id"),
	sqlf.Sprintf("outbound_webhook_id"),
	sqlf.Sprintf("sent_at"),
	sqlf.Sprintf("status_code"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("request"),
	sqlf.Sprintf("response"),
	sqlf.Sprintf("error"),
}

const outboundWebhookCountsForOutboundWebhookQueryFmtstr = `
-- source: internal/database/outbound_webhook_logs:CountsForOutboundWebhook
SELECT
	COUNT(*) AS total,
	COUNT(*) FILTER (WHERE status_code NOT BETWEEN 100 AND 399) AS errored
FROM
	outbound_webhook_logs
WHERE
	outbound_webhook_id = %s
`

const outboundWebhookLogCreateQueryFmtstr = `
-- source: internal/database/outbound_webhook_logs.go:Create
INSERT INTO
	outbound_webhook_logs (
		job_id,
		outbound_webhook_id,
		status_code,
		encryption_key_id,
		request,
		response,
		error
	)
VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

const outboundWebhookLogListForOutboundWebhookQueryFmtstr = `
-- source: internal/database/outbound_webhook_logs.go:ListForOutboundWebhook
SELECT
	%s
FROM
	outbound_webhook_logs
WHERE
	%s
ORDER BY
	id DESC
%s -- LIMIT
`
