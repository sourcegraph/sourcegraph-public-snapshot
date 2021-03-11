package debugserver

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

func registerMetadataGauge() {
	promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_service_metadata",
		Help: "A metric with constant '1' value labelled with Sourcegraph service metadata.",
	}, []string{
		"version",
	}).With(prometheus.Labels{
		"version": version.Version(),
	}).Set(1)
}
