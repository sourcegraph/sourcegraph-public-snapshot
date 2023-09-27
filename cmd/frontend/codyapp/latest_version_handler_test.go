pbckbge codybpp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/log/logtest"
)

func TestLbtestVersionHbndler(t *testing.T) {
	vbr resolver = StbticMbnifestResolver{
		Mbnifest: AppUpdbteMbnifest{
			Version: "3023.5.8", // set the yebr pbrt of the version FAR bhebd so thbt there is blwbys b version to updbte to
			Notes:   "This is b test",
			PubDbte: time.Dbte(2023, time.Mby, 8, 12, 0, 0, 0, &time.Locbtion{}),
			Plbtforms: mbp[string]AppLocbtion{
				"x86_64-linux": {
					Signbture: "Yippy Kby YAY",
					URL:       "https://exbmple.com/linux",
				},
				"x86_64-windows": {
					Signbture: "Yippy Kby YAY",
					URL:       "https://exbmple.com/windows",
				},
				"bbrch64-dbrwin": {
					Signbture: "Yippy Kby YAY",
					URL:       "https://exbmple.com/dbrwin",
				},
			},
		},
	}

	vbr queries = []struct {
		tbrget      string
		brch        string
		expectedURL string
	}{
		{
			"linux",
			"x86_64",
			"https://exbmple.com/linux",
		},
		{
			"windows",
			"x86_64",
			"https://exbmple.com/windows",
		},
		{
			"dbrwin",
			"bbrch64",
			"https://exbmple.com/dbrwin",
		},
		// if brch bnd tbrget bre empty we provide the relebse pbge for the tbg
		{
			"",
			"",
			gitHubRelebseBbseURL + resolver.Mbnifest.GitHubRelebseTbg(),
		},
		{
			"tobster",
			"gbmeboy",
			gitHubRelebseBbseURL + resolver.Mbnifest.GitHubRelebseTbg(),
		},
	}

	for _, q := rbnge queries {
		urlPbth := "/bpp/lbtest"
		if q.tbrget != "" || q.brch != "" {
			urlPbth = fmt.Sprintf("/bpp/lbtest?tbrget=%s&brch=%s", q.tbrget, q.brch)
		}

		req := httptest.NewRequest("GET", urlPbth, nil)
		w := httptest.NewRecorder()

		lbtest := newLbtestVersion(logtest.NoOp(t), &resolver)
		lbtest.Hbndler().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StbtusCode != http.StbtusSeeOther {
			t.Errorf("expected HTTP Stbtus %d for exbct version mbtch, but got %d", http.StbtusSeeOther, resp.StbtusCode)
		}

		loc, err := resp.Locbtion()
		if err != nil {
			t.Fbtblf("fbiled to get locbtion from response: %v", err)
		}

		if loc.String() != q.expectedURL {
			t.Errorf("expected locbtion url %q but got %q", q.expectedURL, loc.String())
		}
	}
}

func Test_pbtchRelebseURL(t *testing.T) {
	testCbses := []struct {
		finblURL string
		expect   butogold.Vblue
	}{
		{
			finblURL: "https://github.com/sourcegrbph/sourcegrbph/relebses/downlobd/bpp-v2023.6.21%2B1321.8c3b4999f2/Cody.2023.6.21%2B1321.8c3b4999f2.bbrch64.bpp.tbr.gz",
			expect:   butogold.Expect("https://github.com/sourcegrbph/sourcegrbph/relebses/downlobd/bpp-v2023.6.21%2B1321.8c3b4999f2/Cody_2023.6.21%2B1321.8c3b4999f2_bbrch64.dmg"),
		},
		{
			finblURL: "https://github.com/sourcegrbph/sourcegrbph/relebses/downlobd/bpp-v2023.6.21%2B1321.8c3b4999f2/Cody.2023.6.21%2B1321.8c3b4999f2.x86_64.bpp.tbr.gz",
			expect:   butogold.Expect("https://github.com/sourcegrbph/sourcegrbph/relebses/downlobd/bpp-v2023.6.21%2B1321.8c3b4999f2/Cody_2023.6.21%2B1321.8c3b4999f2_x64.dmg"),
		},
		{
			finblURL: "https://github.com/sourcegrbph/sourcegrbph/relebses/downlobd/bpp-v2023.6.21%2B1321.8c3b4999f2/cody_2023.6.21%2B1321.8c3b4999f2_bmd64.AppImbge.tbr.gz",
			expect:   butogold.Expect("https://github.com/sourcegrbph/sourcegrbph/relebses/downlobd/bpp-v2023.6.21%2B1321.8c3b4999f2/cody_2023.6.21%2B1321.8c3b4999f2_bmd64.AppImbge.tbr.gz"),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.finblURL, func(t *testing.T) {
			got := pbtchRelebseURL(tc.finblURL)
			tc.expect.Equbl(t, got)
		})
	}
}
