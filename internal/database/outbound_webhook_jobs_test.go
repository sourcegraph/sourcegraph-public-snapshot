package database

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestOutboundWebhookJobs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	runBothEncryptionStates(t, func(t *testing.T, logger log.Logger, db DB, key encryption.Key) {
		store := db.OutboundWebhookJobs(key)

		var (
			scopedJob   *types.OutboundWebhookJob
			unscopedJob *types.OutboundWebhookJob
		)
		payload := []byte(`"TEST"`)

		t.Run("Create", func(t *testing.T) {

			t.Run("bad key", func(t *testing.T) {
				want := errors.New("bad key")
				key := &et.BadKey{Err: want}

				_, have := OutboundWebhookJobsWith(store, key).Create(ctx, "foo", nil, payload)
				assert.ErrorIs(t, have, want)
			})

			t.Run("success", func(t *testing.T) {
				for name, tc := range map[string]struct {
					scope  *string
					target **types.OutboundWebhookJob
				}{
					"scoped": {
						scope:  pointers.Ptr("scope"),
						target: &scopedJob,
					},
					"unscoped": {
						scope:  nil,
						target: &unscopedJob,
					},
				} {
					t.Run(name, func(t *testing.T) {
						job, err := store.Create(ctx, "foo", tc.scope, payload)
						assert.NoError(t, err)
						assert.NotNil(t, job)
						assert.Equal(t, job.EventType, "foo")
						assert.Equal(t, tc.scope, job.Scope)

						assertOutboundWebhookJobFieldsEncrypted(t, ctx, store, job, payload)

						*tc.target = job
					})
				}
			})
		})

		t.Run("GetByID", func(t *testing.T) {
			t.Run("not found", func(t *testing.T) {
				job, err := store.GetByID(ctx, 0)
				assert.True(t, errcode.IsNotFound(err))
				assert.Nil(t, job)
			})

			t.Run("found", func(t *testing.T) {
				job, err := store.GetByID(ctx, scopedJob.ID)
				assert.NoError(t, err)
				assertEqualOutboundWebhookJobs(t, ctx, scopedJob, job)
			})
		})

		t.Run("DeleteBefore", func(t *testing.T) {
			t.Run("nothing to delete due to no records before the time", func(t *testing.T) {
				err := store.DeleteBefore(ctx, time.Time{})
				assert.NoError(t, err)
				assert.Len(t, listOutboundWebhookJobs(t, ctx, store), 2)
			})

			before := scopedJob.QueuedAt.Add(time.Hour)

			t.Run("nothing to delete due to unfinished jobs", func(t *testing.T) {
				err := store.DeleteBefore(ctx, before)
				assert.NoError(t, err)
				assert.Len(t, listOutboundWebhookJobs(t, ctx, store), 2)
			})

			store.Handle().ExecContext(ctx, "UPDATE outbound_webhook_jobs SET finished_at = queued_at")

			t.Run("everything to delete", func(t *testing.T) {
				err := store.DeleteBefore(ctx, before)
				assert.NoError(t, err)
				assert.Len(t, listOutboundWebhookJobs(t, ctx, store), 0)
			})
		})
	})
}

func assertEqualOutboundWebhookJobs(t *testing.T, ctx context.Context, want, have *types.OutboundWebhookJob) {
	t.Helper()

	valueOf := func(e *encryption.Encryptable) string {
		t.Helper()
		return decryptedValue(t, ctx, e)
	}

	assert.Equal(t, want.ID, have.ID)
	assert.Equal(t, want.EventType, have.EventType)
	assert.Equal(t, want.Scope, have.Scope)
	assert.Equal(t, want.State, have.State)
	assert.Equal(t, want.FailureMessage, have.FailureMessage)
	assert.Equal(t, want.QueuedAt, have.QueuedAt)
	assert.Equal(t, want.StartedAt, have.StartedAt)
	assert.Equal(t, want.FinishedAt, have.FinishedAt)
	assert.Equal(t, want.ProcessAfter, have.ProcessAfter)
	assert.Equal(t, want.NumResets, have.NumResets)
	assert.Equal(t, want.NumFailures, have.NumFailures)
	assert.Equal(t, want.LastHeartbeatAt, have.LastHeartbeatAt)
	assert.Equal(t, want.ExecutionLogs, have.ExecutionLogs)
	assert.Equal(t, want.WorkerHostname, have.WorkerHostname)
	assert.Equal(t, want.Cancel, have.Cancel)
	assert.Equal(t, valueOf(want.Payload), valueOf(have.Payload))
}

func assertOutboundWebhookJobFieldsEncrypted(t *testing.T, ctx context.Context, store basestore.ShareableStore, job *types.OutboundWebhookJob, payload []byte) {
	t.Helper()

	if store.(*outboundWebhookJobStore).key == nil {
		return
	}

	decryptPayload, err := job.Payload.Decrypt(ctx)
	require.NoError(t, err)
	assert.Equal(t, payload, []byte(decryptPayload))

	row := store.Handle().QueryRowContext(
		ctx,
		"SELECT payload FROM outbound_webhook_jobs WHERE id = $1",
		job.ID,
	)
	var dbPayload string
	err = row.Scan(&dbPayload)
	assert.NoError(t, err)
	assert.NotEqual(t, dbPayload, decryptPayload)
}

func listOutboundWebhookJobs(t *testing.T, ctx context.Context, store OutboundWebhookJobStore) []*types.OutboundWebhookJob {
	t.Helper()

	s := store.(*outboundWebhookJobStore)

	rows, err := store.Query(ctx, sqlf.Sprintf(
		"SELECT %s FROM outbound_webhook_jobs ORDER BY id",
		sqlf.Join(OutboundWebhookJobColumns, ","),
	))
	require.NoError(t, err)
	defer rows.Close()

	jobs := []*types.OutboundWebhookJob{}
	for rows.Next() {
		var job types.OutboundWebhookJob
		require.NoError(t, s.scanOutboundWebhookJob(&job, rows))
		jobs = append(jobs, &job)
	}

	return jobs
}
