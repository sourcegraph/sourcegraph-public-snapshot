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

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"

	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
)

var githubClientID = env.Get("GITHUB_CLIENT_ID", "", "client ID for GitHub")
var githubClientSecret = env.Get("GITHUB_CLIENT_SECRET", "", "client secret for GitHub")
var logRequests, _ = strconv.ParseBool(env.Get("LOG_REQUESTS", "", "log HTTP requests"))
var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")

var locks = make(map[string]*sync.Mutex)
var locksMu sync.Mutex

var rateLimitRemainingGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "github",
	Name:      "rate_limit_remaining",
	Help:      "Number of calls to GitHub's API remaining before hitting the rate limit.",
})

func init() {
	rateLimitRemainingGauge.Set(5000)
	prometheus.MustRegister(rateLimitRemainingGauge)
}

func main() {
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

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

	var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.URL.Query().Get("access_token")
		if auth := r.Header.Get("Authorization"); auth != "" {
			fields := strings.Fields(auth)
			if len(fields) == 2 && (fields[0] == "token" || fields[0] == "Bearer") {
				accessToken = fields[1]
			}
		}

		q2 := r.URL.Query()

		h2 := make(http.Header)
		h2.Set("User-Agent", r.Header.Get("User-Agent"))
		h2.Set("Accept", r.Header.Get("Accept"))

		if accessToken != "" {
			h2.Set("Authorization", "token "+accessToken)
		} else {
			q2.Set("client_id", githubClientID)
			q2.Set("client_secret", githubClientSecret)
		}

		req2 := &http.Request{
			Method: r.Method,
			URL: &url.URL{
				Scheme:   "https",
				Host:     "api.github.com",
				Path:     r.URL.Path,
				RawQuery: q2.Encode(),
			},
			Header: h2,
		}

		locksMu.Lock()
		lock, ok := locks[accessToken]
		if !ok {
			lock = new(sync.Mutex)
			locks[accessToken] = lock
		}
		locksMu.Unlock()

		lock.Lock()
		resp, err := http.DefaultClient.Do(req2)
		lock.Unlock()
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if accessToken == "" { // do not track user rate limits
			if limit := resp.Header.Get("X-Ratelimit-Remaining"); limit != "" {
				limit, _ := strconv.Atoi(limit)
				rateLimitRemainingGauge.Set(float64(limit))
			}
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

	log.Print("github-proxy: listening on :3180")
	log.Fatal(http.ListenAndServe(":3180", nil))
}
