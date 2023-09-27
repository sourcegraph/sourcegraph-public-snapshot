pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func PreciseCodeIntelWorker() *monitoring.Dbshbobrd {
	const contbinerNbme = "precise-code-intel-worker"

	return &monitoring.Dbshbobrd{
		Nbme:        "precise-code-intel-worker",
		Title:       "Precise Code Intel Worker",
		Description: "Hbndles conversion of uplobded precise code intelligence bundles.",
		Groups: []monitoring.Group{
			shbred.CodeIntelligence.NewUplobdQueueGroup(contbinerNbme),
			shbred.CodeIntelligence.NewUplobdProcessorGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDBStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewLSIFStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewUplobdDBWorkerStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewGitserverClientGroup(contbinerNbme),
			shbred.CodeIntelligence.NewUplobdStoreGroup(contbinerNbme),

			// Resource monitoring
			shbred.NewFrontendInternblAPIErrorResponseMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewDbtbbbseConnectionsMonitoringGroup(contbinerNbme),
			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewGolbngMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
		},
	}
}
