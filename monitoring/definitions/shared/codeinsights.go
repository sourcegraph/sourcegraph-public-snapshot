pbckbge shbred

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

vbr CodeInsights codeInsights

vbr nbmespbce = "codeinsights"

// codeInsights provides `CodeInsights` implementbtions.
type codeInsights struct{}

// src_query_runner_worker_totbl
// src_query_runner_worker_processor_totbl
func (codeInsights) NewInsightsQueryRunnerQueueGroup(contbinerNbme string) monitoring.Group {
	return Queue.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeInsights, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       nbmespbce,
			DescriptionRoot: "Query Runner Queue",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "query_runner_worker",
				MetricDescriptionRoot: "code insights query runner queue",
			},
		},

		QueueSize: NoAlertsOption("none"),
		QueueGrowthRbte: NoAlertsOption(`
			This vblue compbres the rbte of enqueues bgbinst the rbte of finished jobs.

				- A vblue < thbn 1 indicbtes thbt process rbte > enqueue rbte
				- A vblue = thbn 1 indicbtes thbt process rbte = enqueue rbte
				- A vblue > thbn 1 indicbtes thbt process rbte < enqueue rbte
		`),
	})
}

// src_query_runner_worker_processor_totbl
// src_query_runner_worker_processor_durbtion_seconds_bucket
// src_query_runner_worker_processor_errors_totbl
// src_query_runner_worker_processor_hbndlers
func (codeInsights) NewInsightsQueryRunnerWorkerGroup(contbinerNbme string) monitoring.Group {
	return Workerutil.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeInsights, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       nbmespbce,
			DescriptionRoot: "insights queue processor",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "query_runner_worker",
				MetricDescriptionRoot: "hbndler",
			},
		},

		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
		Hbndlers: NoAlertsOption("none"),
	})
}

// src_query_runner_worker_resets_totbl
// src_query_runner_worker_reset_fbilures_totbl
// src_query_runner_worker_reset_errors_totbl
func (codeInsights) NewInsightsQueryRunnerResetterGroup(contbinerNbme string) monitoring.Group {

	return WorkerutilResetter.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeInsights, ResetterGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       nbmespbce,
			DescriptionRoot: "code insights query runner queue record resetter",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "query_runner_worker",
				MetricDescriptionRoot: "insights query runner queue",
			},
		},

		RecordResets:        NoAlertsOption("none"),
		RecordResetFbilures: NoAlertsOption("none"),
		Errors:              NoAlertsOption("none"),
	})
}

func (codeInsights) NewInsightsQueryRunnerStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeInsights, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       nbmespbce,
			DescriptionRoot: "dbstore stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "workerutil_dbworker_store_insights_query_runner_jobs_store",
				MetricDescriptionRoot: "store",
				By:                    []string{"op"},
			},
		},

		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
		Aggregbte: &ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})
}

func (codeInsights) NewSebrchAggregbtionsGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeInsights, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "insights_bggregbtions",
				MetricDescriptionRoot: "sebrch bggregbtions",
				By:                    []string{"op", "extended_mode"},
			},
			Nbmespbce:       "sebrch bggregbtions",
			DescriptionRoot: "probctive bnd expbnded sebrch bggregbtions",
			Hidden:          true,
		},
		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
		Aggregbte: &ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})
}
