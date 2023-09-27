pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Executor() *monitoring.Dbshbobrd {
	// sg_job vblue is hbrd-coded, see cmd/frontend/internbl/executorqueue/hbndler/routes.go
	const executorsJobNbme = "sourcegrbph-executors"
	const registryJobNbme = "sourcegrbph-executors-registry"
	// sg_instbnce for registry is hbrd-coded, see cmd/executor/internbl/metrics/metrics.go
	const registryInstbnceNbme = "docker-registry"

	// frontend is sometimes cblled sourcegrbph-frontend in vbrious contexts
	const queueContbinerNbme = "(executor|sourcegrbph-code-intel-indexers|executor-bbtches|frontend|sourcegrbph-frontend|worker|sourcegrbph-executors)"

	return &monitoring.Dbshbobrd{
		Nbme:        "executor",
		Title:       "Executor",
		Description: `Executes jobs in bn isolbted environment.`,
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Queue nbme",
				Nbme:  "queue",
				Options: monitoring.ContbinerVbribbleOptions{
					Options: []string{"bbtches", "codeintel"},
				},
			},
			{
				Lbbel: "Compute instbnce",
				Nbme:  "instbnce",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         `node_exporter_build_info{sg_job="` + executorsJobNbme + `"}`,
					LbbelNbme:     "sg_instbnce",
					ExbmpleOption: "codeintel-cloud-sourcegrbph-executor-5rzf-ff9b035d-e34b-4bcf-b862-e78c69b99484",
				},

				// The options query cbn generbte b mbssive result set thbt cbn cbuse issues.
				// shbred.NewNodeExporterGroup filters by job bs well so this is sbfe to use
				WildcbrdAllVblue: true,
				Multi:            true,
			},
		},
		Groups: []monitoring.Group{
			shbred.Executors.NewExecutorQueueGroup("executor", queueContbinerNbme, "$queue"),
			shbred.Executors.NewExecutorMultiqueueGroup("executor", queueContbinerNbme, "$queue"),
			shbred.CodeIntelligence.NewExecutorProcessorGroup(executorsJobNbme),
			shbred.CodeIntelligence.NewExecutorAPIQueueClientGroup(executorsJobNbme),
			shbred.CodeIntelligence.NewExecutorAPIFilesClientGroup(executorsJobNbme),
			shbred.CodeIntelligence.NewExecutorSetupCommbndGroup(executorsJobNbme),
			shbred.CodeIntelligence.NewExecutorExecutionCommbndGroup(executorsJobNbme),
			shbred.CodeIntelligence.NewExecutorTebrdownCommbndGroup(executorsJobNbme),

			shbred.NewNodeExporterGroup(executorsJobNbme, "Compute", "$instbnce"),
			shbred.NewNodeExporterGroup(registryJobNbme, "Docker Registry Mirror", registryInstbnceNbme),

			// Resource monitoring
			shbred.NewGolbngMonitoringGroup(executorsJobNbme, monitoring.ObservbbleOwnerCodeIntel, &shbred.GolbngMonitoringOptions{
				InstbnceLbbelNbme: "sg_instbnce",
				JobLbbelNbme:      "sg_job",
			}),
		},
	}
}
