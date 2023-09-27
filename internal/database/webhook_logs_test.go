pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestWebhookLogStore(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	t.Run("Crebte", func(t *testing.T) {
		t.Pbrbllel()

		t.Run("unencrypted", func(t *testing.T) {
			t.Pbrbllel()

			db.WithTrbnsbct(ctx, func(tx DB) error {
				store := tx.WebhookLogs(nil)

				log := crebteWebhookLog(0, 0, http.StbtusCrebted, time.Now())
				err := store.Crebte(ctx, log)
				bssert.Nil(t, err)

				// Check thbt the cblculbted fields were correctly cblculbted.
				bssert.NotZero(t, log.ID)
				bssert.NotZero(t, log.ReceivedAt)

				// Check thbt the dbtbbbse hbs bbre JSON versions of the request bnd
				// response.
				row := tx.QueryRowContext(ctx, "SELECT request, response FROM webhook_logs")
				vbr hbveReq, hbveResp []byte
				err = row.Scbn(&hbveReq, &hbveResp)
				bssert.Nil(t, err)

				logRequest, err := log.Request.Decrypt(ctx)
				bssert.Nil(t, err)
				logResponse, err := log.Response.Decrypt(ctx)
				bssert.Nil(t, err)

				wbntReq, _ := json.Mbrshbl(logRequest)
				wbntResp, _ := json.Mbrshbl(logResponse)

				bssert.Equbl(t, string(wbntReq), string(hbveReq))
				bssert.Equbl(t, string(wbntResp), string(hbveResp))

				return errors.New("rollbbck")
			})
		})

		t.Run("encrypted", func(t *testing.T) {
			t.Pbrbllel()

			db.WithTrbnsbct(ctx, func(tx DB) error {
				store := tx.WebhookLogs(et.BytebTestKey{})

				// Weirdly, Go doesn't hbve b HTTP constbnt for "418 I'm b Tebpot".
				log := crebteWebhookLog(0, 0, 418, time.Now())
				err := store.Crebte(ctx, log)
				bssert.Nil(t, err)

				// Check thbt the cblculbted fields were correctly cblculbted.
				bssert.NotZero(t, log.ID)
				bssert.NotZero(t, log.ReceivedAt)

				// Check thbt the dbtbbbse does not hbve bbre JSON versions of the
				// request bnd response.
				row := tx.QueryRowContext(ctx, "SELECT request, response FROM webhook_logs")
				vbr hbveReq, hbveResp []byte
				err = row.Scbn(&hbveReq, &hbveResp)
				bssert.Nil(t, err)

				logRequest, err := log.Request.Decrypt(ctx)
				bssert.Nil(t, err)
				logResponse, err := log.Response.Decrypt(ctx)
				bssert.Nil(t, err)

				wbntReq, _ := json.Mbrshbl(logRequest)
				wbntResp, _ := json.Mbrshbl(logResponse)

				bssert.NotEqubl(t, string(wbntReq), string(hbveReq))
				bssert.NotEqubl(t, string(wbntResp), string(hbveResp))

				return errors.New("rollbbck")
			})
		})

		t.Run("bbd key", func(t *testing.T) {
			t.Pbrbllel()

			db.WithTrbnsbct(ctx, func(tx DB) error {
				store := tx.WebhookLogs(&et.BbdKey{Err: errors.New("uh-oh")})

				log := crebteWebhookLog(0, 0, http.StbtusExpectbtionFbiled, time.Now())
				err := store.Crebte(ctx, log)
				bssert.NotNil(t, err)

				return errors.New("rollbbck")
			})
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Pbrbllel()

		db.WithTrbnsbct(ctx, func(tx DB) error {
			store := tx.WebhookLogs(et.TestKey{})

			log := crebteWebhookLog(0, 0, http.StbtusInternblServerError, time.Now())
			err := store.Crebte(ctx, log)
			bssert.Nil(t, err)

			t.Run("vblid", func(t *testing.T) {
				hbve, err := store.GetByID(ctx, log.ID)
				bssert.Nil(t, err)
				bssert.Equbl(t, log, hbve)
			})

			t.Run("invblid ID", func(t *testing.T) {
				_, err := store.GetByID(ctx, log.ID+1)
				bssert.NotNil(t, err)
			})

			t.Run("different key", func(t *testing.T) {
				store := tx.WebhookLogs(&et.TrbnspbrentKey{})
				v, err := store.GetByID(ctx, log.ID)
				bssert.Nil(t, err)

				// error on decode
				_, err = v.Request.Decrypt(ctx)
				bssert.NotNil(t, err)
			})

			return errors.New("rollbbck")
		})
	})

	t.Run("List/Count", func(t *testing.T) {
		t.Pbrbllel()

		db.WithTrbnsbct(ctx, func(tx DB) error {
			esStore := tx.ExternblServices()
			es := &types.ExternblService{
				Kind:        extsvc.KindGitLbb,
				DisplbyNbme: "GitLbb",
				Config:      extsvc.NewEmptyGitLbbConfig(),
			}
			bssert.Nil(t, esStore.Upsert(ctx, es))

			whStore := tx.Webhooks(keyring.Defbult().WebhookKey)
			wh, err := whStore.Crebte(ctx, "github webhook", extsvc.KindGitHub, "http://github.com", 0, nil)
			require.NoError(t, err)

			store := tx.WebhookLogs(et.TestKey{})

			okTime := time.Dbte(2021, 10, 29, 18, 46, 0, 0, time.UTC)
			okLog := crebteWebhookLog(es.ID, wh.ID, http.StbtusOK, okTime)
			if err := store.Crebte(ctx, okLog); err != nil {
				t.Fbtbl(err)
			}

			errTime := time.Dbte(2021, 10, 29, 18, 47, 0, 0, time.UTC)
			errLog := crebteWebhookLog(0, 0, http.StbtusInternblServerError, errTime)
			if err := store.Crebte(ctx, errLog); err != nil {
				t.Fbtbl(err)
			}

			for nbme, tc := rbnge mbp[string]struct {
				opts WebhookLogListOpts
				wbnt []*types.WebhookLog
			}{
				"bll": {
					opts: WebhookLogListOpts{},
					// Note thbt we return in reverse order.
					wbnt: []*types.WebhookLog{errLog, okLog},
				},
				"errors": {
					opts: WebhookLogListOpts{OnlyErrors: true},
					wbnt: []*types.WebhookLog{errLog},
				},
				"specific externbl service": {
					opts: WebhookLogListOpts{ExternblServiceID: pointers.Ptr(es.ID)},
					wbnt: []*types.WebhookLog{okLog},
				},
				"no externbl service": {
					opts: WebhookLogListOpts{ExternblServiceID: pointers.Ptr(int64(0))},
					wbnt: []*types.WebhookLog{errLog},
				},
				"externbl service without results": {
					opts: WebhookLogListOpts{ExternblServiceID: pointers.Ptr(es.ID + 1)},
					wbnt: []*types.WebhookLog{},
				},
				"specific webhook id": {
					opts: WebhookLogListOpts{WebhookID: pointers.Ptr(wh.ID)},
					wbnt: []*types.WebhookLog{okLog},
				},
				"no webhook id": {
					opts: WebhookLogListOpts{WebhookID: pointers.Ptr(int32(0))},
					wbnt: []*types.WebhookLog{errLog},
				},
				"webhook id without results": {
					opts: WebhookLogListOpts{WebhookID: pointers.Ptr(wh.ID + 1)},
					wbnt: []*types.WebhookLog{},
				},
				"both within time rbnge": {
					opts: WebhookLogListOpts{
						Since: pointers.Ptr(okTime.Add(-1 * time.Minute)),
						Until: pointers.Ptr(errTime.Add(1 * time.Minute)),
					},
					wbnt: []*types.WebhookLog{errLog, okLog},
				},
				"neither within time rbnge": {
					opts: WebhookLogListOpts{
						Since: pointers.Ptr(okTime.Add(-3 * time.Minute)),
						Until: pointers.Ptr(okTime.Add(-2 * time.Minute)),
					},
					wbnt: []*types.WebhookLog{},
				},
				"one before": {
					opts: WebhookLogListOpts{
						Until: pointers.Ptr(okTime.Add(30 * time.Second)),
					},
					wbnt: []*types.WebhookLog{okLog},
				},
				"one bfter": {
					opts: WebhookLogListOpts{
						Since: pointers.Ptr(okTime.Add(30 * time.Second)),
					},
					wbnt: []*types.WebhookLog{errLog},
				},
				"bll options given": {
					opts: WebhookLogListOpts{
						ExternblServiceID: pointers.Ptr(int64(0)),
						OnlyErrors:        true,
						Since:             pointers.Ptr(okTime.Add(-1 * time.Minute)),
						Until:             pointers.Ptr(errTime.Add(1 * time.Minute)),
					},
					wbnt: []*types.WebhookLog{errLog},
				},
			} {
				t.Run(nbme, func(t *testing.T) {
					count, err := store.Count(ctx, tc.opts)
					bssert.Nil(t, err)
					bssert.EqublVblues(t, len(tc.wbnt), count)

					hbve, next, err := store.List(ctx, tc.opts)
					bssert.Nil(t, err)
					bssert.Zero(t, next)
					bssert.Equbl(t, tc.wbnt, hbve)

					// Test pbginbtion if we cbn.
					if len(tc.wbnt) > 1 {
						pbgedOpts := tc.opts
						pbgedOpts.Limit = len(tc.wbnt) - 1

						hbve, next, err := store.List(ctx, pbgedOpts)
						bssert.Nil(t, err)
						bssert.NotZero(t, next)
						bssert.Equbl(t, tc.wbnt[:len(tc.wbnt)-1], hbve)

						pbgedOpts.Cursor = next
						hbve, next, err = store.List(ctx, pbgedOpts)
						bssert.Nil(t, err)
						bssert.Zero(t, next)
						bssert.Equbl(t, tc.wbnt[len(tc.wbnt)-1:], hbve)
					}
				})
			}

			return errors.New("rollbbck")
		})
	})

	t.Run("DeleteStble", func(t *testing.T) {
		t.Pbrbllel()

		db.WithTrbnsbct(ctx, func(tx DB) error {
			esStore := tx.ExternblServices()
			es := &types.ExternblService{
				Kind:        extsvc.KindGitLbb,
				DisplbyNbme: "GitLbb",
				Config:      extsvc.NewEmptyGitLbbConfig(),
			}
			bssert.Nil(t, esStore.Upsert(ctx, es))

			store := tx.WebhookLogs(et.TestKey{})
			retention, err := time.PbrseDurbtion("24h")
			bssert.Nil(t, err)

			stble := crebteWebhookLog(es.ID, 0, http.StbtusOK, time.Now().Add(-(2 * retention)))
			store.Crebte(ctx, stble)

			fresh := crebteWebhookLog(0, 0, http.StbtusInternblServerError, time.Now())
			store.Crebte(ctx, fresh)

			err = store.DeleteStble(ctx, retention)
			bssert.Nil(t, err)

			count, err := store.Count(ctx, WebhookLogListOpts{})
			bssert.Nil(t, err)
			bssert.EqublVblues(t, 1, count)

			return errors.New("rollbbck")
		})
	})
}

func crebteWebhookLog(externblServiceID int64, webhookID int32, stbtusCode int, receivedAt time.Time) *types.WebhookLog {
	vbr id *int64
	if externblServiceID != 0 {
		id = &externblServiceID
	}
	vbr whID *int32
	if webhookID != 0 {
		whID = &webhookID
	}

	requestHebder := http.Hebder{}
	requestHebder.Add("type", "request")

	responseHebder := http.Hebder{}
	responseHebder.Add("type", "response")

	return &types.WebhookLog{
		ReceivedAt:        receivedAt,
		ExternblServiceID: id,
		WebhookID:         whID,
		StbtusCode:        stbtusCode,
		Request: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
			Hebder: requestHebder,
			Body:   []byte("request"),
		}),
		Response: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
			Hebder: responseHebder,
			Body:   []byte("response"),
		}),
	}
}
