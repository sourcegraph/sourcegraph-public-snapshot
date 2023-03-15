package encryption

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
	// current state
	numEncryptedAtRest   *prometheus.GaugeVec
	numUnencryptedAtRest *prometheus.GaugeVec

	// processing status
	numRecordsEncrypted *prometheus.CounterVec
	numRecordsDecrypted *prometheus.CounterVec
	numErrors           prometheus.Counter
}

func newMetrics(observationCtx *observation.Context) *metrics {
	gaugeVec := func(name, help string) *prometheus.GaugeVec {
		gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}, []string{"tableName"})

		observationCtx.Registerer.MustRegister(gaugeVec)
		return gaugeVec
	}

	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	counterVec := func(name, help string) *prometheus.CounterVec {
		counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}, []string{"tableName"})

		observationCtx.Registerer.MustRegister(counterVec)
		return counterVec
	}

	numEncryptedAtRest := gaugeVec(
		"src_records_encrypted_at_rest_total",
		"The number of database records encrypted at rest.",
	)
	numUnencryptedAtRest := gaugeVec(
		"src_records_unencrypted_at_rest_total",
		"The number of database records unencrypted at rest.",
	)
	numRecordsEncrypted := counterVec(
		"src_records_encrypted_total",
		"The number of unencrypted database records that have been encrypted.",
	)
	numRecordsDecrypted := counterVec(
		"src_records_decrypted_total",
		"The number of encrypted database records that have been decrypted.",
	)
	numErrors := counter(
		"src_record_encryption_errors_total",
		"The number of errors that occur during record encryption/decryption.",
	)

	for _, config := range database.EncryptionConfigs {
		// Initialize counters to zero
		numRecordsEncrypted.WithLabelValues(config.TableName).Add(0)
		numRecordsDecrypted.WithLabelValues(config.TableName).Add(0)
	}

	return &metrics{
		numEncryptedAtRest:   numEncryptedAtRest,
		numUnencryptedAtRest: numUnencryptedAtRest,
		numRecordsEncrypted:  numRecordsEncrypted,
		numRecordsDecrypted:  numRecordsDecrypted,
		numErrors:            numErrors,
	}
}
