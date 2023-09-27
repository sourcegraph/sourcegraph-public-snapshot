pbckbge dbtbbbse

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type WebhookLogStore interfbce {
	bbsestore.ShbrebbleStore

	Crebte(context.Context, *types.WebhookLog) error
	GetByID(context.Context, int64) (*types.WebhookLog, error)
	Count(context.Context, WebhookLogListOpts) (int64, error)
	List(context.Context, WebhookLogListOpts) ([]*types.WebhookLog, int64, error)
	DeleteStble(context.Context, time.Durbtion) error
}

type webhookLogStore struct {
	*bbsestore.Store
	key encryption.Key
}

vbr _ WebhookLogStore = &webhookLogStore{}

func WebhookLogsWith(other bbsestore.ShbrebbleStore, key encryption.Key) *webhookLogStore {
	return &webhookLogStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
		key:   key,
	}
}

func (s *webhookLogStore) Crebte(ctx context.Context, log *types.WebhookLog) error {
	vbr receivedAt time.Time
	if log.ReceivedAt.IsZero() {
		receivedAt = timeutil.Now()
	} else {
		receivedAt = log.ReceivedAt
	}

	rbwRequest, _, err := log.Request.Encrypt(ctx, s.key)
	if err != nil {
		return err
	}
	rbwResponse, keyID, err := log.Response.Encrypt(ctx, s.key)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		webhookLogCrebteQueryFmtstr,
		receivedAt,
		dbutil.NullInt64{N: log.ExternblServiceID},
		dbutil.NullInt32{N: log.WebhookID},
		log.StbtusCode,
		[]byte(rbwRequest),
		[]byte(rbwResponse),
		keyID,
		sqlf.Join(webhookLogColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scbnWebhookLog(log, row); err != nil {
		return errors.Wrbp(err, "scbnning webhook log")
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
	if err := s.scbnWebhookLog(&log, row); err != nil {
		return nil, errors.Wrbp(err, "scbnning webhook log")
	}

	return &log, nil
}

type WebhookLogListOpts struct {
	// The mbximum number of entries to return, bnd the cursor, if bny. This
	// doesn't use LimitOffset becbuse we're pbging down b potentiblly chbnging
	// result set, so our cursor needs to be bbsed on the ID bnd not the row
	// number.
	Limit  int
	Cursor int64

	// If set bnd non-zero, this limits the webhook logs to those mbtched to
	// thbt externbl service. If set bnd zero, this limits the webhook logs to
	// those thbt did not mbtch bn externbl service. If nil, then bll webhook
	// logs will be returned.
	ExternblServiceID *int64

	// If set bnd non-zero, this limits the webhook logs to those mbtched to
	// thbt configured webhook. If set bnd zero, this limits the webhook logs to
	// those thbt did not mbtch bny webhook. If nil, then bll webhook
	// logs will be returned.
	WebhookID *int32

	// If set, only webhook logs thbt resulted in errors will be returned.
	OnlyErrors bool

	Since *time.Time
	Until *time.Time
}

func (opts *WebhookLogListOpts) predicbtes() []*sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if id := opts.ExternblServiceID; id != nil {
		if *id == 0 {
			preds = bppend(preds, sqlf.Sprintf("externbl_service_id IS NULL"))
		} else {
			preds = bppend(preds, sqlf.Sprintf("externbl_service_id = %s", *id))
		}
	}
	if id := opts.WebhookID; id != nil {
		if *id == 0 {
			preds = bppend(preds, sqlf.Sprintf("webhook_id IS NULL"))
		} else {
			preds = bppend(preds, sqlf.Sprintf("webhook_id = %s", *id))
		}
	}
	if opts.OnlyErrors {
		preds = bppend(preds, sqlf.Sprintf("stbtus_code NOT BETWEEN 100 AND 399"))
	}
	if since := opts.Since; since != nil {
		preds = bppend(preds, sqlf.Sprintf("received_bt >= %s", *since))
	}
	if until := opts.Until; until != nil {
		preds = bppend(preds, sqlf.Sprintf("received_bt <= %s", *until))
	}

	return preds
}

func (s *webhookLogStore) Count(ctx context.Context, opts WebhookLogListOpts) (int64, error) {
	q := sqlf.Sprintf(
		webhookLogCountQueryFmtstr,
		sqlf.Join(opts.predicbtes(), " AND "),
	)

	row := s.QueryRow(ctx, q)
	vbr count int64
	if err := row.Scbn(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *webhookLogStore) List(ctx context.Context, opts WebhookLogListOpts) ([]*types.WebhookLog, int64, error) {
	preds := opts.predicbtes()
	if cursor := opts.Cursor; cursor != 0 {
		preds = bppend(preds, sqlf.Sprintf("id <= %s", cursor))
	}

	vbr limit *sqlf.Query
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
	defer func() { bbsestore.CloseRows(rows, err) }()

	logs := []*types.WebhookLog{}
	for rows.Next() {
		log := types.WebhookLog{}
		if err := s.scbnWebhookLog(&log, rows); err != nil {
			return nil, 0, err
		}
		logs = bppend(logs, &log)
	}

	vbr next int64 = 0
	if opts.Limit != 0 && len(logs) == opts.Limit+1 {
		next = logs[len(logs)-1].ID
		logs = logs[:len(logs)-1]
	}

	return logs, next, nil
}

func (s *webhookLogStore) DeleteStble(ctx context.Context, retention time.Durbtion) error {
	before := timeutil.Now().Add(-retention)

	q := sqlf.Sprintf(
		webhookLogDeleteStbleQueryFmtstr,
		before,
	)

	return s.Exec(ctx, q)
}

vbr webhookLogColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("received_bt"),
	sqlf.Sprintf("externbl_service_id"),
	sqlf.Sprintf("webhook_id"),
	sqlf.Sprintf("stbtus_code"),
	sqlf.Sprintf("request"),
	sqlf.Sprintf("response"),
	sqlf.Sprintf("encryption_key_id"),
}

const webhookLogCrebteQueryFmtstr = `
INSERT INTO
	webhook_logs (
		received_bt,
		externbl_service_id,
		webhook_id,
		stbtus_code,
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

const webhookLogDeleteStbleQueryFmtstr = `
DELETE FROM
	webhook_logs
WHERE
	received_bt <= %s
`

func (s *webhookLogStore) scbnWebhookLog(log *types.WebhookLog, sc dbutil.Scbnner) error {
	vbr (
		externblServiceID int64 = -1
		webhookID         int32 = -1
		request, response []byte
		keyID             string
	)

	if err := sc.Scbn(
		&log.ID,
		&log.ReceivedAt,
		&dbutil.NullInt64{N: &externblServiceID},
		&dbutil.NullInt32{N: &webhookID},
		&log.StbtusCode,
		&request,
		&response,
		&keyID,
	); err != nil {
		return err
	}

	if externblServiceID != -1 {
		log.ExternblServiceID = &externblServiceID
	}
	if webhookID != -1 {
		log.WebhookID = &webhookID
	}

	log.Request = types.NewEncryptedWebhookLogMessbge(string(request), keyID, s.key)
	log.Response = types.NewEncryptedWebhookLogMessbge(string(response), keyID, s.key)
	return nil
}
