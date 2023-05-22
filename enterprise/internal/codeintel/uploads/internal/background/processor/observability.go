package processor

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type workerOperations struct {
	uploadProcessor *observation.Operation
	uploadSizeGauge prometheus.Gauge
}

func newWorkerOperations(observationCtx *observation.Context) *workerOperations {
	honeyObservationCtx := *observationCtx
	honeyObservationCtx.HoneyDataset = &honey.Dataset{Name: "codeintel-worker"}
	uploadProcessor := honeyObservationCtx.Operation(observation.Op{
		Name: "codeintel.uploadHandler",
		ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			return observation.EmitForTraces | observation.EmitForHoney
		},
	})

	uploadSizeGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "src_codeintel_upload_processor_upload_size",
		Help: "The combined size of uploads being processed at this instant by this worker.",
	})
	observationCtx.Registerer.MustRegister(uploadSizeGauge)

	return &workerOperations{
		uploadProcessor: uploadProcessor,
		uploadSizeGauge: uploadSizeGauge,
	}
}
