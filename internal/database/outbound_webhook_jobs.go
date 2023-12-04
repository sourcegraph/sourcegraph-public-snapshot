package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OutboundWebhookJobStore interface {
	basestore.ShareableStore
	WithTransact(context.Context, func(OutboundWebhookJobStore) error) error
	With(basestore.ShareableStore) OutboundWebhookJobStore
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	Done(error) error

	Create(ctx context.Context, eventType string, scope *string, payload []byte) (*types.OutboundWebhookJob, error)
	DeleteBefore(ctx context.Context, before time.Time) error
	GetByID(ctx context.Context, id int64) (*types.OutboundWebhookJob, error)
	GetLast(ctx context.Context) (*types.OutboundWebhookJob, error)
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
	*basestore.Store
	key encryption.Key
}

func OutboundWebhookJobsWith(other basestore.ShareableStore, key encryption.Key) OutboundWebhookJobStore {
	return &outboundWebhookJobStore{
		Store: basestore.NewWithHandle(other.Handle()),
		key:   key,
	}
}

func (s *outboundWebhookJobStore) With(other basestore.ShareableStore) OutboundWebhookJobStore {
	return &outboundWebhookJobStore{
		Store: s.Store.With(other),
		key:   s.key,
	}
}

func (s *outboundWebhookJobStore) WithTransact(ctx context.Context, f func(OutboundWebhookJobStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&outboundWebhookJobStore{
			Store: tx,
			key:   s.key,
		})
	})
}

func (s *outboundWebhookJobStore) Create(ctx context.Context, eventType string, scope *string, payload []byte) (*types.OutboundWebhookJob, error) {
	job := &types.OutboundWebhookJob{
		EventType: eventType,
		Scope:     scope,
		Payload:   encryption.NewUnencrypted(string(payload)),
	}

	enc, keyID, err := job.Payload.Encrypt(ctx, s.key)
	if err != nil {
		return nil, errors.Wrap(err, "encrypting payload")
	}

	q := sqlf.Sprintf(
		outboundWebhookJobCreateQueryFmtstr,
		job.EventType,
		job.Scope,
		dbutil.NullStringColumn(keyID),
		[]byte(enc),
		sqlf.Join(OutboundWebhookJobColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scanOutboundWebhookJob(job, row); err != nil {
		return nil, errors.Wrap(err, "scanning outbound webhook job")
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

	var job types.OutboundWebhookJob
	if err := s.scanOutboundWebhookJob(&job, s.QueryRow(ctx, q)); err == sql.ErrNoRows {
		return nil, OutboundWebhookJobNotFoundErr{id: &id}
	} else if err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *outboundWebhookJobStore) GetLast(ctx context.Context) (*types.OutboundWebhookJob, error) {
	q := sqlf.Sprintf(
		outboundWebhookJobGetLastQueryFmtstr,
		sqlf.Join(OutboundWebhookJobColumns, ","),
	)

	var job types.OutboundWebhookJob
	if err := s.scanOutboundWebhookJob(&job, s.QueryRow(ctx, q)); err == sql.ErrNoRows {
		return nil, OutboundWebhookJobNotFoundErr{}
	} else if err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *outboundWebhookJobStore) scanOutboundWebhookJob(job *types.OutboundWebhookJob, sc dbutil.Scanner) error {
	return scanOutboundWebhookJob(s.key, job, sc)
}

func ScanOutboundWebhookJob(key encryption.Key, sc dbutil.Scanner) (*types.OutboundWebhookJob, error) {
	var job types.OutboundWebhookJob
	if err := scanOutboundWebhookJob(key, &job, sc); err != nil {
		return nil, err
	}

	return &job, nil
}

func scanOutboundWebhookJob(key encryption.Key, job *types.OutboundWebhookJob, sc dbutil.Scanner) error {
	var (
		keyID         string
		rawPayload    []byte
		executionLogs []executor.ExecutionLogEntry
	)

	if err := sc.Scan(
		&job.ID,
		&job.EventType,
		&job.Scope,
		&dbutil.NullString{S: &keyID},
		&rawPayload,
		&job.State,
		&job.FailureMessage,
		&job.QueuedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&dbutil.NullTime{Time: &job.LastHeartbeatAt},
		pq.Array(&executionLogs),
		&job.WorkerHostname,
		&job.Cancel,
	); err != nil {
		return err
	}

	if keyID != "" {
		job.Payload = encryption.NewEncrypted(string(rawPayload), keyID, key)
	} else {
		job.Payload = encryption.NewUnencrypted(string(rawPayload))
	}

	job.ExecutionLogs = append(job.ExecutionLogs, executionLogs...)

	return nil
}

var OutboundWebhookJobColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("event_type"),
	sqlf.Sprintf("scope"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("payload"),
	sqlf.Sprintf("state"),
	sqlf.Sprintf("failure_message"),
	sqlf.Sprintf("queued_at"),
	sqlf.Sprintf("started_at"),
	sqlf.Sprintf("finished_at"),
	sqlf.Sprintf("process_after"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_failures"),
	sqlf.Sprintf("last_heartbeat_at"),
	sqlf.Sprintf("execution_logs"),
	sqlf.Sprintf("worker_hostname"),
	sqlf.Sprintf("cancel"),
}

const outboundWebhookJobCreateQueryFmtstr = `
-- source: internal/database/outbound_webhook_jobs.go:Create
INSERT INTO
	outbound_webhook_jobs (
		event_type,
		scope,
		encryption_key_id,
		payload
	)
VALUES (%s, %s, %s, %s)
RETURNING %s
`

const outboundWebhookJobDeleteBeforeQueryFmtstr = `
-- source: internal/database/outbound_webhook_jobs.go:DeleteBefore
DELETE FROM
	outbound_webhook_jobs
WHERE
	finished_at < %s
`

const outboundWebhookJobGetByIDQueryFmtstr = `
-- source: internal/database/outbound_webhook_jobs.go:GetByID
SELECT
	%s
FROM
	outbound_webhook_jobs
WHERE
	id = %s
`

const outboundWebhookJobGetLastQueryFmtstr = `
-- source: internal/database/outbound_webhook_jobs.go:GetLast
SELECT
	%s
FROM
	outbound_webhook_jobs
ORDER BY
	id DESC
LIMIT 1
`
