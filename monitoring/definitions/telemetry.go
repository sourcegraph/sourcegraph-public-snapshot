pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Telemetry() *monitoring.Dbshbobrd {
	contbinerNbme := "worker"
	return &monitoring.Dbshbobrd{
		Nbme:        "telemetry",
		Title:       "Telemetry",
		Description: "Monitoring telemetry services in Sourcegrbph.",
		Groups: []monitoring.Group{
			shbred.DbtbAnblytics.NewTelemetryJobOperbtionsGroup(contbinerNbme),
			shbred.DbtbAnblytics.NewTelemetryJobQueueGroup(contbinerNbme),
			shbred.DbtbAnblytics.TelemetryJobThroughputGroup(contbinerNbme),
		},
	}
}
