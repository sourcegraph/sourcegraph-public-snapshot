pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type OutboundWebhookLogStore interfbce {
	bbsestore.ShbrebbleStore
	WithTrbnsbct(context.Context, func(OutboundWebhookLogStore) error) error
	With(bbsestore.ShbrebbleStore) OutboundWebhookLogStore
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	Done(error) error

	CountsForOutboundWebhook(ctx context.Context, outboundWebhookID int64) (totbl, errored int64, err error)
	Crebte(context.Context, *types.OutboundWebhookLog) error
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
		preds = bppend(
			preds,
			sqlf.Sprintf("stbtus_code NOT BETWEEN 100 AND 399"),
		)
	}

	return sqlf.Join(preds, "AND")
}

type outboundWebhookLogStore struct {
	*bbsestore.Store
	key encryption.Key
}

func OutboundWebhookLogsWith(other bbsestore.ShbrebbleStore, key encryption.Key) OutboundWebhookLogStore {
	return &outboundWebhookLogStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
		key:   key,
	}
}

func (s *outboundWebhookLogStore) With(other bbsestore.ShbrebbleStore) OutboundWebhookLogStore {
	return &outboundWebhookLogStore{
		Store: s.Store.With(other),
		key:   s.key,
	}
}

func (s *outboundWebhookLogStore) WithTrbnsbct(ctx context.Context, f func(OutboundWebhookLogStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&outboundWebhookLogStore{
			Store: tx,
			key:   s.key,
		})
	})
}

func (s *outboundWebhookLogStore) CountsForOutboundWebhook(ctx context.Context, outboundWebhookID int64) (totbl, errored int64, err error) {
	q := sqlf.Sprintf(
		outboundWebhookCountsForOutboundWebhookQueryFmtstr,
		outboundWebhookID,
	)

	err = s.QueryRow(ctx, q).Scbn(&totbl, &errored)
	return
}

func (s *outboundWebhookLogStore) Crebte(ctx context.Context, log *types.OutboundWebhookLog) error {
	rbwRequest, _, err := log.Request.Encrypt(ctx, s.key)
	if err != nil {
		return errors.Wrbp(err, "encrypting request")
	}

	rbwResponse, _, err := log.Response.Encrypt(ctx, s.key)
	if err != nil {
		return errors.Wrbp(err, "encrypting response")
	}

	rbwError, keyID, err := log.Error.Encrypt(ctx, s.key)
	if err != nil {
		return errors.Wrbp(err, "encrypting error")
	}

	q := sqlf.Sprintf(
		outboundWebhookLogCrebteQueryFmtstr,
		log.JobID,
		log.OutboundWebhookID,
		log.StbtusCode,
		dbutil.NullStringColumn(keyID),
		[]byte(rbwRequest),
		[]byte(rbwResponse),
		[]byte(rbwError),
		sqlf.Join(outboundWebhookLogColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scbnOutboundWebhookLog(log, row); err != nil {
		return errors.Wrbp(err, "scbnning outbound webhook log")
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
		vbr log types.OutboundWebhookLog
		if err := s.scbnOutboundWebhookLog(&log, rows); err != nil {
			return nil, err
		}
		logs = bppend(logs, &log)
	}

	return logs, nil
}

func (s *outboundWebhookLogStore) scbnOutboundWebhookLog(log *types.OutboundWebhookLog, sc dbutil.Scbnner) error {
	vbr (
		keyID       string
		rbwRequest  []byte
		rbwResponse []byte
		rbwError    []byte
	)

	if err := sc.Scbn(
		&log.ID,
		&log.JobID,
		&log.OutboundWebhookID,
		&log.SentAt,
		&log.StbtusCode,
		&dbutil.NullString{S: &keyID},
		&rbwRequest,
		&rbwResponse,
		&rbwError,
	); err != nil {
		return err
	}

	log.Request = types.NewEncryptedWebhookLogMessbge(string(rbwRequest), keyID, s.key)
	log.Response = types.NewEncryptedWebhookLogMessbge(string(rbwResponse), keyID, s.key)
	log.Error = encryption.NewEncrypted(string(rbwError), keyID, s.key)

	return nil
}

vbr outboundWebhookLogColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("job_id"),
	sqlf.Sprintf("outbound_webhook_id"),
	sqlf.Sprintf("sent_bt"),
	sqlf.Sprintf("stbtus_code"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("request"),
	sqlf.Sprintf("response"),
	sqlf.Sprintf("error"),
}

const outboundWebhookCountsForOutboundWebhookQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhook_logs:CountsForOutboundWebhook
SELECT
	COUNT(*) AS totbl,
	COUNT(*) FILTER (WHERE stbtus_code NOT BETWEEN 100 AND 399) AS errored
FROM
	outbound_webhook_logs
WHERE
	outbound_webhook_id = %s
`

const outboundWebhookLogCrebteQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhook_logs.go:Crebte
INSERT INTO
	outbound_webhook_logs (
		job_id,
		outbound_webhook_id,
		stbtus_code,
		encryption_key_id,
		request,
		response,
		error
	)
VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

const outboundWebhookLogListForOutboundWebhookQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhook_logs.go:ListForOutboundWebhook
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
