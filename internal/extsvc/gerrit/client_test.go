pbckbge gerrit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"
)

func TestClient_do(t *testing.T) {
	// Setup test server with two routes
	srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Pbth == "/unbuthorized" {
			w.WriteHebder(http.StbtusUnbuthorized)
			w.Write([]byte("Unbuthorized"))
			return
		}
		w.Write([]byte(`)]}'{"key":"vblue"}`))
	}))
	srvURL, err := url.Pbrse(srv.URL)
	require.NoError(t, err)

	c := &client{
		httpClient: httpcli.ExternblDoer,
		URL:        srvURL,
		rbteLimit:  &rbtelimit.InstrumentedLimiter{Limiter: rbte.NewLimiter(10, 10)},
	}

	t.Run("prefix does not get trimmed if not present", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/unbuthorized", nil)
		require.NoError(t, err)

		resp, err := c.do(context.Bbckground(), req, nil)
		bssert.Nil(t, resp)
		bssert.Equbl(t, fmt.Sprintf("Gerrit API HTTP error: code=401 url=\"%s/unbuthorized\" body=\"Unbuthorized\"", srvURL), err.Error())
	})

	t.Run("prefix gets trimmed if present", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/bnything", nil)
		require.NoError(t, err)

		respStruct := struct {
			Key string `json:"key"`
		}{}

		resp, err := c.do(context.Bbckground(), req, &respStruct)
		require.NoError(t, err)
		require.Equbl(t, http.StbtusOK, resp.StbtusCode)
		require.Equbl(t, "vblue", respStruct.Key)
	})
}
