package app

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_prometheusValidator(t *testing.T) {
	type args struct {
		prometheusURL string
		config        conf.Unified
	}
	tests := []struct {
		name                 string
		args                 args
		wantProblemSubstring string
	}{
		{
			name: "no problem if prometheus not set",
			args: args{
				prometheusURL: "",
			},
			wantProblemSubstring: "",
		},
		{
			name: "no problem if no alerts set",
			args: args{
				prometheusURL: "http://prometheus:9090",
				config:        conf.Unified{},
			},
			wantProblemSubstring: "",
		},
		{
			name: "url and alerts set, but malformed prometheus URL",
			args: args{
				prometheusURL: " http://prometheus:9090",
				config: conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						ObservabilityAlerts: []*schema.ObservabilityAlerts{{
							Level: "critical",
						}},
					},
				},
			},
			wantProblemSubstring: "misconfigured",
		},
		{
			name: "prometheus not found (with only observability.alerts configured)",
			args: args{
				prometheusURL: "http://no-prometheus:9090",
				config: conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						ObservabilityAlerts: []*schema.ObservabilityAlerts{{
							Level: "critical",
						}},
					},
				},
			},
			wantProblemSubstring: "failed to fetch alerting configuration",
		},
		{
			name: "prometheus not found (with only observability.silenceAlerts configured)",
			args: args{
				prometheusURL: "http://no-prometheus:9090",
				config: conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						ObservabilitySilenceAlerts: []string{"warning_gitserver_disk_space_remaining"},
					},
				},
			},
			wantProblemSubstring: "failed to fetch alerting configuration",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := newPrometheusValidator(srcprometheus.NewClient(tt.args.prometheusURL))
			problems := validate(tt.args.config)
			if tt.wantProblemSubstring == "" {
				if len(problems) > 0 {
					t.Errorf("expected no problems, got %+v", problems)
				}
			} else {
				found := false
				for _, p := range problems {
					if strings.Contains(p.String(), tt.wantProblemSubstring) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected problem '%s', got %+v", tt.wantProblemSubstring, problems)
				}
			}
		})
	}
}

func TestGrafanaLicensing(t *testing.T) {
	t.Run("licensed requests succeed", func(t *testing.T) {
		users := dbmocks.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		featureFlags := dbmocks.NewMockFeatureFlagStore()
		featureFlags.GetFeatureFlagFunc.SetDefaultReturn(nil, sql.ErrNoRows)

		db := dbmocks.NewStrictMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

		PreMountGrafanaHook = func() error { return nil }
		defer func() { PreMountGrafanaHook = nil }()

		router := mux.NewRouter()
		addGrafana(router, db)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", "/grafana", nil))

		if got, want := rec.Code, http.StatusOK; got != want {
			t.Fatalf("status code: got %d, want %d", got, want)
		}
	})

	t.Run("non-licensed requests fail", func(t *testing.T) {
		users := dbmocks.NewStrictMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		featureFlags := dbmocks.NewMockFeatureFlagStore()
		featureFlags.GetFeatureFlagFunc.SetDefaultReturn(nil, sql.ErrNoRows)

		db := dbmocks.NewStrictMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

		PreMountGrafanaHook = func() error { return errors.New("test fail") }
		defer func() { PreMountGrafanaHook = nil }()

		router := mux.NewRouter()
		// nil db as calls are mocked above
		addGrafana(router, db)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", "/grafana", nil))

		if got, want := rec.Code, http.StatusUnauthorized; got != want {
			t.Fatalf("status code: got %d, want %d", got, want)
		}
		// http.Error appends a trailing newline that won't be present in
		// the error message itself, so we need to remove it.
		if diff := cmp.Diff(strings.TrimSuffix(rec.Body.String(), "\n"), errMonitoringNotLicensed); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestSentryTunnel(t *testing.T) {
	mockProjectID := "1334031"
	var sentryPayload = []byte(fmt.Sprintf(`{"event_id":"6af2790372f046689a858b1d914fe0d5","sent_at":"2022-07-07T17:38:47.215Z","sdk":{"name":"sentry.javascript.browser","version":"6.19.7"},"dsn":"https://randomkey@o19358.ingest.sentry.io/%s"}
{"type":"event","sample_rates":[{}]}
{"message":"foopff","level":"info","event_id":"6af2790372f046689a858b1d914fe0d5","platform":"javascript","timestamp":1657215527.214,"environment":"production","sdk":{"integrations":["InboundFilters","FunctionToString","TryCatch","Breadcrumbs","GlobalHandlers","LinkedErrors","Dedupe","UserAgent"],"name":"sentry.javascript.browser","version":"6.19.7","packages":[{"name":"npm:@sentry/browser","version":"6.19.7"}]},"request":{"url":"https://sourcegraph.test:3443/search","headers":{"Referer":"https://sourcegraph.test:3443/search","User-Agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.53 Safari/537.36"}},"tags":{},"extra":{}}`, mockProjectID))

	router := mux.NewRouter()
	addSentry(router)

	t.Run("POST sentry_tunnel", func(t *testing.T) {
		t.Run("With a valid event", func(t *testing.T) {
			ch := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, "/api") {
					t.Fatalf("mock sentry server called with wrong path")
				}
				ch <- struct{}{}
				w.WriteHeader(http.StatusTeapot)
			}))

			siteConfig := schema.SiteConfiguration{
				Log: &schema.Log{
					Sentry: &schema.Sentry{
						Dsn: fmt.Sprintf("%s/%s", server.URL, mockProjectID),
					},
				},
			}
			conf.Mock(&conf.Unified{SiteConfiguration: siteConfig})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sentry_tunnel", bytes.NewReader(sentryPayload))
			req.Header.Add("Content-Type", "text/plain;charset=UTF-8")
			router.ServeHTTP(rec, req)

			select {
			case <-ch:
			case <-time.After(time.Second):
				t.Fatalf("mock sentry server wasn't called")
			}
			if got, want := rec.Code, http.StatusOK; got != want {
				t.Fatalf("status code: got %d, want %d", got, want)
			}
		})
		t.Run("With an invalid event", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sentry_tunnel", bytes.NewReader([]byte("foobar")))
			req.Header.Add("Content-Type", "text/plain;charset=UTF-8")
			router.ServeHTTP(rec, req)

			if got, want := rec.Code, http.StatusUnprocessableEntity; got != want {
				t.Fatalf("status code: got %d, want %d", got, want)
			}
		})
		t.Run("With an invalid project id", func(t *testing.T) {
			rec := httptest.NewRecorder()
			invalidProjectIDpayload := bytes.Replace(sentryPayload, []byte(mockProjectID), []byte("10000"), 1)
			req := httptest.NewRequest("POST", "/sentry_tunnel", bytes.NewReader(invalidProjectIDpayload))
			req.Header.Add("Content-Type", "text/plain;charset=UTF-8")
			router.ServeHTTP(rec, req)

			if got, want := rec.Code, http.StatusUnauthorized; got != want {
				t.Fatalf("status code: got %d, want %d", got, want)
			}
		})
	})
	t.Run("GET sentry_tunnel", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/sentry_tunnel", nil)
		router.ServeHTTP(rec, req)

		if got, want := rec.Code, http.StatusMethodNotAllowed; got != want {
			t.Fatalf("status code: got %d, want %d", got, want)
		}
	})
}
