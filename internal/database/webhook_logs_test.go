package database

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	keytesting "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestWebhookLogStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := dbtest.NewDB(t)

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		t.Run("unencrypted", func(t *testing.T) {
			t.Parallel()

			tx, err := db.Begin()
			assert.Nil(t, err)
			defer tx.Rollback()
			store := WebhookLogs(tx, nil)

			log := createWebhookLog(0, http.StatusCreated, time.Now())
			err = store.Create(ctx, log)
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

			wantReq, _ := json.Marshal(&log.Request)
			wantResp, _ := json.Marshal(&log.Response)

			assert.Equal(t, string(wantReq), string(haveReq))
			assert.Equal(t, string(wantResp), string(haveResp))
		})

		t.Run("encrypted", func(t *testing.T) {
			t.Parallel()

			tx, err := db.Begin()
			assert.Nil(t, err)
			defer tx.Rollback()
			store := WebhookLogs(tx, keytesting.TestKey{})

			// Weirdly, Go doesn't have a HTTP constant for "418 I'm a Teapot".
			log := createWebhookLog(0, 418, time.Now())
			err = store.Create(ctx, log)
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

			wantReq, _ := json.Marshal(&log.Request)
			wantResp, _ := json.Marshal(&log.Response)

			assert.NotEqual(t, string(wantReq), string(haveReq))
			assert.NotEqual(t, string(wantResp), string(haveResp))
		})

		t.Run("bad key", func(t *testing.T) {
			t.Parallel()

			tx, err := db.Begin()
			assert.Nil(t, err)
			defer tx.Rollback()
			store := WebhookLogs(tx, &keytesting.BadKey{})

			log := createWebhookLog(0, http.StatusExpectationFailed, time.Now())
			err = store.Create(ctx, log)
			assert.NotNil(t, err)
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Parallel()

		tx, err := db.Begin()
		assert.Nil(t, err)
		defer tx.Rollback()
		store := WebhookLogs(tx, keytesting.TestKey{})

		log := createWebhookLog(0, http.StatusInternalServerError, time.Now())
		err = store.Create(ctx, log)
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
			store := WebhookLogs(tx, &keytesting.TransparentKey{})
			_, err := store.GetByID(ctx, log.ID)
			assert.NotNil(t, err)
		})
	})

	t.Run("List/Count", func(t *testing.T) {
		t.Parallel()

		tx, err := db.Begin()
		assert.Nil(t, err)
		defer tx.Rollback()

		esStore := NewDB(tx).ExternalServices()
		es := &types.ExternalService{
			Kind:        extsvc.KindGitLab,
			DisplayName: "GitLab",
			Config:      "{}",
		}
		assert.Nil(t, esStore.Upsert(ctx, es))

		store := WebhookLogs(tx, keytesting.TestKey{})

		okTime := time.Date(2021, 10, 29, 18, 46, 0, 0, time.UTC)
		okLog := createWebhookLog(es.ID, http.StatusOK, okTime)
		store.Create(ctx, okLog)

		errTime := time.Date(2021, 10, 29, 18, 47, 0, 0, time.UTC)
		errLog := createWebhookLog(0, http.StatusInternalServerError, errTime)
		store.Create(ctx, errLog)

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
				opts: WebhookLogListOpts{ExternalServiceID: int64Ptr(es.ID)},
				want: []*types.WebhookLog{okLog},
			},
			"no external service": {
				opts: WebhookLogListOpts{ExternalServiceID: int64Ptr(0)},
				want: []*types.WebhookLog{errLog},
			},
			"external service without results": {
				opts: WebhookLogListOpts{ExternalServiceID: int64Ptr(es.ID + 1)},
				want: []*types.WebhookLog{},
			},
			"both within time range": {
				opts: WebhookLogListOpts{
					Since: timePtr(okTime.Add(-1 * time.Minute)),
					Until: timePtr(errTime.Add(1 * time.Minute)),
				},
				want: []*types.WebhookLog{errLog, okLog},
			},
			"neither within time range": {
				opts: WebhookLogListOpts{
					Since: timePtr(okTime.Add(-3 * time.Minute)),
					Until: timePtr(okTime.Add(-2 * time.Minute)),
				},
				want: []*types.WebhookLog{},
			},
			"one before": {
				opts: WebhookLogListOpts{
					Until: timePtr(okTime.Add(30 * time.Second)),
				},
				want: []*types.WebhookLog{okLog},
			},
			"one after": {
				opts: WebhookLogListOpts{
					Since: timePtr(okTime.Add(30 * time.Second)),
				},
				want: []*types.WebhookLog{errLog},
			},
			"all options given": {
				opts: WebhookLogListOpts{
					ExternalServiceID: int64Ptr(0),
					OnlyErrors:        true,
					Since:             timePtr(okTime.Add(-1 * time.Minute)),
					Until:             timePtr(errTime.Add(1 * time.Minute)),
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
	})

	t.Run("DeleteStale", func(t *testing.T) {
		t.Parallel()

		tx, err := db.Begin()
		assert.Nil(t, err)
		defer tx.Rollback()

		esStore := NewDB(tx).ExternalServices()
		es := &types.ExternalService{
			Kind:        extsvc.KindGitLab,
			DisplayName: "GitLab",
			Config:      "{}",
		}
		assert.Nil(t, esStore.Upsert(ctx, es))

		store := WebhookLogs(tx, keytesting.TestKey{})
		retention, err := time.ParseDuration("24h")
		assert.Nil(t, err)

		stale := createWebhookLog(es.ID, http.StatusOK, time.Now().Add(-(2 * retention)))
		store.Create(ctx, stale)

		fresh := createWebhookLog(0, http.StatusInternalServerError, time.Now())
		store.Create(ctx, fresh)

		err = store.DeleteStale(ctx, retention)
		assert.Nil(t, err)

		count, err := store.Count(ctx, WebhookLogListOpts{})
		assert.Nil(t, err)
		assert.EqualValues(t, 1, count)
	})
}

func createWebhookLog(externalServiceID int64, statusCode int, receivedAt time.Time) *types.WebhookLog {
	var id *int64
	if externalServiceID != 0 {
		id = &externalServiceID
	}

	requestHeader := http.Header{}
	requestHeader.Add("type", "request")

	responseHeader := http.Header{}
	responseHeader.Add("type", "response")

	return &types.WebhookLog{
		ReceivedAt:        receivedAt,
		ExternalServiceID: id,
		StatusCode:        statusCode,
		Request: types.WebhookLogMessage{
			Header: requestHeader,
			Body:   []byte("request"),
		},
		Response: types.WebhookLogMessage{
			Header: responseHeader,
			Body:   []byte("response"),
		},
	}
}

func int64Ptr(v int64) *int64        { return &v }
func timePtr(v time.Time) *time.Time { return &v }
