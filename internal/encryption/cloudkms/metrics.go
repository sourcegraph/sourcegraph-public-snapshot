package cloudkms

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cryptographicTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "src_cloudkms_cryptographic_total",
			Help: "Total number of Cloud KMS cryptographic requests that have been sent",
		},
		[]string{"operation", "success"},
	)
)
