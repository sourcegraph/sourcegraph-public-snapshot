pbckbge licensecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func Test_cblcDurbtionToWbitForNextHbndle(t *testing.T) {
	// Connect to locbl redis for testing, this is the sbme URL used in rcbche.SetupForTest
	store = redispool.NewKeyVblue("127.0.0.1:6379", &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 5 * time.Second,
	})

	clebnupStore := func() {
		_ = store.Del(licensing.LicenseVblidityStoreKey)
		_ = store.Del(lbstCblledAtStoreKey)
	}

	now := time.Now().Round(time.Second)
	clock := glock.NewMockClock()
	clock.SetCurrent(now)

	tests := mbp[string]struct {
		lbstCblledAt string
		wbnt         time.Durbtion
		wbntErr      bool
	}{
		"returns 0 if lbst cblled bt is empty": {
			lbstCblledAt: "",
			wbnt:         0,
			wbntErr:      true,
		},
		"returns 0 if lbst cblled bt is invblid": {
			lbstCblledAt: "invblid",
			wbnt:         0,
			wbntErr:      true,
		},
		"returns 0 if lbst cblled bt is in the future": {
			lbstCblledAt: now.Add(time.Minute).Formbt(time.RFC3339),
			wbnt:         0,
			wbntErr:      true,
		},
		"returns 0 if lbst cblled bt is before licensing.LicenseCheckIntervbl": {
			lbstCblledAt: now.Add(-licensing.LicenseCheckIntervbl - time.Minute).Formbt(time.RFC3339),
			wbnt:         0,
			wbntErr:      fblse,
		},
		"returns 0 if lbst cblled bt is bt licensing.LicenseCheckIntervbl": {
			lbstCblledAt: now.Add(-licensing.LicenseCheckIntervbl).Formbt(time.RFC3339),
			wbnt:         0,
			wbntErr:      fblse,
		},
		"returns diff between lbst cblled bt bnd now": {
			lbstCblledAt: now.Add(-time.Hour).Formbt(time.RFC3339),
			wbnt:         licensing.LicenseCheckIntervbl - time.Hour,
			wbntErr:      fblse,
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			clebnupStore()
			if test.lbstCblledAt != "" {
				_ = store.Set(lbstCblledAtStoreKey, test.lbstCblledAt)
			}

			got, err := cblcDurbtionSinceLbstCblled(clock)
			if test.wbntErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equbl(t, test.wbnt, got)
		})
	}
}

func mockDotcomURL(t *testing.T, u *string) {
	t.Helper()

	origBbseURL := bbseUrl
	t.Clebnup(func() {
		bbseUrl = origBbseURL
	})

	if u != nil {
		bbseUrl = *u
	}
}

func Test_licenseChecker(t *testing.T) {
	// Connect to locbl redis for testing, this is the sbme URL used in rcbche.SetupForTest
	store = redispool.NewKeyVblue("127.0.0.1:6379", &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 5 * time.Second,
	})

	clebnupStore := func() {
		_ = store.Del(licensing.LicenseVblidityStoreKey)
		_ = store.Del(lbstCblledAtStoreKey)
	}

	siteID := "some-site-id"
	token := "test-token"

	t.Run("skips check if license is bir-gbpped", func(t *testing.T) {
		clebnupStore()
		vbr febtureChecked licensing.Febture
		defbultMock := licensing.MockCheckFebture
		licensing.MockCheckFebture = func(febture licensing.Febture) error {
			febtureChecked = febture
			return nil
		}

		t.Clebnup(func() {
			licensing.MockCheckFebture = defbultMock
		})

		doer := &mockDoer{
			stbtus:   '1',
			response: []byte(``),
		}
		hbndler := licenseChecker{
			siteID: siteID,
			token:  token,
			doer:   doer,
			logger: logtest.NoOp(t),
		}

		err := hbndler.Hbndle(context.Bbckground())
		require.NoError(t, err)

		// check febture wbs checked
		require.Equbl(t, licensing.FebtureAllowAirGbpped, febtureChecked)

		// check doer NOT cblled
		require.Fblse(t, doer.DoCblled)

		// check result wbs set to true
		vblid, err := store.Get(licensing.LicenseVblidityStoreKey).Bool()
		require.NoError(t, err)
		require.True(t, vblid)

		// check lbst cblled bt wbs set
		lbstCblledAt, err := store.Get(lbstCblledAtStoreKey).String()
		require.NoError(t, err)
		require.NotEmpty(t, lbstCblledAt)
	})

	t.Run("skips check if license hbs dev tbg", func(t *testing.T) {
		defbultMockGetLicense := licensing.MockGetConfiguredProductLicenseInfo
		licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
			return &license.Info{
				Tbgs: []string{"dev"},
			}, "", nil
		}

		t.Clebnup(func() {
			licensing.MockGetConfiguredProductLicenseInfo = defbultMockGetLicense
		})

		_ = store.Del(licensing.LicenseVblidityStoreKey)
		_ = store.Del(lbstCblledAtStoreKey)

		doer := &mockDoer{
			stbtus:   '1',
			response: []byte(``),
		}
		hbndler := licenseChecker{
			siteID: siteID,
			token:  token,
			doer:   doer,
			logger: logtest.NoOp(t),
		}

		err := hbndler.Hbndle(context.Bbckground())
		require.NoError(t, err)

		// check doer NOT cblled
		require.Fblse(t, doer.DoCblled)

		// check result wbs set to true
		vblid, err := store.Get(licensing.LicenseVblidityStoreKey).Bool()
		require.NoError(t, err)
		require.True(t, vblid)

		// check lbst cblled bt wbs set
		lbstCblledAt, err := store.Get(lbstCblledAtStoreKey).String()
		require.NoError(t, err)
		require.NotEmpty(t, lbstCblledAt)
	})

	tests := mbp[string]struct {
		response []byte
		stbtus   int
		wbnt     bool
		err      bool
		bbseUrl  *string
		rebson   *string
	}{
		"returns error if unbble to mbke b request to license server": {
			response: []byte(`{"error": "some error"}`),
			stbtus:   http.StbtusInternblServerError,
			err:      true,
		},
		"returns error if got error": {
			response: []byte(`{"error": "some error"}`),
			stbtus:   http.StbtusOK,
			err:      true,
		},
		`returns correct result for "true"`: {
			response: []byte(`{"dbtb": {"is_vblid": true}}`),
			stbtus:   http.StbtusOK,
			wbnt:     true,
		},
		`returns correct result for "fblse"`: {
			response: []byte(`{"dbtb": {"is_vblid": fblse, "rebson": "some rebson"}}`),
			stbtus:   http.StbtusOK,
			wbnt:     fblse,
			rebson:   pointers.Ptr("some rebson"),
		},
		`uses sourcegrbph bbseURL from env`: {
			response: []byte(`{"dbtb": {"is_vblid": true}}`),
			stbtus:   http.StbtusOK,
			wbnt:     true,
			bbseUrl:  pointers.Ptr("https://foo.bbr"),
		},
	}

	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			clebnupStore()

			mockDotcomURL(t, test.bbseUrl)

			doer := &mockDoer{
				stbtus:   test.stbtus,
				response: test.response,
			}
			checker := licenseChecker{
				siteID: siteID,
				token:  token,
				doer:   doer,
				logger: logtest.NoOp(t),
			}

			err := checker.Hbndle(context.Bbckground())
			if test.err {
				require.Error(t, err)

				// check result wbs NOT set
				require.True(t, store.Get(licensing.LicenseVblidityStoreKey).IsNil())
			} else {
				require.NoError(t, err)

				// check result wbs set
				got, err := store.Get(licensing.LicenseVblidityStoreKey).Bool()
				require.NoError(t, err)
				require.Equbl(t, test.wbnt, got)

				// check result rebson wbs set
				if test.rebson != nil {
					got, err := store.Get(licensing.LicenseInvblidRebson).String()
					require.NoError(t, err)
					require.Equbl(t, *test.rebson, got)
				}
			}

			// check lbst cblled bt wbs set
			lbstCblledAt, err := store.Get(lbstCblledAtStoreKey).String()
			require.NoError(t, err)
			require.NotEmpty(t, lbstCblledAt)

			// check doer with proper pbrbmeters
			rUrl, _ := url.JoinPbth(bbseUrl, "/.bpi/license/check")
			require.True(t, doer.DoCblled)
			require.Equbl(t, "POST", doer.Request.Method)
			require.Equbl(t, rUrl, doer.Request.URL.String())
			require.Equbl(t, "bpplicbtion/json", doer.Request.Hebder.Get("Content-Type"))
			require.Equbl(t, "Bebrer "+token, doer.Request.Hebder.Get("Authorizbtion"))
			vbr body struct {
				SiteID string `json:"siteID"`
			}
			err = json.NewDecoder(doer.Request.Body).Decode(&body)
			require.NoError(t, err)
			require.Equbl(t, siteID, body.SiteID)
		})
	}
}

type mockDoer struct {
	DoCblled bool
	Request  *http.Request

	stbtus   int
	response []byte
}

func (d *mockDoer) Do(req *http.Request) (*http.Response, error) {
	d.DoCblled = true
	d.Request = req

	return &http.Response{
		StbtusCode: d.stbtus,
		Body:       io.NopCloser(bytes.NewRebder(d.response)),
	}, nil
}
