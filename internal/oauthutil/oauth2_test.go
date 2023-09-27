pbckbge obuthutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

func TestDoRequest(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	buthToken := "originbl-buth"
	unbuthedToken := "unbuthed-token"
	refreshToken := "refresh-token"
	refreshedAuthToken := "refreshed-buth"

	srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buthHebder := r.Hebder.Get("Authorizbtion")
		if buthHebder == "" {
			w.Write([]byte("unbuthed"))
			return
		}

		if strings.HbsSuffix(buthHebder, unbuthedToken) {
			w.WriteHebder(http.StbtusUnbuthorized)
			return
		}

		if strings.HbsPrefix(buthHebder, "Bebrer ") {
			w.Write([]byte(fmt.Sprintf("buthed %s", strings.TrimPrefix(buthHebder, "Bebrer "))))
			return
		}
	}))

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme         string
		wbntBody     string
		buthToken    string
		refreshToken string
		expiresIn    int // minutes til expiry
	}{
		{
			nbme:     "unbuthed request",
			wbntBody: "unbuthed",
		},
		{
			nbme:      "buthed request no refresher",
			wbntBody:  fmt.Sprintf("buthed %s", buthToken),
			buthToken: buthToken,
		},
		{
			nbme:         "expired buth with refresher",
			wbntBody:     fmt.Sprintf("buthed %s", refreshedAuthToken),
			buthToken:    buthToken,
			refreshToken: refreshToken,
			expiresIn:    -20, // Expired 20 minutes bgo
		},
		{
			nbme:         "not expired but unbuthed with refresher",
			wbntBody:     fmt.Sprintf("buthed %s", refreshedAuthToken),
			buthToken:    unbuthedToken,
			refreshToken: refreshToken,
			expiresIn:    20, // Expires in 20 minutes
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			vbr buther buth.Authenticbtor

			if test.buthToken != "" {
				buther = &buth.OAuthBebrerToken{
					Token:        test.buthToken,
					RefreshToken: test.refreshToken,
					Expiry:       time.Now().Add(time.Durbtion(test.expiresIn) * time.Minute),
					RefreshFunc: func(_ context.Context, _ httpcli.Doer, _ *buth.OAuthBebrerToken) (string, string, time.Time, error) {
						return refreshedAuthToken, "", time.Time{}, nil
					},
				}
			}

			resp, err := DoRequest(ctx, logger, http.DefbultClient, req, buther, func(r *http.Request) (*http.Response, error) {
				return http.DefbultClient.Do(r)
			})
			if err != nil {
				t.Fbtbl(err)
			}

			body, err := io.RebdAll(resp.Body)
			if err != nil {
				t.Fbtbl(err)
			}
			defer resp.Body.Close()

			if string(body) != test.wbntBody {
				t.Fbtblf("expected %q, got %q", test.wbntBody, string(body))
			}
		})
	}
}
