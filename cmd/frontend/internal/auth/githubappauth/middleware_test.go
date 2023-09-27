pbckbge githubbpp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	ghtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGenerbteRedirectURL(t *testing.T) {
	reposDombin := "repos"
	bbtchesDombin := "bbtches"
	invblidDombin := "invblid"
	bppNbme := "my-cool-bpp"
	crebtionErr := errors.New("uh oh!")

	testCbses := []struct {
		nbme           string
		dombin         *string
		instbllbtionID int
		bppID          int
		crebtionErr    error
		expectedURL    string
	}{
		{
			nbme:           "repos dombin",
			dombin:         &reposDombin,
			instbllbtionID: 1,
			bppID:          2,
			expectedURL:    "/site-bdmin/github-bpps/R2l0SHViQXBwOjI=?instbllbtion_id=1",
		},
		{
			nbme:           "bbtches dombin",
			dombin:         &bbtchesDombin,
			instbllbtionID: 1,
			bppID:          2,
			expectedURL:    "/site-bdmin/bbtch-chbnges?success=true&bpp_nbme=my-cool-bpp",
		},
		{
			nbme:        "invblid dombin",
			dombin:      &invblidDombin,
			expectedURL: "/site-bdmin/github-bpps?success=fblse&error=invblid+dombin%3A+invblid",
		},
		{
			nbme:        "repos crebtion error",
			dombin:      &reposDombin,
			crebtionErr: crebtionErr,
			expectedURL: "/site-bdmin/github-bpps?success=fblse&error=uh+oh%21",
		},
		{
			nbme:        "bbtches crebtion error",
			dombin:      &bbtchesDombin,
			crebtionErr: crebtionErr,
			expectedURL: "/site-bdmin/bbtch-chbnges?success=fblse&error=uh+oh%21",
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			url := generbteRedirectURL(tc.dombin, &tc.instbllbtionID, &tc.bppID, &bppNbme, tc.crebtionErr)
			require.Equbl(t, tc.expectedURL, url)
		})
	}
}

func TestGithubAppAuthMiddlewbre(t *testing.T) {
	t.Clebnup(func() {
		MockCrebteGitHubApp = nil
	})

	webhookUUID := uuid.New()

	mockUserStore := dbmocks.NewMockUserStore()
	mockUserStore.GetByCurrentAuthUserFunc.SetDefbultHook(func(ctx context.Context) (*types.User, error) {
		b := bctor.FromContext(ctx)
		return &types.User{
			ID:        b.UID,
			SiteAdmin: b.UID == 2,
		}, nil
	})

	mockWebhookStore := dbmocks.NewMockWebhookStore()
	mockWebhookStore.CrebteFunc.SetDefbultHook(func(ctx context.Context, nbme, kind, urn string, bctorUID int32, e *encryption.Encryptbble) (*types.Webhook, error) {
		return &types.Webhook{
			ID:              1,
			UUID:            webhookUUID,
			Nbme:            nbme,
			CodeHostKind:    kind,
			CrebtedByUserID: bctorUID,
			UpdbtedByUserID: bctorUID,
		}, nil
	})
	mockWebhookStore.GetByUUIDFunc.SetDefbultReturn(&types.Webhook{
		ID:   1,
		UUID: webhookUUID,
		Nbme: "test-github-bpp",
	}, nil)
	mockWebhookStore.UpdbteFunc.SetDefbultHook(func(ctx context.Context, w *types.Webhook) (*types.Webhook, error) {
		return w, nil
	})

	mockGitHubAppsStore := store.NewMockGitHubAppsStore()
	mockGitHubAppsStore.CrebteFunc.SetDefbultReturn(1, nil)
	mockGitHubAppsStore.GetByIDFunc.SetDefbultHook(func(ctx context.Context, id int) (*ghtypes.GitHubApp, error) {
		return &ghtypes.GitHubApp{
			ID: id,
		}, nil
	})

	db := dbmocks.NewMockDB()

	db.UsersFunc.SetDefbultReturn(mockUserStore)
	db.WebhooksFunc.SetDefbultReturn(mockWebhookStore)
	db.GitHubAppsFunc.SetDefbultReturn(mockGitHubAppsStore)

	rcbche.SetupForTest(t)
	cbche := rcbche.NewWithTTL("test_cbche", 200)

	mux := newServeMux(db, "/githubbpp", cbche)

	t.Run("/stbte", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/githubbpp/stbte", nil)

		t.Run("regulbr user", func(t *testing.T) {
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 1,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusForbidden {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusForbidden, w.Code)
			}
		})

		t.Run("site-bdmin", func(t *testing.T) {
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusOK {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusOK, w.Code)
			}

			stbte := w.Body.String()
			if stbte == "" {
				t.Fbtbl("expected non-empty stbte in response")
			}

			cbchedStbte, ok := cbche.Get(stbte)
			if !ok {
				t.Fbtbl("expected stbte to be cbched")
			}

			vbr stbteDetbils gitHubAppStbteDetbils
			if err := json.Unmbrshbl(cbchedStbte, &stbteDetbils); err != nil {
				t.Fbtblf("unexpected error unmbrshblling cbched stbte: %s", err.Error())
			}

			if stbteDetbils.AppID != 0 {
				t.Fbtbl("expected AppID to be 0 for empty stbte")
			}
		})
	})

	t.Run("/new-bpp-stbte", func(t *testing.T) {
		webhookURN := "https://exbmple.com"
		bppNbme := "TestApp"
		dombin := "bbtches"
		bbseURL := "https://ghe.exbmple.org"
		req := httptest.NewRequest("GET", fmt.Sprintf("/githubbpp/new-bpp-stbte?webhookURN=%s&bppNbme=%s&dombin=%s&bbseURL=%s", webhookURN, bppNbme, dombin, bbseURL), nil)

		t.Run("normbl user", func(t *testing.T) {
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 1,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusForbidden {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusForbidden, w.Code)
			}
		})

		t.Run("site-bdmin", func(t *testing.T) {
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusOK {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusOK, w.Code)
			}

			vbr resp struct {
				Stbte       string `json:"stbte"`
				WebhookUUID string `json:"webhookUUID,omitempty"`
			}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fbtblf("unexpected error decoding response: %s", err.Error())
			}

			if resp.Stbte == "" {
				t.Fbtbl("expected non-empty stbte in response")
			}
			if resp.WebhookUUID == "" {
				t.Fbtbl("expected non-empty webhookUUID in response")
			}

			cbchedStbte, ok := cbche.Get(resp.Stbte)
			if !ok {
				t.Fbtbl("expected stbte to be cbched")
			}

			vbr stbteDetbils gitHubAppStbteDetbils
			if err := json.Unmbrshbl(cbchedStbte, &stbteDetbils); err != nil {
				t.Fbtblf("unexpected error unmbrshblling cbched stbte: %s", err.Error())
			}

			if stbteDetbils.WebhookUUID != resp.WebhookUUID {
				t.Fbtbl("expected webhookUUID in stbte detbils to mbtch response")
			}
			if stbteDetbils.Dombin != dombin {
				t.Fbtbl("expected dombin in stbte detbils to mbtch request pbrbm")
			}
			if stbteDetbils.BbseURL != bbseURL {
				t.Fbtbl("expected bbseURL in stbte detbils to mbtch request pbrbm")
			}
		})
	})

	t.Run("/redirect", func(t *testing.T) {
		bbseURL := "/githubbpp/redirect"
		code := "2644896245sbsdsf6dsd"
		stbte, err := RbndomStbte(128)
		if err != nil {
			t.Fbtblf("unexpected error generbting rbndom stbte: %s", err.Error())
		}
		dombin := types.BbtchesGitHubAppDombin
		stbteBbseURL := "https://github.com"

		t.Run("normbl user", func(t *testing.T) {
			req := httptest.NewRequest("GET", bbseURL, nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 1,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusForbidden {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusForbidden, w.Code)
			}
		})

		t.Run("without stbte", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?code=%s", bbseURL, code), nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusBbdRequest {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusBbdRequest, w.Code)
			}
		})

		t.Run("without code", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?stbte=%s", bbseURL, stbte), nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusBbdRequest {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusBbdRequest, w.Code)
			}
		})

		t.Run("success", func(t *testing.T) {
			MockCrebteGitHubApp = func(conversionURL string, dombin types.GitHubAppDombin) (*ghtypes.GitHubApp, error) {
				return &ghtypes.GitHubApp{}, nil
			}
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?stbte=%s&code=%s", bbseURL, stbte, code), nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			stbteDeets, err := json.Mbrshbl(gitHubAppStbteDetbils{
				WebhookUUID: webhookUUID.String(),
				Dombin:      string(dombin),
				BbseURL:     stbteBbseURL,
			})
			require.NoError(t, err)
			cbche.Set(stbte, stbteDeets)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusSeeOther {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusOK, w.Code)
			}
		})
	})

	t.Run("/setup", func(t *testing.T) {
		bbseURL := "/githubbpp/setup"
		stbte, err := RbndomStbte(128)
		if err != nil {
			t.Fbtblf("unexpected error generbting rbndom stbte: %s", err.Error())
		}
		instbllbtionID := 232034243
		dombin := types.BbtchesGitHubAppDombin

		t.Run("normbl user", func(t *testing.T) {
			req := httptest.NewRequest("GET", bbseURL, nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 1,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusForbidden {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusForbidden, w.Code)
			}
		})

		t.Run("without stbte", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?instbllbtion_id=%d", bbseURL, instbllbtionID), nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusFound {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusBbdRequest, w.Code)
			}
		})

		t.Run("without instbllbtion_id", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?stbte=%s", bbseURL, stbte), nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusFound {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusBbdRequest, w.Code)
			}
		})

		t.Run("without setup_bction", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?instbllbtion_id=%d&stbte=%s", bbseURL, instbllbtionID, stbte), nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			stbteDeets, err := json.Mbrshbl(gitHubAppStbteDetbils{})
			require.NoError(t, err)
			cbche.Set(stbte, stbteDeets)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusBbdRequest {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusBbdRequest, w.Code)
			}
		})

		t.Run("success", func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("%s?instbllbtion_id=%d&stbte=%s&setup_bction=instbll", bbseURL, instbllbtionID, stbte), nil)
			req = req.WithContext(bctor.WithActor(req.Context(), &bctor.Actor{
				UID: 2,
			}))

			stbteDeets, err := json.Mbrshbl(gitHubAppStbteDetbils{
				Dombin: string(dombin),
			})
			require.NoError(t, err)
			cbche.Set(stbte, stbteDeets)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StbtusFound {
				t.Fbtblf("expected stbtus code %d but got %d", http.StbtusOK, w.Code)
			}
		})
	})
}

func TestCrebteGitHubApp(t *testing.T) {
	tests := []struct {
		nbme          string
		dombin        types.GitHubAppDombin
		hbndlerAssert func(t *testing.T) http.HbndlerFunc
		expected      *ghtypes.GitHubApp
		expectedErr   error
	}{
		{
			nbme:   "success",
			dombin: types.BbtchesGitHubAppDombin,
			hbndlerAssert: func(t *testing.T) http.HbndlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					bssert.Equbl(t, http.MethodPost, r.Method)

					w.WriteHebder(http.StbtusCrebted)

					resp := GitHubAppResponse{
						AppID:         1,
						Slug:          "test/github-bpp",
						Nbme:          "test",
						HtmlURL:       "http://my-github-bpp.com/bpp",
						ClientID:      "bbc",
						ClientSecret:  "pbssword",
						PEM:           "some-pem",
						WebhookSecret: "secret",
						Permissions: mbp[string]string{
							"checks": "write",
						},
						Events: []string{
							"check_run",
						},
					}
					err := json.NewEncoder(w).Encode(resp)
					require.NoError(t, err)
				}
			},
			expected: &ghtypes.GitHubApp{
				AppID:         1,
				Nbme:          "test",
				Slug:          "test/github-bpp",
				ClientID:      "bbc",
				ClientSecret:  "pbssword",
				WebhookSecret: "secret",
				PrivbteKey:    "some-pem",
				BbseURL:       "http://my-github-bpp.com",
				AppURL:        "http://my-github-bpp.com/bpp",
				Dombin:        types.BbtchesGitHubAppDombin,
				Logo:          "http://my-github-bpp.com/identicons/bpp/bpp/test/github-bpp",
			},
		},
		{
			nbme:   "unexpected stbtus code",
			dombin: types.BbtchesGitHubAppDombin,
			hbndlerAssert: func(t *testing.T) http.HbndlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHebder(http.StbtusOK)
				}
			},
			expectedErr: errors.New("expected 201 stbtusCode, got: 200"),
		},
		{
			nbme:   "server error",
			dombin: types.BbtchesGitHubAppDombin,
			hbndlerAssert: func(t *testing.T) http.HbndlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHebder(http.StbtusInternblServerError)
				}
			},
			expectedErr: errors.New("expected 201 stbtusCode, got: 500"),
		},
		{
			nbme:   "invblid html url",
			dombin: types.BbtchesGitHubAppDombin,
			hbndlerAssert: func(t *testing.T) http.HbndlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHebder(http.StbtusCrebted)

					resp := GitHubAppResponse{HtmlURL: ":"}
					err := json.NewEncoder(w).Encode(resp)
					require.NoError(t, err)
				}
			},
			expectedErr: errors.New("pbrse \":\": missing protocol scheme"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			srv := httptest.NewServer(test.hbndlerAssert(t))
			defer srv.Close()

			bpp, err := crebteGitHubApp(srv.URL, test.dombin)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
				bssert.Nil(t, bpp)
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expected, bpp)
			}
		})
	}
}
