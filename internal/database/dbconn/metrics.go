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
	blockedSecondsDesc    *prometheus.Desc
	closedMaxIdleDesc     *prometheus.Desc
	closedMaxLifetimeDesc *prometheus.Desc
	closedMaxIdleTimeDesc *prometheus.Desc
}

func newMetricsCollector(db *sql.DB, dbname, app string) *metricsCollector {
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName("src", "pgsql_conns", name),
			help,
			nil,
			prometheus.Labels{
				"db_name":  dbname,
				"app_name": app,
			},
		)
	}

	return &metricsCollector{
		db:                    db,
		maxOpenDesc:           desc("max_open", "Maximum number of open connections to the database."),
		openDesc:              desc("open", "The number of established connections both in use and idle."),
		inUseDesc:             desc("in_use", "The number of connections currently in use."),
		idleDesc:              desc("idle", "The number of idle connections."),
		waitedForDesc:         desc("waited_for", "The total number of connections waited for."),
		blockedSecondsDesc:    desc("blocked_seconds", "The total time blocked waiting for a new connection."),
		closedMaxIdleDesc:     desc("closed_max_idle", "The total number of connections closed due to SetMaxIdleConns."),
		closedMaxLifetimeDesc: desc("closed_max_lifetime", "The total number of connections closed due to SetConnMaxLifetime."),
		closedMaxIdleTimeDesc: desc("closed_max_idle_time", "The total number of connections closed due to SetConnMaxIdleTime."),
	}
}

// Describe implements the prometheus.Collector interface.
func (c metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpenDesc
	ch <- c.openDesc
	ch <- c.inUseDesc
	ch <- c.idleDesc
	ch <- c.waitedForDesc
	ch <- c.blockedSecondsDesc
	ch <- c.closedMaxIdleDesc
	ch <- c.closedMaxLifetimeDesc
	ch <- c.closedMaxIdleTimeDesc
}

// Collect implements the prometheus.Collector interface.
func (c metricsCollector) Collect(ch chan<- prometheus.Metric) {
	counter := func(desc *prometheus.Desc, value float64) prometheus.Metric {
		return prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value)
	}
	gauge := func(desc *prometheus.Desc, value float64) prometheus.Metric {
		return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value)
	}

	stats := c.db.Stats()
	ch <- gauge(c.maxOpenDesc, float64(stats.MaxOpenConnections))
	ch <- gauge(c.openDesc, float64(stats.OpenConnections))
	ch <- gauge(c.inUseDesc, float64(stats.InUse))
	ch <- gauge(c.idleDesc, float64(stats.Idle))
	ch <- counter(c.waitedForDesc, float64(stats.WaitCount))
	ch <- counter(c.blockedSecondsDesc, stats.WaitDuration.Seconds())
	ch <- counter(c.closedMaxIdleDesc, float64(stats.MaxIdleClosed))
	ch <- counter(c.closedMaxLifetimeDesc, float64(stats.MaxLifetimeClosed))
	ch <- counter(c.closedMaxIdleTimeDesc, float64(stats.MaxIdleTimeClosed))
}
