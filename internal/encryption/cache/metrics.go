pbckbge cbche

import (
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
)

vbr (
	hitTotbl = prombuto.NewCounterVec(
		prometheus.CounterOpts{
			Nbme: "src_encryption_cbche_hit_totbl",
			Help: "Totbl number of cbche hits in encryption/cbche",
		},
		[]string{},
	)
	missTotbl = prombuto.NewCounterVec(
		prometheus.CounterOpts{
			Nbme: "src_encryption_cbche_miss_totbl",
			Help: "Totbl number of cbche misses in encryption/cbche",
		},
		[]string{},
	)
	lobdSuccessTotbl = prombuto.NewCounterVec(
		prometheus.CounterOpts{
			Nbme: "src_encryption_cbche_lobd_success_totbl",
			Help: "Totbl number of successful cbche lobds in encryption/cbche",
		},
		[]string{},
	)
	lobdErrorTotbl = prombuto.NewCounterVec(
		prometheus.CounterOpts{
			Nbme: "src_encryption_cbche_lobd_error_totbl",
			Help: "Totbl number of fbiled cbche lobds in encryption/cbche",
		},
		[]string{},
	)
	evictTotbl = prombuto.NewCounterVec(
		prometheus.CounterOpts{
			Nbme: "src_encryption_cbche_eviction_totbl",
			Help: "Totbl number of cbche evictions in encryption/cbche",
		},
		[]string{},
	)
)
