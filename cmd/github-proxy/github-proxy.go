//docker:user sourcegraph

package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"

	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
)

var logRequests, _ = strconv.ParseBool(env.Get("LOG_REQUESTS", "", "log HTTP requests"))
var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")

// requestMu ensures we only do one request at a time to prevent tripping abuse detection.
var requestMu sync.Mutex

var rateLimitRemainingGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "github",
	Name:      "rate_limit_remaining",
	Help:      "Number of calls to GitHub's API remaining before hitting the rate limit.",
}, []string{"resource"})

func init() {
	rateLimitRemainingGauge.WithLabelValues("core").Set(5000)
	rateLimitRemainingGauge.WithLabelValues("search").Set(30)
	prometheus.MustRegister(rateLimitRemainingGauge)
}

func main() {
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init("github-proxy")

	// Filter log output by level.
	if lvl, err := log15.LvlFromString(env.LogLevel); err == nil {
		log15.Root().SetHandler(log15.LvlFilterHandler(lvl, log15.StderrHandler))
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c
		os.Exit(0)
	}()

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		log.Printf("Profiler available on %s/pprof", profBindAddr)
	}

	var authenticateRequest func(query url.Values, header http.Header)
	if clientID, clientSecret := conf.GetTODO().GithubClientID, conf.GetTODO().GithubClientSecret; clientID != "" && clientSecret != "" {
		authenticateRequest = func(query url.Values, header http.Header) {
			query.Set("client_id", clientID)
			query.Set("client_secret", clientSecret)
		}
	}

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q2 := r.URL.Query()

		h2 := make(http.Header)
		h2.Set("User-Agent", r.Header.Get("User-Agent"))
		h2.Set("Accept", r.Header.Get("Accept"))
		h2.Set("Content-Type", r.Header.Get("Content-Type"))
		if r.Header.Get("Authorization") != "" {
			h2.Set("Authorization", r.Header.Get("Authorization"))
		}

		// Authenticate for higher rate limits.
		if authenticateRequest != nil {
			authenticateRequest(q2, h2)
		}

		req2 := &http.Request{
			Method: r.Method,
			Body:   r.Body,
			URL: &url.URL{
				Scheme:   "https",
				Host:     "api.github.com",
				Path:     r.URL.Path,
				RawQuery: q2.Encode(),
			},
			Header: h2,
		}

		requestMu.Lock()
		resp, err := http.DefaultClient.Do(req2)
		requestMu.Unlock()
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if limit := resp.Header.Get("X-Ratelimit-Remaining"); limit != "" {
			limit, _ := strconv.Atoi(limit)
			resource := "core"
			if strings.HasPrefix(r.URL.Path, "/search/") {
				resource = "search"
			} else if r.URL.Path == "/graphql" {
				resource = "graphql"
			}
			rateLimitRemainingGauge.WithLabelValues(resource).Set(float64(limit))
		}

		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		resp.Body.Close()
	})
	if logRequests {
		h = handlers.LoggingHandler(os.Stdout, h)
	}
	h = prometheus.InstrumentHandler("github-proxy", h)
	http.Handle("/", h)

	log15.Info("github-proxy: listening", "addr", ":3180")
	log.Fatal(http.ListenAndServe(":3180", nil))
}
