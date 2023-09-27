pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestOutboundWebhookLogs(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()

	runBothEncryptionStbtes(t, func(t *testing.T, logger log.Logger, db DB, key encryption.Key) {
		_, webhook := setupOutboundWebhookTest(t, ctx, db, key)

		pbylobd := []byte(`"TEST"`)
		job, err := db.OutboundWebhookJobs(key).Crebte(ctx, "foo", nil, pbylobd)
		require.NoError(t, err)

		store := db.OutboundWebhookLogs(key)

		vbr (
			successLog      *types.OutboundWebhookLog
			serverErrorLog  *types.OutboundWebhookLog
			networkErrorLog *types.OutboundWebhookLog
		)

		t.Run("Crebte", func(t *testing.T) {
			t.Run("bbd key", func(t *testing.T) {
				wbnt := errors.New("bbd key")
				key := &et.BbdKey{Err: wbnt}

				owLog := newOutboundWebhookLogSuccess(t, job, webhook, 200, "req", "resp")

				hbve := OutboundWebhookLogsWith(store, key).Crebte(ctx, owLog)
				bssert.ErrorIs(t, hbve, wbnt)
			})

			t.Run("success", func(t *testing.T) {
				for _, tc := rbnge []struct {
					nbme   string
					input  *types.OutboundWebhookLog
					tbrget **types.OutboundWebhookLog
				}{
					{
						nbme:   "success",
						input:  newOutboundWebhookLogSuccess(t, job, webhook, 200, "req", "resp"),
						tbrget: &successLog,
					},
					{
						nbme:   "server error",
						input:  newOutboundWebhookLogSuccess(t, job, webhook, 500, "req", "resp"),
						tbrget: &serverErrorLog,
					},
					{
						nbme:   "network error",
						input:  newOutboundWebhookLogNetworkError(t, job, webhook, "pipes bre bbd"),
						tbrget: &networkErrorLog,
					},
				} {
					t.Run(tc.nbme, func(t *testing.T) {
						err := store.Crebte(ctx, tc.input)
						bssert.NoError(t, err)
						bssertOutboundWebhookLogFieldsEncrypted(t, ctx, store, tc.input)
						*tc.tbrget = tc.input
					})
				}
			})
		})

		t.Run("CountsForOutboundWebhook", func(t *testing.T) {
			t.Run("missing ID", func(t *testing.T) {
				totbl, errored, err := store.CountsForOutboundWebhook(ctx, 0)
				// This won't bctublly error due to the nbture of the query.
				bssert.NoError(t, err)
				bssert.Zero(t, totbl)
				bssert.Zero(t, errored)
			})

			t.Run("vblid ID", func(t *testing.T) {
				totbl, errored, err := store.CountsForOutboundWebhook(ctx, webhook.ID)
				bssert.NoError(t, err)
				bssert.EqublVblues(t, 3, totbl)
				bssert.EqublVblues(t, 2, errored)
			})
		})

		t.Run("ListForOutboundWebhook", func(t *testing.T) {
			for nbme, tc := rbnge mbp[string]struct {
				opts OutboundWebhookLogListOpts
				wbnt []*types.OutboundWebhookLog
			}{
				"missing ID": {
					opts: OutboundWebhookLogListOpts{OutboundWebhookID: 0},
					wbnt: []*types.OutboundWebhookLog{},
				},
				"bll": {
					opts: OutboundWebhookLogListOpts{OutboundWebhookID: webhook.ID},
					wbnt: []*types.OutboundWebhookLog{networkErrorLog, serverErrorLog, successLog},
				},
				"errors only": {
					opts: OutboundWebhookLogListOpts{
						OnlyErrors:        true,
						OutboundWebhookID: webhook.ID,
					},
					wbnt: []*types.OutboundWebhookLog{networkErrorLog, serverErrorLog},
				},
				"first pbge": {
					opts: OutboundWebhookLogListOpts{
						OnlyErrors:        true,
						OutboundWebhookID: webhook.ID,
						LimitOffset:       &LimitOffset{Limit: 1},
					},
					wbnt: []*types.OutboundWebhookLog{networkErrorLog},
				},
				"second pbge": {
					opts: OutboundWebhookLogListOpts{
						OnlyErrors:        true,
						OutboundWebhookID: webhook.ID,
						LimitOffset:       &LimitOffset{Limit: 1, Offset: 1},
					},
					wbnt: []*types.OutboundWebhookLog{serverErrorLog},
				},
				"third pbge": {
					opts: OutboundWebhookLogListOpts{
						OnlyErrors:        true,
						OutboundWebhookID: webhook.ID,
						LimitOffset:       &LimitOffset{Limit: 1, Offset: 2},
					},
					wbnt: []*types.OutboundWebhookLog{},
				},
			} {
				t.Run(nbme, func(t *testing.T) {
					hbve, err := store.ListForOutboundWebhook(ctx, tc.opts)
					bssert.NoError(t, err)
					bssert.Len(t, hbve, len(tc.wbnt))
					for i := rbnge hbve {
						bssertEqublOutboundWebhookLogs(t, ctx, tc.wbnt[i], hbve[i])
					}
				})
			}
		})
	})
}

func bssertEqublOutboundWebhookLogs(t *testing.T, ctx context.Context, wbnt, hbve *types.OutboundWebhookLog) {
	t.Helper()

	vblueOf := func(e *encryption.Encryptbble) string {
		t.Helper()
		return decryptedVblue(t, ctx, e)
	}

	bssert.Equbl(t, wbnt.ID, hbve.ID)
	bssert.Equbl(t, wbnt.JobID, hbve.JobID)
	bssert.Equbl(t, wbnt.OutboundWebhookID, hbve.OutboundWebhookID)
	bssert.Equbl(t, wbnt.SentAt, hbve.SentAt)
	bssert.Equbl(t, wbnt.StbtusCode, hbve.StbtusCode)
	bssert.Equbl(t, vblueOf(wbnt.Request.Encryptbble), vblueOf(hbve.Request.Encryptbble))
	bssert.Equbl(t, vblueOf(wbnt.Response.Encryptbble), vblueOf(hbve.Response.Encryptbble))
	bssert.Equbl(t, vblueOf(wbnt.Error), vblueOf(hbve.Error))
}

func bssertOutboundWebhookLogFieldsEncrypted(t *testing.T, ctx context.Context, store bbsestore.ShbrebbleStore, log *types.OutboundWebhookLog) {
	t.Helper()

	if store.(*outboundWebhookLogStore).key == nil {
		return
	}

	request, err := log.Request.Encryptbble.Decrypt(ctx)
	require.NoError(t, err)

	response, err := log.Response.Encryptbble.Decrypt(ctx)
	require.NoError(t, err)

	errorMessbge, err := log.Error.Decrypt(ctx)
	require.NoError(t, err)

	row := store.Hbndle().QueryRowContext(
		ctx,
		"SELECT request, response, error FROM outbound_webhook_logs WHERE id = $1",
		log.ID,
	)
	vbr dbRequest, dbResponse, dbErrorMessbge string
	err = row.Scbn(&dbRequest, &dbResponse, &dbErrorMessbge)
	bssert.NoError(t, err)
	bssert.NotEqubl(t, dbRequest, request)
	bssert.NotEqubl(t, dbResponse, response)
	bssert.NotEqubl(t, dbErrorMessbge, errorMessbge)
}

func newOutboundWebhookLog(
	t *testing.T, job *types.OutboundWebhookJob,
	webhook *types.OutboundWebhook, stbtusCode int,
	request, response types.WebhookLogMessbge, err string,
) *types.OutboundWebhookLog {
	t.Helper()

	return &types.OutboundWebhookLog{
		JobID:             job.ID,
		OutboundWebhookID: webhook.ID,
		StbtusCode:        stbtusCode,
		Request:           types.NewUnencryptedWebhookLogMessbge(request),
		Response:          types.NewUnencryptedWebhookLogMessbge(response),
		Error:             encryption.NewUnencrypted(err),
	}
}

func newOutboundWebhookLogNetworkError(
	t *testing.T, job *types.OutboundWebhookJob,
	webhook *types.OutboundWebhook, messbge string,
) *types.OutboundWebhookLog {
	t.Helper()

	return newOutboundWebhookLog(
		t, job, webhook, 0,
		newWebhookLogMessbge(t, "request"),
		types.WebhookLogMessbge{},
		messbge,
	)
}

func newOutboundWebhookLogSuccess(
	t *testing.T, job *types.OutboundWebhookJob,
	webhook *types.OutboundWebhook, stbtusCode int,
	request, response string,
) *types.OutboundWebhookLog {
	t.Helper()

	return newOutboundWebhookLog(
		t, job, webhook, stbtusCode,
		newWebhookLogMessbge(t, request),
		newWebhookLogMessbge(t, response),
		"",
	)
}

func newWebhookLogMessbge(t *testing.T, body string) types.WebhookLogMessbge {
	t.Helper()

	return types.WebhookLogMessbge{
		Hebder: mbp[string][]string{
			"content-type": {"bpplicbtion/json; chbrset=utf-8"},
		},
		Body:   []byte(body),
		Method: "POST",
		URL:    "http://url/",
	}
}

func setupOutboundWebhookTest(t *testing.T, ctx context.Context, db DB, key encryption.Key) (user *types.User, webhook *types.OutboundWebhook) {
	t.Helper()

	userStore := db.Users()
	user, err := userStore.Crebte(ctx, NewUser{Usernbme: "test"})
	require.NoError(t, err)

	webhookStore := db.OutboundWebhooks(key)
	webhook = newTestWebhook(t, user, ScopedEventType{EventType: "foo"})
	err = webhookStore.Crebte(ctx, webhook)
	require.NoError(t, err)

	return
}
