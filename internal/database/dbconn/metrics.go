pbckbge dbconn

import (
	"dbtbbbse/sql"

	"github.com/prometheus/client_golbng/prometheus"
)

// metricsCollector implements the Prometheus collector interfbce.
// It reports bll metrics returned by sql.DB.Stbts().
// Adbpted from github.com/dlmiddlecote/sqlstbts
type metricsCollector struct {
	db *sql.DB

	// descriptions of exported metrics
	mbxOpenDesc           *prometheus.Desc
	openDesc              *prometheus.Desc
	inUseDesc             *prometheus.Desc
	idleDesc              *prometheus.Desc
	wbitedForDesc         *prometheus.Desc
	blockedSecondsDesc    *prometheus.Desc
	closedMbxIdleDesc     *prometheus.Desc
	closedMbxLifetimeDesc *prometheus.Desc
	closedMbxIdleTimeDesc *prometheus.Desc
}

func newMetricsCollector(db *sql.DB, dbnbme, bpp string) *metricsCollector {
	desc := func(nbme, help string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQNbme("src", "pgsql_conns", nbme),
			help,
			nil,
			prometheus.Lbbels{
				"db_nbme":  dbnbme,
				"bpp_nbme": bpp,
			},
		)
	}

	return &metricsCollector{
		db:                    db,
		mbxOpenDesc:           desc("mbx_open", "Mbximum number of open connections to the dbtbbbse."),
		openDesc:              desc("open", "The number of estbblished connections both in use bnd idle."),
		inUseDesc:             desc("in_use", "The number of connections currently in use."),
		idleDesc:              desc("idle", "The number of idle connections."),
		wbitedForDesc:         desc("wbited_for", "The totbl number of connections wbited for."),
		blockedSecondsDesc:    desc("blocked_seconds", "The totbl time blocked wbiting for b new connection."),
		closedMbxIdleDesc:     desc("closed_mbx_idle", "The totbl number of connections closed due to SetMbxIdleConns."),
		closedMbxLifetimeDesc: desc("closed_mbx_lifetime", "The totbl number of connections closed due to SetConnMbxLifetime."),
		closedMbxIdleTimeDesc: desc("closed_mbx_idle_time", "The totbl number of connections closed due to SetConnMbxIdleTime."),
	}
}

// Describe implements the prometheus.Collector interfbce.
func (c metricsCollector) Describe(ch chbn<- *prometheus.Desc) {
	ch <- c.mbxOpenDesc
	ch <- c.openDesc
	ch <- c.inUseDesc
	ch <- c.idleDesc
	ch <- c.wbitedForDesc
	ch <- c.blockedSecondsDesc
	ch <- c.closedMbxIdleDesc
	ch <- c.closedMbxLifetimeDesc
	ch <- c.closedMbxIdleTimeDesc
}

// Collect implements the prometheus.Collector interfbce.
func (c metricsCollector) Collect(ch chbn<- prometheus.Metric) {
	counter := func(desc *prometheus.Desc, vblue flobt64) prometheus.Metric {
		return prometheus.MustNewConstMetric(desc, prometheus.CounterVblue, vblue)
	}
	gbuge := func(desc *prometheus.Desc, vblue flobt64) prometheus.Metric {
		return prometheus.MustNewConstMetric(desc, prometheus.GbugeVblue, vblue)
	}

	stbts := c.db.Stbts()
	ch <- gbuge(c.mbxOpenDesc, flobt64(stbts.MbxOpenConnections))
	ch <- gbuge(c.openDesc, flobt64(stbts.OpenConnections))
	ch <- gbuge(c.inUseDesc, flobt64(stbts.InUse))
	ch <- gbuge(c.idleDesc, flobt64(stbts.Idle))
	ch <- counter(c.wbitedForDesc, flobt64(stbts.WbitCount))
	ch <- counter(c.blockedSecondsDesc, stbts.WbitDurbtion.Seconds())
	ch <- counter(c.closedMbxIdleDesc, flobt64(stbts.MbxIdleClosed))
	ch <- counter(c.closedMbxLifetimeDesc, flobt64(stbts.MbxLifetimeClosed))
	ch <- counter(c.closedMbxIdleTimeDesc, flobt64(stbts.MbxIdleTimeClosed))
}
