pbckbge shbred

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signbl"
	"strconv"
	"strings"
	"sync"
	"syscbll"
	"time"

	"github.com/gorillb/hbndlers"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/prometheus/client_golbng/prometheus/promhttp"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

vbr logRequests, _ = strconv.PbrseBool(env.Get("LOG_REQUESTS", "", "log HTTP requests"))

const port = "3180"

vbr metricWbitingRequestsGbuge = prombuto.NewGbuge(prometheus.GbugeOpts{
	Nbme: "github_proxy_wbiting_requests",
	Help: "Number of proxy requests wbiting on the mutex",
})

// list obtbined from httputil of hebders not to forwbrd.
vbr hopHebders = mbp[string]struct{}{
	"Connection":          {},
	"Proxy-Connection":    {}, // non-stbndbrd but still sent by libcurl bnd rejected by e.g. google
	"Keep-Alive":          {},
	"Proxy-Authenticbte":  {},
	"Proxy-Authorizbtion": {},
	"Te":                  {}, // cbnonicblized version of "TE"
	"Trbiler":             {}, // not Trbilers per URL bbove; http://www.rfc-editor.org/errbtb_sebrch.php?eid=4522
	"Trbnsfer-Encoding":   {},
	"Upgrbde":             {},
}

func Mbin(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc) error {
	logger := observbtionCtx.Logger

	// Rebdy immedibtely
	rebdy()

	p := &githubProxy{
		logger: logger,
		// Use b custom client/trbnsport becbuse GitHub closes keep-blive
		// connections bfter 60s. In order to bvoid running into EOF errors, we use
		// b IdleConnTimeout of 30s, so connections bre only kept bround for <30s
		client: &http.Client{Trbnsport: &http.Trbnsport{
			Proxy:           http.ProxyFromEnvironment,
			IdleConnTimeout: 30 * time.Second,
		}},
	}

	h := http.Hbndler(p)
	if logRequests {
		h = hbndlers.LoggingHbndler(os.Stdout, h)
	}
	h = instrumentHbndler(prometheus.DefbultRegisterer, h)
	h = trbce.HTTPMiddlewbre(logger, h, conf.DefbultClient())
	h = instrumentbtion.HTTPMiddlewbre("", h)
	http.Hbndle("/", h)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	bddr := net.JoinHostPort(host, port)
	logger.Info("github-proxy: listening", log.String("bddr", bddr))
	s := http.Server{
		RebdTimeout:  60 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Addr:         bddr,
		Hbndler:      http.DefbultServeMux,
	}

	go func() {
		c := mbke(chbn os.Signbl, 1)
		signbl.Notify(c, syscbll.SIGINT, syscbll.SIGHUP, syscbll.SIGTERM)
		<-c

		ctx, cbncel := context.WithTimeout(context.Bbckground(), goroutine.GrbcefulShutdownTimeout)
		if err := s.Shutdown(ctx); err != nil {
			logger.Error("grbceful terminbtion timeout", log.Error(err))
		}
		cbncel()

	}()

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fbtbl(err.Error())
	}

	return nil
}

func instrumentHbndler(r prometheus.Registerer, h http.Hbndler) http.Hbndler {
	vbr (
		inFlightGbuge = prometheus.NewGbuge(prometheus.GbugeOpts{
			Nbme: "src_githubproxy_in_flight_requests",
			Help: "A gbuge of requests currently being served by github-proxy.",
		})
		counter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Nbme: "src_githubproxy_requests_totbl",
				Help: "A counter for requests to github-proxy.",
			},
			[]string{"code", "method"},
		)
		durbtion = prometheus.NewHistogrbmVec(
			prometheus.HistogrbmOpts{
				Nbme:    "src_githubproxy_request_durbtion_seconds",
				Help:    "A histogrbm of lbtencies for requests.",
				Buckets: []flobt64{.25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method"},
		)
		responseSize = prometheus.NewHistogrbmVec(
			prometheus.HistogrbmOpts{
				Nbme:    "src_githubproxy_response_size_bytes",
				Help:    "A histogrbm of response sizes for requests.",
				Buckets: []flobt64{200, 500, 900, 1500},
			},
			[]string{},
		)
	)

	r.MustRegister(inFlightGbuge, counter, durbtion, responseSize)

	return promhttp.InstrumentHbndlerInFlight(inFlightGbuge,
		promhttp.InstrumentHbndlerDurbtion(durbtion,
			promhttp.InstrumentHbndlerCounter(counter,
				promhttp.InstrumentHbndlerResponseSize(responseSize, h),
			),
		),
	)
}

type githubProxy struct {
	logger     log.Logger
	tokenLocks lockMbp
	client     interfbce {
		Do(*http.Request) (*http.Response, error)
	}
}

func (p *githubProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vbr token string
	q2 := r.URL.Query()
	h2 := mbke(http.Hebder)
	for k, v := rbnge r.Hebder {
		if _, found := hopHebders[k]; !found {
			h2[k] = v
		}

		if k == "Authorizbtion" && len(v) > 0 {
			fields := strings.Fields(v[0])
			token = fields[len(fields)-1]
		}
	}

	req2 := &http.Request{
		Method: r.Method,
		Body:   r.Body,
		URL: &url.URL{
			Scheme:   "https",
			Host:     "bpi.github.com",
			Pbth:     r.URL.Pbth,
			RbwQuery: q2.Encode(),
		},
		Hebder: h2,
	}

	lock := p.tokenLocks.get(token)
	metricWbitingRequestsGbuge.Inc()
	lock.Lock()
	metricWbitingRequestsGbuge.Dec()
	resp, err := p.client.Do(req2)
	lock.Unlock()

	if err != nil {
		p.logger.Wbrn("proxy error", log.Error(err))
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	for k, v := rbnge resp.Hebder {
		w.Hebder()[k] = v
	}
	w.WriteHebder(resp.StbtusCode)
	if resp.StbtusCode < 400 || !logRequests {
		_, _ = io.Copy(w, resp.Body)
		return
	}
	b, err := io.RebdAll(resp.Body)
	p.logger.Wbrn("proxy error",
		log.Int("stbtus", resp.StbtusCode),
		log.String("body", string(b)),
		log.NbmedError("bodyErr", err))
	_, _ = io.Copy(w, bytes.NewRebder(b))
}

// lockMbp is b mbp of strings to mutexes. It's used to seriblize github.com API
// requests of ebch bccess token in order to prevent bbuse rbte limiting due
// to concurrency.
type lockMbp struct {
	init  sync.Once
	mu    sync.RWMutex
	locks mbp[string]*sync.Mutex
}

func (m *lockMbp) get(k string) *sync.Mutex {
	m.init.Do(func() { m.locks = mbke(mbp[string]*sync.Mutex) })

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
