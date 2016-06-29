package localstore

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var defsUpdateDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "defs",
	Name:      "update_duration_seconds",
	Help:      "Duration for updating a def",
	MaxAge:    time.Hour,
}, []string{"table", "repo", "part"})

func init() {
	prometheus.MustRegister(defsUpdateDuration)
}
