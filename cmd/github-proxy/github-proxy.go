package main

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

var logRequests, _ = strconv.ParseBool(env.Get("LOG_REQUESTS", "", "log HTTP requests"))

const port = "3180"

var metricWaitingRequestsGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "github_proxy_waiting_requests",
	Help: "Number of proxy requests waiting on the mutex",
})

// list obtained from httputil of headers not to forward.
var hopHeaders = map[string]struct{}{
	"Connection":          {},
	"Proxy-Connection":    {}, // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Te":                  {}, // canonicalized version of "TE"
	"Trailer":             {}, // not Trailers per URL above; http://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func main() {
	env.Lock()
	env.HandleHelpFlag()

	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	}, log.NewSentrySinkWith(
		log.SentrySink{
			ClientOptions: sentry.ClientOptions{SampleRate: 0.2},
		},
	)) // Experimental: DevX is observing how sampling affects the errors signal

	defer liblog.Sync()
	conf.Init()
	go conf.Watch(liblog.Update(conf.GetLogSinks))
	tracer.Init(log.Scoped("tracer", "internal tracer package"), conf.DefaultClient())
	trace.Init()

	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()

	logger := log.Scoped("server", "the github-proxy service")

	p := &githubProxy{
		logger: logger,
		// Use a custom client/transport because GitHub closes keep-alive
		// connections after 60s. In order to avoid running into EOF errors, we use
		// a IdleConnTimeout of 30s, so connections are only kept around for <30s
		client: &http.Client{Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			IdleConnTimeout: 30 * time.Second,
		}},
	}

	h := http.Handler(p)
	if logRequests {
		h = handlers.LoggingHandler(os.Stdout, h)
	}
	h = instrumentHandler(prometheus.DefaultRegisterer, h)
	h = trace.HTTPMiddleware(logger, h, conf.DefaultClient())
	h = instrumentation.HTTPMiddleware("", h)
	http.Handle("/", h)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	logger.Info("github-proxy: listening", log.String("addr", addr))
	s := http.Server{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Addr:         addr,
		Handler:      http.DefaultServeMux,
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		<-c

		ctx, cancel := context.WithTimeout(context.Background(), goroutine.GracefulShutdownTimeout)
		if err := s.Shutdown(ctx); err != nil {
			logger.Error("graceful termination timeout", log.Error(err))
		}
		cancel()

		os.Exit(0)
	}()

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err.Error())
	}
}

func instrumentHandler(r prometheus.Registerer, h http.Handler) http.Handler {
	var (
		inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "src_githubproxy_in_flight_requests",
			Help: "A gauge of requests currently being served by github-proxy.",
		})
		counter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "src_githubproxy_requests_total",
				Help: "A counter for requests to github-proxy.",
			},
			[]string{"code", "method"},
		)
		duration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "src_githubproxy_request_duration_seconds",
				Help:    "A histogram of latencies for requests.",
				Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method"},
		)
		responseSize = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "src_githubproxy_response_size_bytes",
				Help:    "A histogram of response sizes for requests.",
				Buckets: []float64{200, 500, 900, 1500},
			},
			[]string{},
		)
	)

	r.MustRegister(inFlightGauge, counter, duration, responseSize)

	return promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(duration,
			promhttp.InstrumentHandlerCounter(counter,
				promhttp.InstrumentHandlerResponseSize(responseSize, h),
			),
		),
	)
}

type githubProxy struct {
	logger     log.Logger
	tokenLocks lockMap
	client     interface {
		Do(*http.Request) (*http.Response, error)
	}
}

func (p *githubProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var token string
	q2 := r.URL.Query()
	h2 := make(http.Header)
	for k, v := range r.Header {
		if _, found := hopHeaders[k]; !found {
			h2[k] = v
		}

		if k == "Authorization" && len(v) > 0 {
			fields := strings.Fields(v[0])
			token = fields[len(fields)-1]
		}
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

	lock := p.tokenLocks.get(token)
	metricWaitingRequestsGauge.Inc()
	lock.Lock()
	metricWaitingRequestsGauge.Dec()
	resp, err := p.client.Do(req2)
	lock.Unlock()

	if err != nil {
		p.logger.Warn("proxy error", log.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	if resp.StatusCode < 400 || !logRequests {
		_, _ = io.Copy(w, resp.Body)
		return
	}
	b, err := io.ReadAll(resp.Body)
	p.logger.Warn("proxy error",
		log.Int("status", resp.StatusCode),
		log.String("body", string(b)),
		log.NamedError("bodyErr", err))
	_, _ = io.Copy(w, bytes.NewReader(b))
}

// lockMap is a map of strings to mutexes. It's used to serialize github.com API
// requests of each access token in order to prevent abuse rate limiting due
// to concurrency.
type lockMap struct {
	init  sync.Once
	mu    sync.RWMutex
	locks map[string]*sync.Mutex
}

func (m *lockMap) get(k string) *sync.Mutex {
	m.init.Do(func() { m.locks = make(map[string]*sync.Mutex) })

	m.mu.RLock()
	lock, ok := m.locks[k]
	m.mu.RUnlock()

	if ok {
		return lock
	}

	m.mu.Lock()
	lock, ok = m.locks[k]
	if !ok {
		lock = &sync.Mutex{}
		m.locks[k] = lock
	}
	m.mu.Unlock()

	return lock
}
