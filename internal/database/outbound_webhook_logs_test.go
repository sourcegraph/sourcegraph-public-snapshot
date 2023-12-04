package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestOutboundWebhookLogs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	runBothEncryptionStates(t, func(t *testing.T, logger log.Logger, db DB, key encryption.Key) {
		_, webhook := setupOutboundWebhookTest(t, ctx, db, key)

		payload := []byte(`"TEST"`)
		job, err := db.OutboundWebhookJobs(key).Create(ctx, "foo", nil, payload)
		require.NoError(t, err)

		store := db.OutboundWebhookLogs(key)

		var (
			successLog      *types.OutboundWebhookLog
			serverErrorLog  *types.OutboundWebhookLog
			networkErrorLog *types.OutboundWebhookLog
		)

		t.Run("Create", func(t *testing.T) {
			t.Run("bad key", func(t *testing.T) {
				want := errors.New("bad key")
				key := &et.BadKey{Err: want}

				owLog := newOutboundWebhookLogSuccess(t, job, webhook, 200, "req", "resp")

				have := OutboundWebhookLogsWith(store, key).Create(ctx, owLog)
				assert.ErrorIs(t, have, want)
			})

			t.Run("success", func(t *testing.T) {
				for _, tc := range []struct {
					name   string
					input  *types.OutboundWebhookLog
					target **types.OutboundWebhookLog
				}{
					{
						name:   "success",
						input:  newOutboundWebhookLogSuccess(t, job, webhook, 200, "req", "resp"),
						target: &successLog,
					},
					{
						name:   "server error",
						input:  newOutboundWebhookLogSuccess(t, job, webhook, 500, "req", "resp"),
						target: &serverErrorLog,
					},
					{
						name:   "network error",
						input:  newOutboundWebhookLogNetworkError(t, job, webhook, "pipes are bad"),
						target: &networkErrorLog,
					},
				} {
					t.Run(tc.name, func(t *testing.T) {
						err := store.Create(ctx, tc.input)
						assert.NoError(t, err)
						assertOutboundWebhookLogFieldsEncrypted(t, ctx, store, tc.input)
						*tc.target = tc.input
					})
				}
			})
		})

		t.Run("CountsForOutboundWebhook", func(t *testing.T) {
			t.Run("missing ID", func(t *testing.T) {
				total, errored, err := store.CountsForOutboundWebhook(ctx, 0)
				// This won't actually error due to the nature of the query.
				assert.NoError(t, err)
				assert.Zero(t, total)
				assert.Zero(t, errored)
			})

			t.Run("valid ID", func(t *testing.T) {
				total, errored, err := store.CountsForOutboundWebhook(ctx, webhook.ID)
				assert.NoError(t, err)
				assert.EqualValues(t, 3, total)
				assert.EqualValues(t, 2, errored)
			})
		})

		t.Run("ListForOutboundWebhook", func(t *testing.T) {
			for name, tc := range map[string]struct {
				opts OutboundWebhookLogListOpts
				want []*types.OutboundWebhookLog
			}{
				"missing ID": {
					opts: OutboundWebhookLogListOpts{OutboundWebhookID: 0},
					want: []*types.OutboundWebhookLog{},
				},
				"all": {
					opts: OutboundWebhookLogListOpts{OutboundWebhookID: webhook.ID},
					want: []*types.OutboundWebhookLog{networkErrorLog, serverErrorLog, successLog},
				},
				"errors only": {
					opts: OutboundWebhookLogListOpts{
						OnlyErrors:        true,
						OutboundWebhookID: webhook.ID,
					},
					want: []*types.OutboundWebhookLog{networkErrorLog, serverErrorLog},
				},
				"first page": {
					opts: OutboundWebhookLogListOpts{
						OnlyErrors:        true,
						OutboundWebhookID: webhook.ID,
						LimitOffset:       &LimitOffset{Limit: 1},
					},
					want: []*types.OutboundWebhookLog{networkErrorLog},
				},
				"second page": {
					opts: OutboundWebhookLogListOpts{
						OnlyErrors:        true,
						OutboundWebhookID: webhook.ID,
						LimitOffset:       &LimitOffset{Limit: 1, Offset: 1},
					},
					want: []*types.OutboundWebhookLog{serverErrorLog},
				},
				"third page": {
					opts: OutboundWebhookLogListOpts{
						OnlyErrors:        true,
						OutboundWebhookID: webhook.ID,
						LimitOffset:       &LimitOffset{Limit: 1, Offset: 2},
					},
					want: []*types.OutboundWebhookLog{},
				},
			} {
				t.Run(name, func(t *testing.T) {
					have, err := store.ListForOutboundWebhook(ctx, tc.opts)
					assert.NoError(t, err)
					assert.Len(t, have, len(tc.want))
					for i := range have {
						assertEqualOutboundWebhookLogs(t, ctx, tc.want[i], have[i])
					}
				})
			}
		})
	})
}

func assertEqualOutboundWebhookLogs(t *testing.T, ctx context.Context, want, have *types.OutboundWebhookLog) {
	t.Helper()

	valueOf := func(e *encryption.Encryptable) string {
		t.Helper()
		return decryptedValue(t, ctx, e)
	}

	assert.Equal(t, want.ID, have.ID)
	assert.Equal(t, want.JobID, have.JobID)
	assert.Equal(t, want.OutboundWebhookID, have.OutboundWebhookID)
	assert.Equal(t, want.SentAt, have.SentAt)
	assert.Equal(t, want.StatusCode, have.StatusCode)
	assert.Equal(t, valueOf(want.Request.Encryptable), valueOf(have.Request.Encryptable))
	assert.Equal(t, valueOf(want.Response.Encryptable), valueOf(have.Response.Encryptable))
	assert.Equal(t, valueOf(want.Error), valueOf(have.Error))
}

func assertOutboundWebhookLogFieldsEncrypted(t *testing.T, ctx context.Context, store basestore.ShareableStore, log *types.OutboundWebhookLog) {
	t.Helper()

	if store.(*outboundWebhookLogStore).key == nil {
		return
	}

	request, err := log.Request.Encryptable.Decrypt(ctx)
	require.NoError(t, err)

	response, err := log.Response.Encryptable.Decrypt(ctx)
	require.NoError(t, err)

	errorMessage, err := log.Error.Decrypt(ctx)
	require.NoError(t, err)

	row := store.Handle().QueryRowContext(
		ctx,
		"SELECT request, response, error FROM outbound_webhook_logs WHERE id = $1",
		log.ID,
	)
	var dbRequest, dbResponse, dbErrorMessage string
	err = row.Scan(&dbRequest, &dbResponse, &dbErrorMessage)
	assert.NoError(t, err)
	assert.NotEqual(t, dbRequest, request)
	assert.NotEqual(t, dbResponse, response)
	assert.NotEqual(t, dbErrorMessage, errorMessage)
}

func newOutboundWebhookLog(
	t *testing.T, job *types.OutboundWebhookJob,
	webhook *types.OutboundWebhook, statusCode int,
	request, response types.WebhookLogMessage, err string,
) *types.OutboundWebhookLog {
	t.Helper()

	return &types.OutboundWebhookLog{
		JobID:             job.ID,
		OutboundWebhookID: webhook.ID,
		StatusCode:        statusCode,
		Request:           types.NewUnencryptedWebhookLogMessage(request),
		Response:          types.NewUnencryptedWebhookLogMessage(response),
		Error:             encryption.NewUnencrypted(err),
	}
}

func newOutboundWebhookLogNetworkError(
	t *testing.T, job *types.OutboundWebhookJob,
	webhook *types.OutboundWebhook, message string,
) *types.OutboundWebhookLog {
	t.Helper()

	return newOutboundWebhookLog(
		t, job, webhook, 0,
		newWebhookLogMessage(t, "request"),
		types.WebhookLogMessage{},
		message,
	)
}

func newOutboundWebhookLogSuccess(
	t *testing.T, job *types.OutboundWebhookJob,
	webhook *types.OutboundWebhook, statusCode int,
	request, response string,
) *types.OutboundWebhookLog {
	t.Helper()

	return newOutboundWebhookLog(
		t, job, webhook, statusCode,
		newWebhookLogMessage(t, request),
		newWebhookLogMessage(t, response),
		"",
	)
}

func newWebhookLogMessage(t *testing.T, body string) types.WebhookLogMessage {
	t.Helper()

	return types.WebhookLogMessage{
		Header: map[string][]string{
			"content-type": {"application/json; charset=utf-8"},
		},
		Body:   []byte(body),
		Method: "POST",
		URL:    "http://url/",
	}
}

func setupOutboundWebhookTest(t *testing.T, ctx context.Context, db DB, key encryption.Key) (user *types.User, webhook *types.OutboundWebhook) {
	t.Helper()

	userStore := db.Users()
	user, err := userStore.Create(ctx, NewUser{Username: "test"})
	require.NoError(t, err)

	webhookStore := db.OutboundWebhooks(key)
	webhook = newTestWebhook(t, user, ScopedEventType{EventType: "foo"})
	err = webhookStore.Create(ctx, webhook)
	require.NoError(t, err)

	return
}
