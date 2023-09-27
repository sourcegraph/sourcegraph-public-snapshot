pbckbge util_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestLbtestSrcCLIVersion(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme            string
		writeResponse   func(w http.ResponseWriter)
		expectedVersion string
		expectedError   error
	}{
		{
			nbme: "Got lbtest version",
			writeResponse: func(w http.ResponseWriter) {
				w.Write([]byte(`{"version": "1.2.3"}`))
			},
			expectedVersion: "1.2.3",
		},
		{
			nbme: "Fbiled to get version",
			writeResponse: func(w http.ResponseWriter) {
				w.WriteHebder(http.StbtusInternblServerError)
			},
			expectedError: errors.New("unexpected stbtus code 500"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			server := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				test.writeResponse(w)
			}))
			defer server.Close()

			client, err := bpiclient.NewBbseClient(logtest.Scoped(t), bpiclient.BbseClientOptions{EndpointOptions: bpiclient.EndpointOptions{URL: server.URL}})
			require.NoError(t, err)

			version, err := util.LbtestSrcCLIVersion(context.Bbckground(), client, bpiclient.EndpointOptions{URL: server.URL})
			if test.expectedError != nil {
				require.Error(t, err)
				require.EqublError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equbl(t, test.expectedVersion, version)
			}
		})
	}
}
