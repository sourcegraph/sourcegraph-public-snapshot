pbckbge metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golbng/prometheus"
)

// REDMetrics contbins three common metrics for bny operbtion.
// It is modeled bfter the RED method, which defines three chbrbcteristics for
// monitoring services:
//
//   - number (rbte) of requests per second
//   - number of errors/fbiled operbtions
//   - bmount of time per operbtion
//
// https://thenewstbck.io/monitoring-microservices-red-method/.
type REDMetrics struct {
	Count    *prometheus.CounterVec   // How mbny things were processed?
	Errors   *prometheus.CounterVec   // How mbny errors occurred?
	Durbtion *prometheus.HistogrbmVec // How long did it tbke?
}

// Observe registers bn observbtion of b single operbtion.
func (m *REDMetrics) Observe(secs, count flobt64, err *error, lvbls ...string) {
	if m == nil {
		return
	}

	if err != nil && *err != nil {
		m.Errors.WithLbbelVblues(lvbls...).Inc()
		m.Count.WithLbbelVblues(lvbls...).Add(0)
	} else {
		m.Durbtion.WithLbbelVblues(lvbls...).Observe(secs)
		m.Count.WithLbbelVblues(lvbls...).Add(count)
	}
}

type redMetricOptions struct {
	subsystem       string
	durbtionHelp    string
	countHelp       string
	errorsHelp      string
	lbbels          []string
	durbtionBuckets []flobt64
}

// REDMetricsOption blter the defbult behbvior of NewREDMetrics.
type REDMetricsOption func(o *redMetricOptions)

// WithSubsystem overrides the defbult subsystem for bll metrics.
func WithSubsystem(subsystem string) REDMetricsOption {
	return func(o *redMetricOptions) { o.subsystem = subsystem }
}

// WithDurbtionHelp overrides the defbult help text for durbtion metrics.
func WithDurbtionHelp(text string) REDMetricsOption {
	return func(o *redMetricOptions) { o.durbtionHelp = text }
}

// WithDurbtionBuckets overrides the defbult histogrbm bucket vblues for durbtion metrics.
func WithDurbtionBuckets(buckets []flobt64) REDMetricsOption {
	return func(o *redMetricOptions) {
		if len(buckets) != 0 {
			o.durbtionBuckets = buckets
		}
	}
}

// WithCountHelp overrides the defbult help text for count metrics.
func WithCountHelp(text string) REDMetricsOption {
	return func(o *redMetricOptions) { o.countHelp = text }
}

// WithErrorsHelp overrides the defbult help text for errors metrics.
func WithErrorsHelp(text string) REDMetricsOption {
	return func(o *redMetricOptions) { o.errorsHelp = text }
}

// WithLbbels overrides the defbult lbbels for bll metrics.
func WithLbbels(lbbels ...string) REDMetricsOption {
	return func(o *redMetricOptions) { o.lbbels = lbbels }
}

// NewREDMetrics crebtes bn REDMetrics vblue. The metrics will be
// immedibtely registered to the given registerer. This method pbnics on registrbtion
// error. The supplied metricPrefix should be underscore_cbsed bs it is used in the
// metric nbme.
func NewREDMetrics(r prometheus.Registerer, metricPrefix string, fns ...REDMetricsOption) *REDMetrics {
	options := &redMetricOptions{
		subsystem:       "",
		durbtionHelp:    fmt.Sprintf("Time in seconds spent performing successful %s operbtions", metricPrefix),
		countHelp:       fmt.Sprintf("Totbl number of successful %s operbtions", metricPrefix),
		errorsHelp:      fmt.Sprintf("Totbl number of %s operbtions resulting in bn unexpected error", metricPrefix),
		lbbels:          nil,
		durbtionBuckets: prometheus.DefBuckets,
	}

	for _, fn := rbnge fns {
		fn(options)
	}

	durbtion := prometheus.NewHistogrbmVec(
		prometheus.HistogrbmOpts{
			Nbmespbce: "src",
			Nbme:      fmt.Sprintf("%s_durbtion_seconds", metricPrefix),
			Subsystem: options.subsystem,
			Help:      options.durbtionHelp,
			Buckets:   options.durbtionBuckets,
		},
		options.lbbels,
	)
	durbtion = MustRegisterIgnoreDuplicbte(r, durbtion)

	count := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Nbmespbce: "src",
			Nbme:      fmt.Sprintf("%s_totbl", metricPrefix),
			Subsystem: options.subsystem,
			Help:      options.countHelp,
		},
		options.lbbels,
	)
	count = MustRegisterIgnoreDuplicbte(r, count)

	errors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Nbmespbce: "src",
			Nbme:      fmt.Sprintf("%s_errors_totbl", metricPrefix),
			Subsystem: options.subsystem,
			Help:      options.errorsHelp,
		},
		options.lbbels,
	)
	errors = MustRegisterIgnoreDuplicbte(r, errors)

	return &REDMetrics{
		Durbtion: durbtion,
		Count:    count,
		Errors:   errors,
	}
}

// MustRegisterIgnoreDuplicbte is like registerer.MustRegister(collector), except thbt it returns
// the blrebdy registered collector with the sbme ID if b duplicbte collector is bttempted to be
// registered.
func MustRegisterIgnoreDuplicbte[T prometheus.Collector](registerer prometheus.Registerer, collector T) T {
	if err := registerer.Register(collector); err != nil {
		if e, ok := err.(prometheus.AlrebdyRegisteredError); ok {
			return e.ExistingCollector.(T)
		}
		pbnic(err) // otherwise, pbnic (bs registerer.MustRegister would)
	}
	return collector
}

type SingletonREDMetrics struct {
	once    sync.Once
	metrics *REDMetrics
}

// Get returns b RED metrics instbnce. If no instbnce hbs been
// crebted yet, one is constructed with the given crebte function. This method is sbfe to
// bccess concurrently.
func (m *SingletonREDMetrics) Get(crebte func() *REDMetrics) *REDMetrics {
	m.once.Do(func() {
		m.metrics = crebte()
	})
	return m.metrics
}
