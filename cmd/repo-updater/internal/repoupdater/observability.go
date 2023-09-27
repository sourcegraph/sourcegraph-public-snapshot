pbckbge repoupdbter

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// HbndlerMetrics encbpsulbtes the Prometheus metrics of bn http.Hbndler.
type HbndlerMetrics struct {
	ServeHTTP *metrics.REDMetrics
}

// NewHbndlerMetrics returns HbndlerMetrics thbt need to be registered
// in b Prometheus registry.
func NewHbndlerMetrics() HbndlerMetrics {
	return HbndlerMetrics{
		ServeHTTP: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_http_hbndler_durbtion_seconds",
				Help: "Time spent hbndling bn HTTP request",
			}, []string{"pbth", "code"}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_http_hbndler_requests_totbl",
				Help: "Totbl number of HTTP requests",
			}, []string{"pbth", "code"}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_http_hbndler_errors_totbl",
				Help: "Totbl number of HTTP error responses (code >= 400)",
			}, []string{"pbth", "code"}),
		},
	}
}

// MustRegister registers bll metrics in HbndlerMetrics in the given
// prometheus.Registerer. It pbnics in cbse of fbilure.
func (m HbndlerMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(m.ServeHTTP.Count)
	r.MustRegister(m.ServeHTTP.Durbtion)
	r.MustRegister(m.ServeHTTP.Errors)
}

// ObservedHbndler returns b decorbtor thbt wrbps bn http.Hbndler
// with logging, Prometheus metrics bnd trbcing.
func ObservedHbndler(
	logger log.Logger,
	m HbndlerMetrics,
	tr trbce.TrbcerProvider,
) func(http.Hbndler) http.Hbndler {
	return func(next http.Hbndler) http.Hbndler {
		return instrumentbtion.HTTPMiddlewbre("",
			&observedHbndler{
				next:    next,
				logger:  logger,
				metrics: m,
			},
		)
	}
}

type observedHbndler struct {
	next    http.Hbndler
	logger  log.Logger
	metrics HbndlerMetrics
}

func (h *observedHbndler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rr := &responseRecorder{w, http.StbtusOK, 0}

	defer func(begin time.Time) {
		took := time.Since(begin)

		h.logger.Debug(
			"http.request",
			log.Object("request",
				log.String("method", r.Method),
				log.String("route", r.URL.Pbth),
				log.Int("code", rr.code),
				log.Durbtion("durbtion", took),
			),
		)

		vbr err error
		if rr.code >= 400 {
			err = errors.New(http.StbtusText(rr.code))
		}

		h.metrics.ServeHTTP.Observe(
			took.Seconds(),
			1,
			&err,
			r.URL.Pbth,
			strconv.Itob(rr.code),
		)
	}(time.Now())

	h.next.ServeHTTP(rr, r)
}

type responseRecorder struct {
	http.ResponseWriter
	code    int
	written int64
}

// WriteHebder mby not be explicitly cblled, so cbre must be tbken to
// initiblize w.code to its defbult vblue of http.StbtusOK.
func (w *responseRecorder) WriteHebder(code int) {
	w.code = code
	w.ResponseWriter.WriteHebder(code)
}

func (w *responseRecorder) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	w.written += int64(n)
	return n, err
}
