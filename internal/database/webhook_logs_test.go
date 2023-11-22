package database

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestWebhookLogStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		t.Run("unencrypted", func(t *testing.T) {
			t.Parallel()

			db.WithTransact(ctx, func(tx DB) error {
				store := tx.WebhookLogs(nil)

				log := createWebhookLog(0, 0, http.StatusCreated, time.Now())
				err := store.Create(ctx, log)
				assert.Nil(t, err)

				// Check that the calculated fields were correctly calculated.
				assert.NotZero(t, log.ID)
				assert.NotZero(t, log.ReceivedAt)

				// Check that the database has bare JSON versions of the request and
				// response.
				row := tx.QueryRowContext(ctx, "SELECT request, response FROM webhook_logs")
				var haveReq, haveResp []byte
				err = row.Scan(&haveReq, &haveResp)
				assert.Nil(t, err)

				logRequest, err := log.Request.Decrypt(ctx)
				assert.Nil(t, err)
				logResponse, err := log.Response.Decrypt(ctx)
				assert.Nil(t, err)

				wantReq, _ := json.Marshal(logRequest)
				wantResp, _ := json.Marshal(logResponse)

				assert.Equal(t, string(wantReq), string(haveReq))
				assert.Equal(t, string(wantResp), string(haveResp))

				return errors.New("rollback")
			})
		})

		t.Run("encrypted", func(t *testing.T) {
			t.Parallel()

			db.WithTransact(ctx, func(tx DB) error {
				store := tx.WebhookLogs(et.ByteaTestKey{})

				// Weirdly, Go doesn't have a HTTP constant for "418 I'm a Teapot".
				log := createWebhookLog(0, 0, 418, time.Now())
				err := store.Create(ctx, log)
				assert.Nil(t, err)

				// Check that the calculated fields were correctly calculated.
				assert.NotZero(t, log.ID)
				assert.NotZero(t, log.ReceivedAt)

				// Check that the database does not have bare JSON versions of the
				// request and response.
				row := tx.QueryRowContext(ctx, "SELECT request, response FROM webhook_logs")
				var haveReq, haveResp []byte
				err = row.Scan(&haveReq, &haveResp)
				assert.Nil(t, err)

				logRequest, err := log.Request.Decrypt(ctx)
				assert.Nil(t, err)
				logResponse, err := log.Response.Decrypt(ctx)
				assert.Nil(t, err)

				wantReq, _ := json.Marshal(logRequest)
				wantResp, _ := json.Marshal(logResponse)

				assert.NotEqual(t, string(wantReq), string(haveReq))
				assert.NotEqual(t, string(wantResp), string(haveResp))

				return errors.New("rollback")
			})
		})

		t.Run("bad key", func(t *testing.T) {
			t.Parallel()

			db.WithTransact(ctx, func(tx DB) error {
				store := tx.WebhookLogs(&et.BadKey{Err: errors.New("uh-oh")})

				log := createWebhookLog(0, 0, http.StatusExpectationFailed, time.Now())
				err := store.Create(ctx, log)
				assert.NotNil(t, err)

				return errors.New("rollback")
			})
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Parallel()

		db.WithTransact(ctx, func(tx DB) error {
			store := tx.WebhookLogs(et.TestKey{})

			log := createWebhookLog(0, 0, http.StatusInternalServerError, time.Now())
			err := store.Create(ctx, log)
			assert.Nil(t, err)

			t.Run("valid", func(t *testing.T) {
				have, err := store.GetByID(ctx, log.ID)
				assert.Nil(t, err)
				assert.Equal(t, log, have)
			})

			t.Run("invalid ID", func(t *testing.T) {
				_, err := store.GetByID(ctx, log.ID+1)
				assert.NotNil(t, err)
			})

			t.Run("different key", func(t *testing.T) {
				store := tx.WebhookLogs(&et.TransparentKey{})
				v, err := store.GetByID(ctx, log.ID)
				assert.Nil(t, err)

				// error on decode
				_, err = v.Request.Decrypt(ctx)
				assert.NotNil(t, err)
			})

			return errors.New("rollback")
		})
	})

	t.Run("List/Count", func(t *testing.T) {
		t.Parallel()

		db.WithTransact(ctx, func(tx DB) error {
			esStore := tx.ExternalServices()
			es := &types.ExternalService{
				Kind:        extsvc.KindGitLab,
				DisplayName: "GitLab",
				Config:      extsvc.NewEmptyGitLabConfig(),
			}
			assert.Nil(t, esStore.Upsert(ctx, es))

			whStore := tx.Webhooks(keyring.Default().WebhookKey)
			wh, err := whStore.Create(ctx, "github webhook", extsvc.KindGitHub, "http://github.com", 0, nil)
			require.NoError(t, err)

			store := tx.WebhookLogs(et.TestKey{})

			okTime := time.Date(2021, 10, 29, 18, 46, 0, 0, time.UTC)
			okLog := createWebhookLog(es.ID, wh.ID, http.StatusOK, okTime)
			if err := store.Create(ctx, okLog); err != nil {
				t.Fatal(err)
			}

			errTime := time.Date(2021, 10, 29, 18, 47, 0, 0, time.UTC)
			errLog := createWebhookLog(0, 0, http.StatusInternalServerError, errTime)
			if err := store.Create(ctx, errLog); err != nil {
				t.Fatal(err)
			}

			for name, tc := range map[string]struct {
				opts WebhookLogListOpts
				want []*types.WebhookLog
			}{
				"all": {
					opts: WebhookLogListOpts{},
					// Note that we return in reverse order.
					want: []*types.WebhookLog{errLog, okLog},
				},
				"errors": {
					opts: WebhookLogListOpts{OnlyErrors: true},
					want: []*types.WebhookLog{errLog},
				},
				"specific external service": {
					opts: WebhookLogListOpts{ExternalServiceID: pointers.Ptr(es.ID)},
					want: []*types.WebhookLog{okLog},
				},
				"no external service": {
					opts: WebhookLogListOpts{ExternalServiceID: pointers.Ptr(int64(0))},
					want: []*types.WebhookLog{errLog},
				},
				"external service without results": {
					opts: WebhookLogListOpts{ExternalServiceID: pointers.Ptr(es.ID + 1)},
					want: []*types.WebhookLog{},
				},
				"specific webhook id": {
					opts: WebhookLogListOpts{WebhookID: pointers.Ptr(wh.ID)},
					want: []*types.WebhookLog{okLog},
				},
				"no webhook id": {
					opts: WebhookLogListOpts{WebhookID: pointers.Ptr(int32(0))},
					want: []*types.WebhookLog{errLog},
				},
				"webhook id without results": {
					opts: WebhookLogListOpts{WebhookID: pointers.Ptr(wh.ID + 1)},
					want: []*types.WebhookLog{},
				},
				"both within time range": {
					opts: WebhookLogListOpts{
						Since: pointers.Ptr(okTime.Add(-1 * time.Minute)),
						Until: pointers.Ptr(errTime.Add(1 * time.Minute)),
					},
					want: []*types.WebhookLog{errLog, okLog},
				},
				"neither within time range": {
					opts: WebhookLogListOpts{
						Since: pointers.Ptr(okTime.Add(-3 * time.Minute)),
						Until: pointers.Ptr(okTime.Add(-2 * time.Minute)),
					},
					want: []*types.WebhookLog{},
				},
				"one before": {
					opts: WebhookLogListOpts{
						Until: pointers.Ptr(okTime.Add(30 * time.Second)),
					},
					want: []*types.WebhookLog{okLog},
				},
				"one after": {
					opts: WebhookLogListOpts{
						Since: pointers.Ptr(okTime.Add(30 * time.Second)),
					},
					want: []*types.WebhookLog{errLog},
				},
				"all options given": {
					opts: WebhookLogListOpts{
						ExternalServiceID: pointers.Ptr(int64(0)),
						OnlyErrors:        true,
						Since:             pointers.Ptr(okTime.Add(-1 * time.Minute)),
						Until:             pointers.Ptr(errTime.Add(1 * time.Minute)),
					},
					want: []*types.WebhookLog{errLog},
				},
			} {
				t.Run(name, func(t *testing.T) {
					count, err := store.Count(ctx, tc.opts)
					assert.Nil(t, err)
					assert.EqualValues(t, len(tc.want), count)

					have, next, err := store.List(ctx, tc.opts)
					assert.Nil(t, err)
					assert.Zero(t, next)
					assert.Equal(t, tc.want, have)

					// Test pagination if we can.
					if len(tc.want) > 1 {
						pagedOpts := tc.opts
						pagedOpts.Limit = len(tc.want) - 1

						have, next, err := store.List(ctx, pagedOpts)
						assert.Nil(t, err)
						assert.NotZero(t, next)
						assert.Equal(t, tc.want[:len(tc.want)-1], have)

						pagedOpts.Cursor = next
						have, next, err = store.List(ctx, pagedOpts)
						assert.Nil(t, err)
						assert.Zero(t, next)
						assert.Equal(t, tc.want[len(tc.want)-1:], have)
					}
				})
			}

			return errors.New("rollback")
		})
	})

	t.Run("DeleteStale", func(t *testing.T) {
		t.Parallel()

		db.WithTransact(ctx, func(tx DB) error {
			esStore := tx.ExternalServices()
			es := &types.ExternalService{
				Kind:        extsvc.KindGitLab,
				DisplayName: "GitLab",
				Config:      extsvc.NewEmptyGitLabConfig(),
			}
			assert.Nil(t, esStore.Upsert(ctx, es))

			store := tx.WebhookLogs(et.TestKey{})
			retention, err := time.ParseDuration("24h")
			assert.Nil(t, err)

			stale := createWebhookLog(es.ID, 0, http.StatusOK, time.Now().Add(-(2 * retention)))
			store.Create(ctx, stale)

			fresh := createWebhookLog(0, 0, http.StatusInternalServerError, time.Now())
			store.Create(ctx, fresh)

			err = store.DeleteStale(ctx, retention)
			assert.Nil(t, err)

			count, err := store.Count(ctx, WebhookLogListOpts{})
			assert.Nil(t, err)
			assert.EqualValues(t, 1, count)

			return errors.New("rollback")
		})
	})
}

func createWebhookLog(externalServiceID int64, webhookID int32, statusCode int, receivedAt time.Time) *types.WebhookLog {
	var id *int64
	if externalServiceID != 0 {
		id = &externalServiceID
	}
	var whID *int32
	if webhookID != 0 {
		whID = &webhookID
	}

	requestHeader := http.Header{}
	requestHeader.Add("type", "request")

	responseHeader := http.Header{}
	responseHeader.Add("type", "response")

	return &types.WebhookLog{
		ReceivedAt:        receivedAt,
		ExternalServiceID: id,
		WebhookID:         whID,
		StatusCode:        statusCode,
		Request: types.NewUnencryptedWebhookLogMessage(types.WebhookLogMessage{
			Header: requestHeader,
			Body:   []byte("request"),
		}),
		Response: types.NewUnencryptedWebhookLogMessage(types.WebhookLogMessage{
			Header: responseHeader,
			Body:   []byte("response"),
		}),
	}
}
