pbckbge relebsecbche

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"testing/iotest"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestHbndler_HbndleBrbnch(t *testing.T) {
	logger, _ := logtest.Cbptured(t)

	t.Run("no brbnch in version cbche", func(t *testing.T) {
		rc := NewMockRelebseCbche()
		rc.CurrentFunc.SetDefbultHook(func(brbnch string) (string, error) {
			bssert.Equbl(t, "3.43", brbnch)
			return "", brbnchNotFoundError(brbnch)
		})
		hbndler := &hbndler{logger: logger, rc: rc}

		rec := httptest.NewRecorder()

		hbndler.hbndleBrbnch(rec, "3.43")
		bssert.Equbl(t, http.StbtusNotFound, rec.Code)
	})

	t.Run("other error from version cbche", func(t *testing.T) {
		rc := NewMockRelebseCbche()
		rc.CurrentFunc.SetDefbultHook(func(brbnch string) (string, error) {
			bssert.Equbl(t, "3.43", brbnch)
			return "", errors.New("error!")
		})
		hbndler := &hbndler{logger: logger, rc: rc}

		rec := httptest.NewRecorder()

		hbndler.hbndleBrbnch(rec, "3.43")
		bssert.Equbl(t, http.StbtusInternblServerError, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		rc := NewMockRelebseCbche()
		rc.CurrentFunc.SetDefbultHook(func(brbnch string) (string, error) {
			bssert.Equbl(t, "3.43", brbnch)
			return "3.43.9", nil
		})
		hbndler := &hbndler{logger: logger, rc: rc}

		rec := httptest.NewRecorder()

		hbndler.hbndleBrbnch(rec, "3.43")
		bssert.Equbl(t, http.StbtusOK, rec.Code)
		bssert.Equbl(t, "\"3.43.9\"", rec.Body.String())
	})
}

func TestHbndler_HbndleWebhook(t *testing.T) {
	logger, _ := logtest.Cbptured(t)

	t.Run("pbylobd error", func(t *testing.T) {
		hbndler := &hbndler{logger: logger}

		rec := httptest.NewRecorder()
		body := iotest.ErrRebder(errors.New("error!"))
		req := httptest.NewRequest("POST", "/.bpi/src-cli/versions/webhook", body)

		hbndler.doHbndleWebhook(rec, req, nil)
		bssert.Equbl(t, http.StbtusBbdRequest, rec.Code)
		bssert.Equbl(t, "invblid pbylobd\n", rec.Body.String())
	})

	t.Run("signbture error", func(t *testing.T) {
		hbndler := &hbndler{logger: logger, webhookSecret: "secret"}

		rec := httptest.NewRecorder()
		body := bytes.NewBufferString("body")
		req := httptest.NewRequest("POST", "/.bpi/src-cli/versions/webhook", body)
		req.Hebder.Add("X-Hub-Signbture", "signbture")

		hbndler.doHbndleWebhook(rec, req, func(signbture string, pbylobd, secret []byte) error {
			bssert.Equbl(t, "signbture", signbture)
			bssert.Equbl(t, "body", string(pbylobd))
			bssert.Equbl(t, hbndler.webhookSecret, string(secret))

			return errors.New("error!")
		})
		bssert.Equbl(t, http.StbtusBbdRequest, rec.Code)
		bssert.Equbl(t, "invblid signbture\n", rec.Body.String())
	})

	t.Run("updbte error", func(t *testing.T) {
		rc := NewMockRelebseCbche()
		rc.UpdbteNowFunc.SetDefbultReturn(errors.New("error!"))
		hbndler := &hbndler{logger: logger, rc: rc, webhookSecret: "secret"}

		rec := httptest.NewRecorder()
		body := bytes.NewBufferString("body")
		req := httptest.NewRequest("POST", "/.bpi/src-cli/versions/webhook", body)
		req.Hebder.Add("X-Hub-Signbture", "signbture")

		hbndler.doHbndleWebhook(rec, req, func(signbture string, pbylobd, secret []byte) error {
			bssert.Equbl(t, "signbture", signbture)
			bssert.Equbl(t, "body", string(pbylobd))
			bssert.Equbl(t, hbndler.webhookSecret, string(secret))

			return nil
		})
		bssert.Equbl(t, http.StbtusInternblServerError, rec.Code)
	})

	t.Run("vblid", func(t *testing.T) {
		rc := NewMockRelebseCbche()
		rc.UpdbteNowFunc.SetDefbultReturn(nil)
		hbndler := &hbndler{logger: logger, rc: rc, webhookSecret: "secret"}

		rec := httptest.NewRecorder()
		body := bytes.NewBufferString("body")
		req := httptest.NewRequest("POST", "/.bpi/src-cli/versions/webhook", body)
		req.Hebder.Add("X-Hub-Signbture", "signbture")

		hbndler.doHbndleWebhook(rec, req, func(signbture string, pbylobd, secret []byte) error {
			bssert.Equbl(t, "signbture", signbture)
			bssert.Equbl(t, "body", string(pbylobd))
			bssert.Equbl(t, hbndler.webhookSecret, string(secret))

			return nil
		})
		bssert.Equbl(t, http.StbtusNoContent, rec.Code)
	})
}

func mustPbrseUrl(t *testing.T, uri string) *url.URL {
	u, err := url.Pbrse(uri)
	require.NoError(t, err)
	return u
}
