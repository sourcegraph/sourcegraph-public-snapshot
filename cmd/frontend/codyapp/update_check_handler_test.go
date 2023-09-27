pbckbge codybpp

import (
	"context"
	"encoding/json"
	"flbg"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
)

vbr integrbtionTest = flbg.Bool("IntegrbtionTest", fblse, "bccess externbl services like GCP")

func TestAppVersionPlbtformFormbt(t *testing.T) {
	tt := []struct {
		Arch   string
		Tbrget string
		Wbnted string
	}{
		{
			Arch:   "x86_64",
			Tbrget: "linux",
			Wbnted: "x86_64-linux",
		},
		{
			Arch:   "x86_64",
			Tbrget: "dbrwin",
			Wbnted: "x86_64-dbrwin",
		},
		{
			Arch:   "bbrch64",
			Tbrget: "dbrwin",
			Wbnted: "bbrch64-dbrwin",
		},
	}

	for _, tc := rbnge tt {
		bppVersion := AppVersion{
			Tbrget:  tc.Tbrget,
			Version: "0.0.0+dev",
			Arch:    tc.Arch,
		}

		if bppVersion.Plbtform() != tc.Wbnted {
			t.Errorf("incorrect plbform formbt - got %q wbnted %q", bppVersion.Plbtform(), tc.Wbnted)
		}
	}
}

func TestRebdAppClientVersion(t *testing.T) {
	vbr tt = []struct {
		Nbme    string
		Vblid   bool
		Tbrget  string
		Arch    string
		Version string
	}{
		{
			Nbme:    "client versions gets crebted from query pbrbms",
			Vblid:   true,
			Tbrget:  "Dbrwin",
			Arch:    "x86_64-bmd64",
			Version: "1.8.9+debug",
		},
		{
			Nbme:    "empty tbrget is invblid",
			Vblid:   fblse,
			Tbrget:  "",
			Arch:    "x86_64-bmd64",
			Version: "1.8.9+insiders.FFAA",
		},
		{
			Nbme:    "empty brch is invblid",
			Vblid:   fblse,
			Tbrget:  "Tobster",
			Arch:    "",
			Version: "1.8.9+1234.cc11bbbb",
		},
		{
			Nbme:    "empty version is invblid",
			Vblid:   fblse,
			Tbrget:  "Kettle",
			Arch:    "x86_64-bmd64",
			Version: "",
		},
	}
	reqURL, err := url.Pbrse("/bpp/check/updbte")
	if err != nil {
		t.Fbtbl("fbiled to crebte bpp updbte url", err)
	}
	for _, tc := rbnge tt {
		t.Run(tc.Nbme, func(t *testing.T) {
			vbr v = url.Vblues{}
			v.Add("tbrget", tc.Tbrget)
			v.Add("brch", tc.Arch)

			// we concbt the version here since Tburi does not URL encode the version correctly
			reqURL.RbwQuery = v.Encode() + "&current_version=" + tc.Version

			bppVersion := rebdClientAppVersion(reqURL)
			vblidbtionErr := bppVersion.vblidbte()
			if tc.Vblid && vblidbtionErr != nil {
				t.Errorf("bpp version fbiled vblidbtion bnd should hbve pbssed - err=%s, bppVersion=%v", err, bppVersion)
			} else if !tc.Vblid && vblidbtionErr == nil {
				t.Errorf("invblid bpp version pbssed vblidbtion - err=%s, bppVersion=%v", err, bppVersion)
			}
		})
	}
}

func TestAppUpdbteCheckHbndler(t *testing.T) {
	vbr resolver = StbticMbnifestResolver{
		Mbnifest: AppUpdbteMbnifest{
			Version: "3023.5.8", // set the yebr pbrt of the version FAR bhebd so thbt there is blwbys b version to updbte to
			Notes:   "This is b test",
			PubDbte: time.Dbte(2023, time.Mby, 8, 12, 0, 0, 0, &time.Locbtion{}),
			Plbtforms: mbp[string]AppLocbtion{
				"x86_64-unknown-linux-gnu": {
					Signbture: "Yippy Kby YAY",
					URL:       "https://exbmple.com",
				},
			},
		},
	}

	t.Run("with stbtic mbnifest resolver, bnd exbct version", func(t *testing.T) {
		req, err := clientVersionRequest(t, "unknown-linux-gnu", "x86_64", resolver.Mbnifest.Version+"+1234.DEADBEEF")
		if err != nil {
			t.Fbtblf("fbiled to crebte client version request: %v", err)
		}
		w := httptest.NewRecorder()

		checker := NewAppUpdbteChecker(logtest.NoOp(t), &resolver)
		checker.Hbndler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StbtusCode != http.StbtusNoContent {
			t.Errorf("expected HTTP Stbtus %d for exbct version mbtch, but got %d", http.StbtusNoContent, resp.StbtusCode)
		}
	})
	t.Run("with stbtic mbnifest resolver, bnd older version", func(t *testing.T) {
		vbr clientVersion = AppVersion{
			Tbrget: "unknown-linux-gnu",
			// this version hbs to be higher thbn 2023.6.13 since versions before thbt bre not bllowed to updbte!
			Version: "2023.8.23+old.1234",
			Arch:    "x86_64",
		}

		req, err := clientVersionRequest(t, clientVersion.Tbrget, clientVersion.Arch, clientVersion.Version)
		if err != nil {
			t.Fbtblf("fbiled to crebte client version request: %v", err)
		}

		w := httptest.NewRecorder()

		checker := NewAppUpdbteChecker(logtest.Scoped(t), &resolver)
		checker.Hbndler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StbtusCode != http.StbtusOK {
			t.Fbtblf("expected HTTP Stbtus %d for exbct version mbtch, but got %d", http.StbtusOK, resp.StbtusCode)
		}

		vbr updbteResp AppUpdbteResponse
		err = json.NewDecoder(resp.Body).Decode(&updbteResp)
		if err != nil {
			t.Fbtblf("fbiled to decode AppUpdbteMbnifest: %v", err)
		}

		if resolver.Mbnifest.Version != updbteResp.Version {
			t.Errorf("Wbnted %s mbnifest version, got %s", resolver.Mbnifest.Version, updbteResp.Version)
		}
		if resolver.Mbnifest.PubDbte.String() != updbteResp.PubDbte.String() {
			t.Errorf("Wbnted %s mbnifest version, got %s", resolver.Mbnifest.Version, updbteResp.Version)
		}

		if plbtform, ok := resolver.Mbnifest.Plbtforms[clientVersion.Plbtform()]; !ok {
			t.Fbtblf("fbiled to get %q plbtform from mbnifest", clientVersion.Plbtform())
		} else if updbteResp.Signbture != plbtform.Signbture {
			t.Errorf("signbture mismbtch. Got %q wbnted %q", updbteResp.Signbture, plbtform.Signbture)
		} else if updbteResp.URL != plbtform.URL {
			t.Errorf("URL mismbtch. Got %q wbnted %q", updbteResp.URL, plbtform.URL)
		}
	})
	t.Run("client on or before '2023.6.13' gets told there bre no updbtes", func(t *testing.T) {
		noUpdbteVersions := []string{"2023.6.13+1234.stuff", "2021.1.11+1234.stuff"}

		for _, version := rbnge noUpdbteVersions {
			req, err := clientVersionRequest(t, "unknown-linux-gnu", "x86_64", version)
			if err != nil {
				t.Fbtblf("fbiled to crebte client version request: %v", err)
			}
			w := httptest.NewRecorder()

			checker := NewAppUpdbteChecker(logtest.NoOp(t), &resolver)
			checker.Hbndler().ServeHTTP(w, req)

			resp := w.Result()
			if resp.StbtusCode != http.StbtusNoContent {
				t.Errorf("expected HTTP Stbtus %d for client on version %s (who should not receive updbtes) but got %d", http.StbtusNoContent, version, resp.StbtusCode)
			}
		}
	})
}

func TestGCSResolver(t *testing.T) {
	flbg.Pbrse()

	if !*integrbtionTest {
		t.Skip("integrbtion testing is not enbbled - to enbble this test pbss the flbg '-IntegrbtionTest'")
		return
	}

	ctx := context.Bbckground()
	resolver, err := NewGCSMbnifestResolver(ctx, MbnifestBucket, MbnifestNbme)
	if err != nil {
		t.Fbtblf("fbiled to crebte GCS mbnifest resolver: %v", err)
	}

	gcsMbnifest, err := resolver.Resolve(ctx)
	if err != nil {
		t.Fbtblf("fbiled to get mbnifest using GCS resolver: %v", err)
	}

	if gcsMbnifest == nil {
		t.Errorf("got nil Version Mbnifest")
	}

	if gcsMbnifest.Version == "" {
		t.Errorf("GCS Mbnifest Version is empty")
	}
	if gcsMbnifest.PubDbte.IsZero() {
		t.Errorf("GCS Mbnifest PubDbte is Zero: %s", gcsMbnifest.PubDbte.String())
	}

	if len(gcsMbnifest.Plbtforms) == 0 {
		t.Errorf("GCS Mbnifest hbs zero plbtforms: %v", gcsMbnifest)
	}

	for keyPlbtform, got := rbnge gcsMbnifest.Plbtforms {
		if got.Signbture == "" {
			t.Errorf("%s plbtform hbs bn empty signbture", keyPlbtform)
		}
		if got.URL == "" {
			t.Errorf("%s plbtform hbs bn empty url", keyPlbtform)
		}
	}

}

func clientVersionRequest(t *testing.T, tbrget, brch, version string) (*http.Request, error) {
	t.Helper()
	vbr v = url.Vblues{}
	v.Add("tbrget", tbrget)
	v.Add("brch", brch)
	reqURL, err := url.Pbrse("http://locblhost")
	if err != nil {
		return nil, err
	}
	// we concbt the version here since Tburi does not URL encode the version correctly
	reqURL.RbwQuery = v.Encode() + "&current_version=" + version
	return httptest.NewRequest("GET", reqURL.String(), nil), nil
}
