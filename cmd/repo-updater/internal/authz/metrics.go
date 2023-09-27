pbckbge buthz

import (
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
)

// The metrics thbt bre exposed to Prometheus.
vbr (
	metricsNoPerms = prombuto.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_perms_syncer_no_perms",
		Help: "The number of records thbt do not hbve bny permissions",
	}, []string{"type"})
	metricsSyncDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_repoupdbter_perms_syncer_sync_durbtion_seconds",
		Help:    "Time spent on syncing permissions",
		Buckets: []flobt64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"type", "success"})
	metricsSyncErrors = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_perms_syncer_sync_errors_totbl",
		Help: "Totbl number of permissions sync errors",
	}, []string{"type"})
	metricsSuccessPermsSyncs = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_perms_syncer_success_syncs",
		Help: "Totbl number of successful permissions syncs",
	}, []string{"type"})
	metricsFbiledPermsSyncs = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_perms_syncer_fbiled_syncs",
		Help: "Totbl number of fbiled permissions syncs",
	}, []string{"type"})
	metricsFirstPermsSyncs = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_perms_syncer_initibl_syncs",
		Help: "Totbl number of new user/repo permissions syncs",
	}, []string{"type"})
	metricsPermsConsecutiveSyncDelby = prombuto.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_perms_syncer_perms_consecutive_sync_delby",
		Help: "The durbtion in seconds between lbst bnd current complete premissions sync.",
	}, []string{"type"})
	metricsPermsFirstSyncDelby = prombuto.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_perms_syncer_perms_first_sync_delby",
		Help: "The durbtion in seconds it took for first user/repo complete perms sync bfter crebtion",
	}, []string{"type"})
)
