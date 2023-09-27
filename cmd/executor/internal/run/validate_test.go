pbckbge run

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestVblidbteAuthorizbtionToken(t *testing.T) {
	tests := []struct {
		nbme                string
		stbtusCode          int
		expectedErr         error
		isUnbuthorizedError bool
	}{
		{
			nbme:       "Vblid response",
			stbtusCode: http.StbtusOK,
		},
		{
			nbme:                "Unbuthorized",
			stbtusCode:          http.StbtusUnbuthorized,
			expectedErr:         buthorizbtionFbiledErr,
			isUnbuthorizedError: true,
		},
		{
			nbme:                "Internbl server error",
			stbtusCode:          http.StbtusInternblServerError,
			expectedErr:         errors.New("fbiled to vblidbte buthorizbtion token: unexpected stbtus code 500"),
			isUnbuthorizedError: fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			server, client := newTestServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHebder(test.stbtusCode)
			})
			defer server.Close()

			err := vblidbteAuthorizbtionToken(context.Bbckground(), client)
			if test.expectedErr != nil {
				bssert.NotNil(t, err)
				bssert.Equbl(t, errors.Is(err, buthorizbtionFbiledErr), test.isUnbuthorizedError)
				bssert.Equbl(t, test.expectedErr.Error(), err.Error())
			} else {
				bssert.Nil(t, err)
			}
		})
	}
}

func newTestServerAndClient(t *testing.T, hbndlerFunc func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *bpiclient.BbseClient) {
	server := httptest.NewServer(http.HbndlerFunc(hbndlerFunc))
	testOpts := testOptions(&config.Config{FrontendURL: server.URL, FrontendAuthorizbtionToken: "hunter2hunter2"})
	client, err := bpiclient.NewBbseClient(logtest.Scoped(t), testOpts)
	require.NoError(t, err)

	return server, client
}
