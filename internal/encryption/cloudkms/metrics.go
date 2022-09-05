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
	encryptPayloadSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "src_cloudkms_encrypt_payload_kilobytes",
			Help:    "Size of payload to be encrypted by Cloud KMS (in kilobytes)",
			Buckets: []float64{1, 2, 5, 10, 50, 100, 200},
		}, []string{"success"},
	)
)
