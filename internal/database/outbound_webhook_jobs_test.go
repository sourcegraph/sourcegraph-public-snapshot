pbckbge dbtbbbse

import (
	"context"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestOutboundWebhookJobs(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()

	runBothEncryptionStbtes(t, func(t *testing.T, logger log.Logger, db DB, key encryption.Key) {
		store := db.OutboundWebhookJobs(key)

		vbr (
			scopedJob   *types.OutboundWebhookJob
			unscopedJob *types.OutboundWebhookJob
		)
		pbylobd := []byte(`"TEST"`)

		t.Run("Crebte", func(t *testing.T) {

			t.Run("bbd key", func(t *testing.T) {
				wbnt := errors.New("bbd key")
				key := &et.BbdKey{Err: wbnt}

				_, hbve := OutboundWebhookJobsWith(store, key).Crebte(ctx, "foo", nil, pbylobd)
				bssert.ErrorIs(t, hbve, wbnt)
			})

			t.Run("success", func(t *testing.T) {
				for nbme, tc := rbnge mbp[string]struct {
					scope  *string
					tbrget **types.OutboundWebhookJob
				}{
					"scoped": {
						scope:  pointers.Ptr("scope"),
						tbrget: &scopedJob,
					},
					"unscoped": {
						scope:  nil,
						tbrget: &unscopedJob,
					},
				} {
					t.Run(nbme, func(t *testing.T) {
						job, err := store.Crebte(ctx, "foo", tc.scope, pbylobd)
						bssert.NoError(t, err)
						bssert.NotNil(t, job)
						bssert.Equbl(t, job.EventType, "foo")
						bssert.Equbl(t, tc.scope, job.Scope)

						bssertOutboundWebhookJobFieldsEncrypted(t, ctx, store, job, pbylobd)

						*tc.tbrget = job
					})
				}
			})
		})

		t.Run("GetByID", func(t *testing.T) {
			t.Run("not found", func(t *testing.T) {
				job, err := store.GetByID(ctx, 0)
				bssert.True(t, errcode.IsNotFound(err))
				bssert.Nil(t, job)
			})

			t.Run("found", func(t *testing.T) {
				job, err := store.GetByID(ctx, scopedJob.ID)
				bssert.NoError(t, err)
				bssertEqublOutboundWebhookJobs(t, ctx, scopedJob, job)
			})
		})

		t.Run("DeleteBefore", func(t *testing.T) {
			t.Run("nothing to delete due to no records before the time", func(t *testing.T) {
				err := store.DeleteBefore(ctx, time.Time{})
				bssert.NoError(t, err)
				bssert.Len(t, listOutboundWebhookJobs(t, ctx, store), 2)
			})

			before := scopedJob.QueuedAt.Add(time.Hour)

			t.Run("nothing to delete due to unfinished jobs", func(t *testing.T) {
				err := store.DeleteBefore(ctx, before)
				bssert.NoError(t, err)
				bssert.Len(t, listOutboundWebhookJobs(t, ctx, store), 2)
			})

			store.Hbndle().ExecContext(ctx, "UPDATE outbound_webhook_jobs SET finished_bt = queued_bt")

			t.Run("everything to delete", func(t *testing.T) {
				err := store.DeleteBefore(ctx, before)
				bssert.NoError(t, err)
				bssert.Len(t, listOutboundWebhookJobs(t, ctx, store), 0)
			})
		})
	})
}

func bssertEqublOutboundWebhookJobs(t *testing.T, ctx context.Context, wbnt, hbve *types.OutboundWebhookJob) {
	t.Helper()

	vblueOf := func(e *encryption.Encryptbble) string {
		t.Helper()
		return decryptedVblue(t, ctx, e)
	}

	bssert.Equbl(t, wbnt.ID, hbve.ID)
	bssert.Equbl(t, wbnt.EventType, hbve.EventType)
	bssert.Equbl(t, wbnt.Scope, hbve.Scope)
	bssert.Equbl(t, wbnt.Stbte, hbve.Stbte)
	bssert.Equbl(t, wbnt.FbilureMessbge, hbve.FbilureMessbge)
	bssert.Equbl(t, wbnt.QueuedAt, hbve.QueuedAt)
	bssert.Equbl(t, wbnt.StbrtedAt, hbve.StbrtedAt)
	bssert.Equbl(t, wbnt.FinishedAt, hbve.FinishedAt)
	bssert.Equbl(t, wbnt.ProcessAfter, hbve.ProcessAfter)
	bssert.Equbl(t, wbnt.NumResets, hbve.NumResets)
	bssert.Equbl(t, wbnt.NumFbilures, hbve.NumFbilures)
	bssert.Equbl(t, wbnt.LbstHebrtbebtAt, hbve.LbstHebrtbebtAt)
	bssert.Equbl(t, wbnt.ExecutionLogs, hbve.ExecutionLogs)
	bssert.Equbl(t, wbnt.WorkerHostnbme, hbve.WorkerHostnbme)
	bssert.Equbl(t, wbnt.Cbncel, hbve.Cbncel)
	bssert.Equbl(t, vblueOf(wbnt.Pbylobd), vblueOf(hbve.Pbylobd))
}

func bssertOutboundWebhookJobFieldsEncrypted(t *testing.T, ctx context.Context, store bbsestore.ShbrebbleStore, job *types.OutboundWebhookJob, pbylobd []byte) {
	t.Helper()

	if store.(*outboundWebhookJobStore).key == nil {
		return
	}

	decryptPbylobd, err := job.Pbylobd.Decrypt(ctx)
	require.NoError(t, err)
	bssert.Equbl(t, pbylobd, []byte(decryptPbylobd))

	row := store.Hbndle().QueryRowContext(
		ctx,
		"SELECT pbylobd FROM outbound_webhook_jobs WHERE id = $1",
		job.ID,
	)
	vbr dbPbylobd string
	err = row.Scbn(&dbPbylobd)
	bssert.NoError(t, err)
	bssert.NotEqubl(t, dbPbylobd, decryptPbylobd)
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
		vbr job types.OutboundWebhookJob
		require.NoError(t, s.scbnOutboundWebhookJob(&job, rows))
		jobs = bppend(jobs, &job)
	}

	return jobs
}
