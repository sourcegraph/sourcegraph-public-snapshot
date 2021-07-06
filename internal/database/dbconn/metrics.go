package dbconn

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

// metricsCollector implements the Prometheus collector interface.
// It reports all metrics returned by sql.DB.Stats().
// Adapted from github.com/dlmiddlecote/sqlstats
type metricsCollector struct {
	db *sql.DB

	// descriptions of exported metrics
	maxOpenDesc           *prometheus.Desc
	openDesc              *prometheus.Desc
	inUseDesc             *prometheus.Desc
	idleDesc              *prometheus.Desc
	waitedForDesc         *prometheus.Desc
	closedMaxIdleDesc     *prometheus.Desc
	closedMaxLifetimeDesc *prometheus.Desc
	closedMaxIdleTimeDesc *prometheus.Desc
	blockedSecondsHist    prometheus.Histogram
}

func newMetricsCollector(db *sql.DB, dbname, app string) *metricsCollector {
	const (
		namespace = "src"
		subsystem = "pgsql_conns"
	)

	labels := prometheus.Labels{
		"db_name":  dbname,
		"app_name": app,
	}

	return &metricsCollector{
		db: db,
		maxOpenDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_open"),
			"Maximum number of open connections to the database.",
			nil,
			labels,
		),
		openDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "open"),
			"The number of established connections both in use and idle.",
			nil,
			labels,
		),
		inUseDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "in_use"),
			"The number of connections currently in use.",
			nil,
			labels,
		),
		idleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "idle"),
			"The number of idle connections.",
			nil,
			labels,
		),
		waitedForDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "waited_for"),
			"The total number of connections waited for.",
			nil,
			labels,
		),
		closedMaxIdleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "closed_max_idle"),
			"The total number of connections closed due to SetMaxIdleConns.",
			nil,
			labels,
		),
		closedMaxLifetimeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "closed_max_lifetime"),
			"The total number of connections closed due to SetConnMaxLifetime.",
			nil,
			labels,
		),
		closedMaxIdleTimeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "closed_max_idle_time"),
			"The total number of connections closed due to SetConnMaxIdleTime.",
			nil,
			labels,
		),
		blockedSecondsHist: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "blocked_seconds",
			Help:        "The total time blocked waiting for a new connection.",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		}),
	}
}

// Describe implements the prometheus.Collector interface.
func (c metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpenDesc
	ch <- c.openDesc
	ch <- c.inUseDesc
	ch <- c.idleDesc
	ch <- c.waitedForDesc
	ch <- c.closedMaxIdleDesc
	ch <- c.closedMaxLifetimeDesc
	ch <- c.closedMaxIdleTimeDesc
	c.blockedSecondsHist.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (c metricsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.db.Stats()

	ch <- prometheus.MustNewConstMetric(
		c.maxOpenDesc,
		prometheus.GaugeValue,
		float64(stats.MaxOpenConnections),
	)
	ch <- prometheus.MustNewConstMetric(
		c.openDesc,
		prometheus.GaugeValue,
		float64(stats.OpenConnections),
	)
	ch <- prometheus.MustNewConstMetric(
		c.inUseDesc,
		prometheus.GaugeValue,
		float64(stats.InUse),
	)
	ch <- prometheus.MustNewConstMetric(
		c.idleDesc,
		prometheus.GaugeValue,
		float64(stats.Idle),
	)
	ch <- prometheus.MustNewConstMetric(
		c.waitedForDesc,
		prometheus.CounterValue,
		float64(stats.WaitCount),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxIdleDesc,
		prometheus.CounterValue,
		float64(stats.MaxIdleClosed),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxLifetimeDesc,
		prometheus.CounterValue,
		float64(stats.MaxLifetimeClosed),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxIdleTimeDesc,
		prometheus.CounterValue,
		float64(stats.MaxIdleTimeClosed),
	)

	c.blockedSecondsHist.Observe(stats.WaitDuration.Seconds())
	c.blockedSecondsHist.Collect(ch)
}
