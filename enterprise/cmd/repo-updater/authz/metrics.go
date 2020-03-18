package authz

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	usersWithNoPerms = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "authz_no_perms_users",
		Help:      "The number of users who do not have any permissions",
	})
	reposWithNoPerms = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "authz_no_perms_repos",
		Help:      "The number of repositories that do not have any permissions",
	})

	permsSyncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "authz_perms_sync_errors_total",
		Help:      "Total number of permissions sync errors",
	}, []string{})
	permsSyncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "authz_perms_sync_duration_seconds",
		Help:      "Time spent on permissions sync",
		Buckets:   []float64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"success"})
)
