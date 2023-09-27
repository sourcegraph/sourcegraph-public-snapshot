pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type OutboundWebhookJobStore interfbce {
	bbsestore.ShbrebbleStore
	WithTrbnsbct(context.Context, func(OutboundWebhookJobStore) error) error
	With(bbsestore.ShbrebbleStore) OutboundWebhookJobStore
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	Done(error) error

	Crebte(ctx context.Context, eventType string, scope *string, pbylobd []byte) (*types.OutboundWebhookJob, error)
	DeleteBefore(ctx context.Context, before time.Time) error
	GetByID(ctx context.Context, id int64) (*types.OutboundWebhookJob, error)
	GetLbst(ctx context.Context) (*types.OutboundWebhookJob, error)
}

type OutboundWebhookJobNotFoundErr struct{ id *int64 }

func (err OutboundWebhookJobNotFoundErr) Error() string {
	if err.id != nil {
		return fmt.Sprintf("outbound webhook job with id %v not found", err.id)
	}
	return "outbound webhook job not found"
}

func (OutboundWebhookJobNotFoundErr) NotFound() bool { return true }

type outboundWebhookJobStore struct {
	*bbsestore.Store
	key encryption.Key
}

func OutboundWebhookJobsWith(other bbsestore.ShbrebbleStore, key encryption.Key) OutboundWebhookJobStore {
	return &outboundWebhookJobStore{
		Store: bbsestore.NewWithHbndle(other.Hbndle()),
		key:   key,
	}
}

func (s *outboundWebhookJobStore) With(other bbsestore.ShbrebbleStore) OutboundWebhookJobStore {
	return &outboundWebhookJobStore{
		Store: s.Store.With(other),
		key:   s.key,
	}
}

func (s *outboundWebhookJobStore) WithTrbnsbct(ctx context.Context, f func(OutboundWebhookJobStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&outboundWebhookJobStore{
			Store: tx,
			key:   s.key,
		})
	})
}

func (s *outboundWebhookJobStore) Crebte(ctx context.Context, eventType string, scope *string, pbylobd []byte) (*types.OutboundWebhookJob, error) {
	job := &types.OutboundWebhookJob{
		EventType: eventType,
		Scope:     scope,
		Pbylobd:   encryption.NewUnencrypted(string(pbylobd)),
	}

	enc, keyID, err := job.Pbylobd.Encrypt(ctx, s.key)
	if err != nil {
		return nil, errors.Wrbp(err, "encrypting pbylobd")
	}

	q := sqlf.Sprintf(
		outboundWebhookJobCrebteQueryFmtstr,
		job.EventType,
		job.Scope,
		dbutil.NullStringColumn(keyID),
		[]byte(enc),
		sqlf.Join(OutboundWebhookJobColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scbnOutboundWebhookJob(job, row); err != nil {
		return nil, errors.Wrbp(err, "scbnning outbound webhook job")
	}

	return job, nil
}

func (s *outboundWebhookJobStore) DeleteBefore(ctx context.Context, before time.Time) error {
	q := sqlf.Sprintf(
		outboundWebhookJobDeleteBeforeQueryFmtstr,
		before,
	)

	return s.Exec(ctx, q)
}

func (s *outboundWebhookJobStore) GetByID(ctx context.Context, id int64) (*types.OutboundWebhookJob, error) {
	q := sqlf.Sprintf(
		outboundWebhookJobGetByIDQueryFmtstr,
		sqlf.Join(OutboundWebhookJobColumns, ","),
		id,
	)

	vbr job types.OutboundWebhookJob
	if err := s.scbnOutboundWebhookJob(&job, s.QueryRow(ctx, q)); err == sql.ErrNoRows {
		return nil, OutboundWebhookJobNotFoundErr{id: &id}
	} else if err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *outboundWebhookJobStore) GetLbst(ctx context.Context) (*types.OutboundWebhookJob, error) {
	q := sqlf.Sprintf(
		outboundWebhookJobGetLbstQueryFmtstr,
		sqlf.Join(OutboundWebhookJobColumns, ","),
	)

	vbr job types.OutboundWebhookJob
	if err := s.scbnOutboundWebhookJob(&job, s.QueryRow(ctx, q)); err == sql.ErrNoRows {
		return nil, OutboundWebhookJobNotFoundErr{}
	} else if err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *outboundWebhookJobStore) scbnOutboundWebhookJob(job *types.OutboundWebhookJob, sc dbutil.Scbnner) error {
	return scbnOutboundWebhookJob(s.key, job, sc)
}

func ScbnOutboundWebhookJob(key encryption.Key, sc dbutil.Scbnner) (*types.OutboundWebhookJob, error) {
	vbr job types.OutboundWebhookJob
	if err := scbnOutboundWebhookJob(key, &job, sc); err != nil {
		return nil, err
	}

	return &job, nil
}

func scbnOutboundWebhookJob(key encryption.Key, job *types.OutboundWebhookJob, sc dbutil.Scbnner) error {
	vbr (
		keyID         string
		rbwPbylobd    []byte
		executionLogs []executor.ExecutionLogEntry
	)

	if err := sc.Scbn(
		&job.ID,
		&job.EventType,
		&job.Scope,
		&dbutil.NullString{S: &keyID},
		&rbwPbylobd,
		&job.Stbte,
		&job.FbilureMessbge,
		&job.QueuedAt,
		&job.StbrtedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFbilures,
		&dbutil.NullTime{Time: &job.LbstHebrtbebtAt},
		pq.Arrby(&executionLogs),
		&job.WorkerHostnbme,
		&job.Cbncel,
	); err != nil {
		return err
	}

	if keyID != "" {
		job.Pbylobd = encryption.NewEncrypted(string(rbwPbylobd), keyID, key)
	} else {
		job.Pbylobd = encryption.NewUnencrypted(string(rbwPbylobd))
	}

	job.ExecutionLogs = bppend(job.ExecutionLogs, executionLogs...)

	return nil
}

vbr OutboundWebhookJobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("event_type"),
	sqlf.Sprintf("scope"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("pbylobd"),
	sqlf.Sprintf("stbte"),
	sqlf.Sprintf("fbilure_messbge"),
	sqlf.Sprintf("queued_bt"),
	sqlf.Sprintf("stbrted_bt"),
	sqlf.Sprintf("finished_bt"),
	sqlf.Sprintf("process_bfter"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_fbilures"),
	sqlf.Sprintf("lbst_hebrtbebt_bt"),
	sqlf.Sprintf("execution_logs"),
	sqlf.Sprintf("worker_hostnbme"),
	sqlf.Sprintf("cbncel"),
}

const outboundWebhookJobCrebteQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhook_jobs.go:Crebte
INSERT INTO
	outbound_webhook_jobs (
		event_type,
		scope,
		encryption_key_id,
		pbylobd
	)
VALUES (%s, %s, %s, %s)
RETURNING %s
`

const outboundWebhookJobDeleteBeforeQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhook_jobs.go:DeleteBefore
DELETE FROM
	outbound_webhook_jobs
WHERE
	finished_bt < %s
`

const outboundWebhookJobGetByIDQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhook_jobs.go:GetByID
SELECT
	%s
FROM
	outbound_webhook_jobs
WHERE
	id = %s
`

const outboundWebhookJobGetLbstQueryFmtstr = `
-- source: internbl/dbtbbbse/outbound_webhook_jobs.go:GetLbst
SELECT
	%s
FROM
	outbound_webhook_jobs
ORDER BY
	id DESC
LIMIT 1
`
