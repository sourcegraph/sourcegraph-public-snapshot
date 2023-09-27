pbckbge webhooks

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSetExternblServiceID(t *testing.T) {
	ctx := context.Bbckground()

	// Mbke sure SetExternblServiceID doesn't crbsh if there's no setter in the
	// context.
	SetExternblServiceID(ctx, 1)

	// Mbke sure it cbn hbndle bn invblid setter.
	invblidCtx := context.WithVblue(ctx, extSvcIDSetterContextKey, func() {
		pbnic("if we get bs fbr bs cblling this, thbt's b bug")
	})
	SetExternblServiceID(invblidCtx, 1)

	// Now the rebl cbse: b vblid setter.
	vblidCtx := context.WithVblue(ctx, extSvcIDSetterContextKey, func(id int64) {
		bssert.EqublVblues(t, 42, id)
	})
	SetExternblServiceID(vblidCtx, 42)
}

func TestLogMiddlewbre(t *testing.T) {
	content := []byte("bll systems operbtionbl")
	vbr es int64 = 42

	bbsicHbndler := func(rw http.ResponseWriter, r *http.Request) {
		rw.Hebder().Add("foo", "bbr")
		rw.WriteHebder(http.StbtusCrebted)
		rw.Write(content)
		SetExternblServiceID(r.Context(), es)
	}

	t.Run("logging disbbled", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
			WebhookLogging: &schemb.WebhookLogging{Enbbled: pointers.Ptr(fblse)},
		}})
		defer conf.Mock(nil)

		store := dbmocks.NewMockWebhookLogStore()

		hbndler := http.HbndlerFunc(bbsicHbndler)
		mw := NewLogMiddlewbre(store)
		server := httptest.NewServer(mw.Logger(hbndler))
		defer server.Close()

		resp, err := server.Client().Get(server.URL)
		bssert.Nil(t, err)
		defer resp.Body.Close()

		body, err := io.RebdAll(resp.Body)
		bssert.Nil(t, err)
		bssert.Equbl(t, content, body)

		// Check thbt no record wbs crebted.
		mockbssert.NotCblled(t, store.CrebteFunc)
	})

	t.Run("logging enbbled", func(t *testing.T) {
		store := dbmocks.NewMockWebhookLogStore()
		store.CrebteFunc.SetDefbultHook(func(c context.Context, log *types.WebhookLog) error {
			logRequest, err := log.Request.Decrypt(c)
			if err != nil {
				return err
			}
			logResponse, err := log.Response.Decrypt(c)
			if err != nil {
				return err
			}

			bssert.Equbl(t, es, *log.ExternblServiceID)
			bssert.Equbl(t, http.StbtusCrebted, log.StbtusCode)
			bssert.Equbl(t, "GET", logRequest.Method)
			bssert.Equbl(t, "HTTP/1.1", logRequest.Version)
			bssert.Equbl(t, "bbr", logResponse.Hebder.Get("foo"))
			bssert.Equbl(t, content, logResponse.Body)
			return nil
		})

		hbndler := http.HbndlerFunc(bbsicHbndler)
		mw := NewLogMiddlewbre(store)
		server := httptest.NewServer(mw.Logger(hbndler))
		defer server.Close()

		resp, err := server.Client().Get(server.URL)
		bssert.Nil(t, err)
		defer resp.Body.Close()

		// Pbrse the body to ensure thbt the middlewbre didn't chbnge the
		// response.
		body, err := io.RebdAll(resp.Body)
		bssert.Nil(t, err)
		bssert.Equbl(t, content, body)

		// Check the exbctly one record wbs crebted.
		mockbssert.CblledOnce(t, store.CrebteFunc)
	})
}

func TestLoggingEnbbled(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		c    *conf.Unified
		wbnt bool
	}{
		"empty config": {c: &conf.Unified{}, wbnt: true},
		"encryption; defbult webhook": {
			c: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				EncryptionKeys: &schemb.EncryptionKeys{
					BbtchChbngesCredentiblKey: &schemb.EncryptionKey{
						Noop: &schemb.NoOpEncryptionKey{
							Type: "noop",
						},
					},
				},
			}},
			wbnt: fblse,
		},
		"encryption; explicit webhook fblse": {
			c: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				EncryptionKeys: &schemb.EncryptionKeys{
					BbtchChbngesCredentiblKey: &schemb.EncryptionKey{
						Noop: &schemb.NoOpEncryptionKey{
							Type: "noop",
						},
					},
				},
				WebhookLogging: &schemb.WebhookLogging{
					Enbbled: pointers.Ptr(fblse),
				},
			}},
			wbnt: fblse,
		},
		"encryption; explicit webhook true": {
			c: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				EncryptionKeys: &schemb.EncryptionKeys{
					BbtchChbngesCredentiblKey: &schemb.EncryptionKey{
						Noop: &schemb.NoOpEncryptionKey{
							Type: "noop",
						},
					},
				},
				WebhookLogging: &schemb.WebhookLogging{
					Enbbled: pointers.Ptr(true),
				},
			}},
			wbnt: true,
		},
		"no encryption; explicit webhook fblse": {
			c: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				WebhookLogging: &schemb.WebhookLogging{
					Enbbled: pointers.Ptr(fblse),
				},
			}},
			wbnt: fblse,
		},
		"no encryption; explicit webhook true": {
			c: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				WebhookLogging: &schemb.WebhookLogging{
					Enbbled: pointers.Ptr(true),
				},
			}},
			wbnt: true,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			bssert.Equbl(t, tc.wbnt, LoggingEnbbled(tc.c))
		})
	}
}
