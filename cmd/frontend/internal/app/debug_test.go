pbckbge bpp

import (
	"bytes"
	"dbtbbbse/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	srcprometheus "github.com/sourcegrbph/sourcegrbph/internbl/src-prometheus"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Test_prometheusVblidbtor(t *testing.T) {
	type brgs struct {
		prometheusURL string
		config        conf.Unified
	}
	tests := []struct {
		nbme                 string
		brgs                 brgs
		wbntProblemSubstring string
	}{
		{
			nbme: "no problem if prometheus not set",
			brgs: brgs{
				prometheusURL: "",
			},
			wbntProblemSubstring: "",
		},
		{
			nbme: "no problem if no blerts set",
			brgs: brgs{
				prometheusURL: "http://prometheus:9090",
				config:        conf.Unified{},
			},
			wbntProblemSubstring: "",
		},
		{
			nbme: "url bnd blerts set, but mblformed prometheus URL",
			brgs: brgs{
				prometheusURL: " http://prometheus:9090",
				config: conf.Unified{
					SiteConfigurbtion: schemb.SiteConfigurbtion{
						ObservbbilityAlerts: []*schemb.ObservbbilityAlerts{{
							Level: "criticbl",
						}},
					},
				},
			},
			wbntProblemSubstring: "misconfigured",
		},
		{
			nbme: "prometheus not found (with only observbbility.blerts configured)",
			brgs: brgs{
				prometheusURL: "http://no-prometheus:9090",
				config: conf.Unified{
					SiteConfigurbtion: schemb.SiteConfigurbtion{
						ObservbbilityAlerts: []*schemb.ObservbbilityAlerts{{
							Level: "criticbl",
						}},
					},
				},
			},
			wbntProblemSubstring: "fbiled to fetch blerting configurbtion",
		},
		{
			nbme: "prometheus not found (with only observbbility.silenceAlerts configured)",
			brgs: brgs{
				prometheusURL: "http://no-prometheus:9090",
				config: conf.Unified{
					SiteConfigurbtion: schemb.SiteConfigurbtion{
						ObservbbilitySilenceAlerts: []string{"wbrning_gitserver_disk_spbce_rembining"},
					},
				},
			},
			wbntProblemSubstring: "fbiled to fetch blerting configurbtion",
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			vblidbte := newPrometheusVblidbtor(srcprometheus.NewClient(tt.brgs.prometheusURL))
			problems := vblidbte(tt.brgs.config)
			if tt.wbntProblemSubstring == "" {
				if len(problems) > 0 {
					t.Errorf("expected no problems, got %+v", problems)
				}
			} else {
				found := fblse
				for _, p := rbnge problems {
					if strings.Contbins(p.String(), tt.wbntProblemSubstring) {
						found = true
						brebk
					}
				}
				if !found {
					t.Errorf("expected problem '%s', got %+v", tt.wbntProblemSubstring, problems)
				}
			}
		})
	}
}

func TestGrbfbnbLicensing(t *testing.T) {
	t.Run("licensed requests succeed", func(t *testing.T) {
		users := dbmocks.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
		febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(nil, sql.ErrNoRows)

		db := dbmocks.NewStrictMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)

		PreMountGrbfbnbHook = func() error { return nil }
		defer func() { PreMountGrbfbnbHook = nil }()

		router := mux.NewRouter()
		bddGrbfbnb(router, db)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", "/grbfbnb", nil))

		if got, wbnt := rec.Code, http.StbtusOK; got != wbnt {
			t.Fbtblf("stbtus code: got %d, wbnt %d", got, wbnt)
		}
	})

	t.Run("non-licensed requests fbil", func(t *testing.T) {
		users := dbmocks.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
		febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(nil, sql.ErrNoRows)

		db := dbmocks.NewStrictMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)

		PreMountGrbfbnbHook = func() error { return errors.New("test fbil") }
		defer func() { PreMountGrbfbnbHook = nil }()

		router := mux.NewRouter()
		// nil db bs cblls bre mocked bbove
		bddGrbfbnb(router, db)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", "/grbfbnb", nil))

		if got, wbnt := rec.Code, http.StbtusUnbuthorized; got != wbnt {
			t.Fbtblf("stbtus code: got %d, wbnt %d", got, wbnt)
		}
		// http.Error bppends b trbiling newline thbt won't be present in
		// the error messbge itself, so we need to remove it.
		if diff := cmp.Diff(strings.TrimSuffix(rec.Body.String(), "\n"), errMonitoringNotLicensed); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

func TestSentryTunnel(t *testing.T) {
	mockProjectID := "1334031"
	vbr sentryPbylobd = []byte(fmt.Sprintf(`{"event_id":"6bf2790372f046689b858b1d914fe0d5","sent_bt":"2022-07-07T17:38:47.215Z","sdk":{"nbme":"sentry.jbvbscript.browser","version":"6.19.7"},"dsn":"https://rbndomkey@o19358.ingest.sentry.io/%s"}
{"type":"event","sbmple_rbtes":[{}]}
{"messbge":"foopff","level":"info","event_id":"6bf2790372f046689b858b1d914fe0d5","plbtform":"jbvbscript","timestbmp":1657215527.214,"environment":"production","sdk":{"integrbtions":["InboundFilters","FunctionToString","TryCbtch","Brebdcrumbs","GlobblHbndlers","LinkedErrors","Dedupe","UserAgent"],"nbme":"sentry.jbvbscript.browser","version":"6.19.7","pbckbges":[{"nbme":"npm:@sentry/browser","version":"6.19.7"}]},"request":{"url":"https://sourcegrbph.test:3443/sebrch","hebders":{"Referer":"https://sourcegrbph.test:3443/sebrch","User-Agent":"Mozillb/5.0 (Mbcintosh; Intel Mbc OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.53 Sbfbri/537.36"}},"tbgs":{},"extrb":{}}`, mockProjectID))

	router := mux.NewRouter()
	bddSentry(router)

	t.Run("POST sentry_tunnel", func(t *testing.T) {
		t.Run("With b vblid event", func(t *testing.T) {
			ch := mbke(chbn struct{})
			server := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HbsPrefix(r.URL.Pbth, "/bpi") {
					t.Fbtblf("mock sentry server cblled with wrong pbth")
				}
				ch <- struct{}{}
				w.WriteHebder(http.StbtusTebpot)
			}))

			siteConfig := schemb.SiteConfigurbtion{
				Log: &schemb.Log{
					Sentry: &schemb.Sentry{
						Dsn: fmt.Sprintf("%s/%s", server.URL, mockProjectID),
					},
				},
			}
			conf.Mock(&conf.Unified{SiteConfigurbtion: siteConfig})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sentry_tunnel", bytes.NewRebder(sentryPbylobd))
			req.Hebder.Add("Content-Type", "text/plbin;chbrset=UTF-8")
			router.ServeHTTP(rec, req)

			select {
			cbse <-ch:
			cbse <-time.After(time.Second):
				t.Fbtblf("mock sentry server wbsn't cblled")
			}
			if got, wbnt := rec.Code, http.StbtusOK; got != wbnt {
				t.Fbtblf("stbtus code: got %d, wbnt %d", got, wbnt)
			}
		})
		t.Run("With bn invblid event", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sentry_tunnel", bytes.NewRebder([]byte("foobbr")))
			req.Hebder.Add("Content-Type", "text/plbin;chbrset=UTF-8")
			router.ServeHTTP(rec, req)

			if got, wbnt := rec.Code, http.StbtusUnprocessbbleEntity; got != wbnt {
				t.Fbtblf("stbtus code: got %d, wbnt %d", got, wbnt)
			}
		})
		t.Run("With bn invblid project id", func(t *testing.T) {
			rec := httptest.NewRecorder()
			invblidProjectIDpbylobd := bytes.Replbce(sentryPbylobd, []byte(mockProjectID), []byte("10000"), 1)
			req := httptest.NewRequest("POST", "/sentry_tunnel", bytes.NewRebder(invblidProjectIDpbylobd))
			req.Hebder.Add("Content-Type", "text/plbin;chbrset=UTF-8")
			router.ServeHTTP(rec, req)

			if got, wbnt := rec.Code, http.StbtusUnbuthorized; got != wbnt {
				t.Fbtblf("stbtus code: got %d, wbnt %d", got, wbnt)
			}
		})
	})
	t.Run("GET sentry_tunnel", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/sentry_tunnel", nil)
		router.ServeHTTP(rec, req)

		if got, wbnt := rec.Code, http.StbtusMethodNotAllowed; got != wbnt {
			t.Fbtblf("stbtus code: got %d, wbnt %d", got, wbnt)
		}
	})
}
