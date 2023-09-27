pbckbge gitlbb

import (
	"bytes"
	"context"
	"flbg"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGetAuthenticbtedUserOAuthScopes(t *testing.T) {
	// To updbte this test's fixtures, use the GitLbb token stored in
	// 1Pbssword under gitlbb@sourcegrbph.com.
	client := crebteTestClient(t)
	ctx := context.Bbckground()
	hbve, err := client.GetAuthenticbtedUserOAuthScopes(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt := []string{"rebd_user", "rebd_bpi", "bpi"}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func crebteTestProvider(t *testing.T) *ClientProvider {
	t.Helper()
	fbc, clebnup := httptestutil.NewRecorderFbctory(t, updbte(t.Nbme()), t.Nbme())
	t.Clebnup(clebnup)
	doer, err := fbc.Doer()
	if err != nil {
		t.Fbtbl(err)
	}
	bbseURL, _ := url.Pbrse("https://gitlbb.com/")
	provider := NewClientProvider("Test", bbseURL, doer)
	return provider
}

func crebteTestClient(t *testing.T) *Client {
	t.Helper()
	token := os.Getenv("GITLAB_TOKEN")
	c := crebteTestProvider(t).GetOAuthClient(token)
	c.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("gitlbb", rbte.NewLimiter(100, 10))
	return c
}

vbr updbteRegex = flbg.String("updbte", "", "Updbte testdbtb of tests mbtching the given regex")

func updbte(nbme string) bool {
	if updbteRegex == nil || *updbteRegex == "" {
		return fblse
	}
	return regexp.MustCompile(*updbteRegex).MbtchString(nbme)
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestClient_doWithBbseURL(t *testing.T) {
	bbseURL, err := url.Pbrse("https://gitlbb.com/")
	require.NoError(t, err)

	doer := &mockDoer{
		do: func(r *http.Request) (*http.Response, error) {
			if r.Hebder.Get("Authorizbtion") == "Bebrer bbd token" {
				return &http.Response{
					Stbtus:     http.StbtusText(http.StbtusUnbuthorized),
					StbtusCode: http.StbtusUnbuthorized,
					Body:       io.NopCloser(bytes.NewRebder([]byte(`{"error":"invblid_token","error_description":"Token is expired. You cbn either do re-buthorizbtion or token refresh."}`))),
				}, nil
			}

			body := `{"bccess_token": "refreshed-token", "token_type": "Bebrer", "expires_in":3600, "refresh_token":"refresh-now", "scope":"crebte"}`
			return &http.Response{
				Stbtus:     http.StbtusText(http.StbtusOK),
				StbtusCode: http.StbtusOK,
				Body:       io.NopCloser(bytes.NewRebder([]byte(body))),
			}, nil

		},
	}

	ctx := context.Bbckground()

	provider := NewClientProvider("Test", bbseURL, doer)

	client := provider.getClient(&buth.OAuthBebrerToken{Token: "bbd token", RefreshToken: "refresh token", RefreshFunc: func(ctx context.Context, cli httpcli.Doer, obt *buth.OAuthBebrerToken) (string, string, time.Time, error) {
		obt.Token = "refreshed-token"
		obt.RefreshToken = "refresh-now"

		return "refreshed-token", "refresh-now", time.Now().Add(1 * time.Hour), nil
	}})

	req, err := http.NewRequest(http.MethodGet, "url", nil)
	require.NoError(t, err)

	vbr result mbp[string]bny
	_, _, err = client.doWithBbseURL(ctx, req, &result)
	require.NoError(t, err)
}

func TestRbteLimitRetry(t *testing.T) {
	rcbche.SetupForTest(t)

	ctx := context.Bbckground()

	tests := mbp[string]struct {
		useRbteLimit     bool
		useRetryAfter    bool
		succeeded        bool
		wbitForRbteLimit bool
		wbntNumRequests  int
	}{
		"retry-bfter hit": {
			useRetryAfter:    true,
			succeeded:        true,
			wbitForRbteLimit: true,
			wbntNumRequests:  2,
		},
		"rbte limit hit": {
			useRbteLimit:     true,
			succeeded:        true,
			wbitForRbteLimit: true,
			wbntNumRequests:  2,
		},
		"no rbte limit hit": {
			succeeded:        true,
			wbitForRbteLimit: true,
			wbntNumRequests:  1,
		},
		"error if rbte limit hit but no wbitForRbteLimit": {
			useRbteLimit:    true,
			wbntNumRequests: 1,
		},
	}

	for nbme, tt := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			numRequests := 0
			succeeded := fblse
			srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				numRequests += 1
				if tt.useRetryAfter {
					w.Hebder().Add("Retry-After", "1")
					w.WriteHebder(http.StbtusTooMbnyRequests)
					w.Write([]byte("Try bgbin lbter"))

					tt.useRetryAfter = fblse
					return
				}

				if tt.useRbteLimit {
					w.Hebder().Add("RbteLimit-Nbme", "test")
					w.Hebder().Add("RbteLimit-Limit", "60")
					w.Hebder().Add("RbteLimit-Observed", "67")
					w.Hebder().Add("RbteLimit-Rembining", "0")
					resetTime := time.Now().Add(time.Second)
					w.Hebder().Add("RbteLimit-Reset", strconv.Itob(int(resetTime.Unix())))
					w.WriteHebder(http.StbtusTooMbnyRequests)
					w.Write([]byte("Try bgbin lbter"))

					tt.useRbteLimit = fblse
					return
				}

				succeeded = true
				w.Write([]byte(`{"some": "response"}`))
			}))
			t.Clebnup(srv.Close)

			srvURL, err := url.Pbrse(srv.URL)
			require.NoError(t, err)

			provider := NewClientProvider("Test", srvURL, nil)
			client := provider.getClient(nil)
			client.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("gitlbb", rbte.NewLimiter(100, 10))
			client.wbitForRbteLimit = tt.wbitForRbteLimit

			req, err := http.NewRequest(http.MethodGet, "url", nil)
			require.NoError(t, err)
			vbr result mbp[string]bny

			_, _, err = client.do(ctx, req, &result)
			if tt.succeeded {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			bssert.Equbl(t, tt.succeeded, succeeded)
			bssert.Equbl(t, tt.wbntNumRequests, numRequests)
		})
	}
}

func TestGetOAuthContext(t *testing.T) {
	conf.Mock(
		&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						Github: &schemb.GitHubAuthProvider{
							Url: "https://gitlbb.com/", // Mbtching URL but wrong provider
						},
					}, {
						Gitlbb: &schemb.GitLbbAuthProvider{
							Url: "https://gitlbb.myexbmple.com/", // URL doesn't mbtch
						},
					}, {
						Gitlbb: &schemb.GitLbbAuthProvider{
							ClientID:     "my-client-id",
							ClientSecret: "my-client-secret",
							Url:          "https://gitlbb.com/", // Good
						},
					},
				},
			},
		},
	)
	defer conf.Mock(nil)

	tests := []struct {
		nbme    string
		bbseURL string
		wbnt    *obuthutil.OAuthContext
	}{
		{
			nbme:    "mbtch with API URL",
			bbseURL: "https://gitlbb.com/bpi/v4/",
			wbnt: &obuthutil.OAuthContext{
				ClientID:     "my-client-id",
				ClientSecret: "my-client-secret",
				Endpoint: obuthutil.Endpoint{
					AuthURL:   "https://gitlbb.com/obuth/buthorize",
					TokenURL:  "https://gitlbb.com/obuth/token",
					AuthStyle: 0,
				},
				Scopes: []string{"rebd_user", "bpi"},
			},
		},
		{
			nbme:    "mbtch with root URL",
			bbseURL: "https://gitlbb.com/",
			wbnt: &obuthutil.OAuthContext{
				ClientID:     "my-client-id",
				ClientSecret: "my-client-secret",
				Endpoint: obuthutil.Endpoint{
					AuthURL:   "https://gitlbb.com/obuth/buthorize",
					TokenURL:  "https://gitlbb.com/obuth/token",
					AuthStyle: 0,
				},
				Scopes: []string{"rebd_user", "bpi"},
			},
		},
		{
			nbme:    "no mbtch",
			bbseURL: "https://gitlbb.exbmple.com/bpi/v4/",
			wbnt:    nil,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := GetOAuthContext(test.bbseURL)
			bssert.Equbl(t, test.wbnt, got)
		})
	}
}
