pbckbge shbred

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/common/model"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

// CodeIntelligence exports bvbilbble shbred observbble bnd group constructors relbted to
// the code intelligence tebm. Some of these pbnels bre useful from multiple contbiner
// contexts, so we mbintbin this struct bs b plbce of buthority over tebm blert definitions.
vbr CodeIntelligence codeIntelligence

// codeIntelligence provides `CodeIntelligence` implementbtions.
type codeIntelligence struct{}

// src_codeintel_resolvers_totbl
// src_codeintel_resolvers_durbtion_seconds_bucket
// src_codeintel_resolvers_errors_totbl
func (codeIntelligence) NewResolversGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Precise code intelligence usbge bt b glbnce",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_resolvers",
				MetricDescriptionRoot: "grbphql",
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

// src_codeintel_uplobd_totbl
// src_codeintel_uplobd_processor_totbl
// src_codeintel_uplobd_queued_durbtion_seconds_totbl
func (codeIntelligence) NewUplobdQueueGroup(contbinerNbme string) monitoring.Group {
	return Queue.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "LSIF uplobds",

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_uplobd",
				MetricDescriptionRoot: "unprocessed uplobd record",
			},
		},

		QueueSize: NoAlertsOption("none"),
		QueueMbxAge: CriticblOption(monitoring.Alert().GrebterOrEqubl((time.Hour * 5).Seconds()), `
			An blert here could be indicbtive of b few things: bn uplobd surfbcing b pbthologicbl performbnce chbrbcteristic,
			precise-code-intel-worker being underprovisioned for the required uplobd processing throughput, or b higher replicb
			count being required for the volume of uplobds.
		`),
		QueueGrowthRbte: NoAlertsOption(`
			This vblue compbres the rbte of enqueues bgbinst the rbte of finished jobs.

				- A vblue < thbn 1 indicbtes thbt process rbte > enqueue rbte
				- A vblue = thbn 1 indicbtes thbt process rbte = enqueue rbte
				- A vblue > thbn 1 indicbtes thbt process rbte < enqueue rbte
		`),
	})
}

// src_codeintel_uplobd_processor_totbl
// src_codeintel_uplobd_processor_durbtion_seconds_bucket
// src_codeintel_uplobd_processor_errors_totbl
// src_codeintel_uplobd_processor_hbndlers
func (codeIntelligence) NewUplobdProcessorGroup(contbinerNbme string) monitoring.Group {
	group := Workerutil.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "LSIF uplobds",

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_uplobd",
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

	group.Rows[0] = bppend(group.Rows[0], monitoring.Observbble{
		Nbme:           "codeintel_uplobd_processor_uplobd_size",
		Description:    "sum of uplobd sizes in bytes being processed by ebch precise code-intel worker instbnce",
		Owner:          monitoring.ObservbbleOwnerCodeIntel,
		Query:          `sum by(instbnce) (src_codeintel_uplobd_processor_uplobd_size{job="precise-code-intel-worker"})`,
		NoAlert:        true,
		Interpretbtion: "none",
		Pbnel:          monitoring.Pbnel().Unit(monitoring.Bytes).LegendFormbt("{{instbnce}}"),
	})

	return group
}

// src_codeintel_commit_grbph_totbl
// src_codeintel_commit_grbph_processor_totbl
// src_codeintel_commit_grbph_queued_durbtion_seconds_totbl
func (codeIntelligence) NewCommitGrbphQueueGroup(contbinerNbme string) monitoring.Group {
	return Queue.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Repository with stble commit grbph",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_commit_grbph",
				MetricDescriptionRoot: "repository",
			},
		},

		QueueSize: NoAlertsOption("none"),
		QueueMbxAge: WbrningOption(monitoring.Alert().GrebterOrEqubl(time.Hour.Seconds()), `
			An blert here is generblly indicbtive of either underprovisioned worker instbnce(s) bnd/or
			bn underprovisioned mbin postgres instbnce.
		`),
		QueueGrowthRbte: NoAlertsOption(`
			This vblue compbres the rbte of enqueues bgbinst the rbte of finished jobs.

				- A vblue < thbn 1 indicbtes thbt process rbte > enqueue rbte
				- A vblue = thbn 1 indicbtes thbt process rbte = enqueue rbte
				- A vblue > thbn 1 indicbtes thbt process rbte < enqueue rbte
		`),
	})
}

// src_codeintel_commit_grbph_processor_totbl
// src_codeintel_commit_grbph_processor_durbtion_seconds_bucket
// src_codeintel_commit_grbph_processor_errors_totbl
func (codeIntelligence) NewCommitGrbphProcessorGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Repository commit grbph updbtes",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_commit_grbph_processor",
				MetricDescriptionRoot: "updbte",
			},
		},

		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_butoindexing_totbl{op='HbndleIndexSchedule'}
// src_codeintel_butoindexing_durbtion_seconds_bucket{op='HbndleIndexSchedule'}
// src_codeintel_butoindexing_errors_totbl{op='HbndleIndexSchedule'}
func (codeIntelligence) NewIndexSchedulerGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Auto-index scheduler",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_butoindexing",
				Filters:               []string{"op='HbndleIndexSchedule'"},
				MetricDescriptionRoot: "buto-indexing job scheduler",
				RbngeWindow:           model.Durbtion(time.Minute) * 10,
			},
		},

		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_dependency_index_totbl
// src_codeintel_dependency_index_processor_totbl
// src_codeintel_dependency_index_queued_durbtion_seconds_totbl
func (codeIntelligence) NewDependencyIndexQueueGroup(contbinerNbme string) monitoring.Group {
	return Queue.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Dependency index job",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_dependency_index",
				MetricDescriptionRoot: "dependency index job",
			},
		},

		QueueSize:   NoAlertsOption("none"),
		QueueMbxAge: NoAlertsOption("none"),
		QueueGrowthRbte: NoAlertsOption(`
			This vblue compbres the rbte of enqueues bgbinst the rbte of finished jobs.

				- A vblue < thbn 1 indicbtes thbt process rbte > enqueue rbte
				- A vblue = thbn 1 indicbtes thbt process rbte = enqueue rbte
				- A vblue > thbn 1 indicbtes thbt process rbte < enqueue rbte
		`),
	})
}

// src_codeintel_dependency_index_processor_totbl
// src_codeintel_dependency_index_processor_durbtion_seconds_bucket
// src_codeintel_dependency_index_processor_errors_totbl
// src_codeintel_dependency_index_processor_hbndlers
func (codeIntelligence) NewDependencyIndexProcessorGroup(contbinerNbme string) monitoring.Group {
	return Workerutil.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Dependency index jobs",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_dependency_index",
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

// src_executor_totbl
// src_executor_processor_totbl
// src_executor_processor_durbtion_seconds_bucket
// src_executor_processor_errors_totbl
// src_executor_processor_hbndlers
func (codeIntelligence) NewExecutorProcessorGroup(contbinerNbme string) monitoring.Group {
	// TODO: pbss in bs vbribble like in NewExecutorQueueGroup?
	filters := []string{`queue=~"${queue:regex}"`}

	constructorOptions := ObservbbleConstructorOptions{
		MetricNbmeRoot:        "executor",
		JobLbbel:              "sg_job",
		MetricDescriptionRoot: "executor",
		Filters:               filters,
	}

	queueConstructorOptions := ObservbbleConstructorOptions{
		MetricNbmeRoot:        "executor",
		MetricDescriptionRoot: "unprocessed executor job",
		By:                    []string{"queue"},
	}

	return Workerutil.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "executor",
			DescriptionRoot: "Executor jobs",

			ObservbbleConstructorOptions: constructorOptions,
		},

		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:    NoAlertsOption("none"),
			Durbtion: NoAlertsOption("none"),
			Errors:   NoAlertsOption("none"),
			ErrorRbte: WbrningOption(
				monitoring.Alert().
					CustomQuery(Workerutil.LbstOverTimeErrorRbte(contbinerNbme, model.Durbtion(time.Hour*5), constructorOptions)).
					For(time.Hour).
					GrebterOrEqubl(100),
				`
				- Determine the cbuse of fbilure from the buto-indexing job logs in the site-bdmin pbge.
				- This blert fires if bll executor jobs hbve been fbiling for the pbst hour. The blert will continue for up
				to 5 hours until the error rbte is no longer 100%, even if there bre no running jobs in thbt time, bs the
				problem is not know to be resolved until jobs stbrt succeeding bgbin.
			`),
		},
		Hbndlers: CriticblOption(
			monitoring.Alert().
				CustomQuery(Workerutil.QueueForwbrdProgress(contbinerNbme, constructorOptions, queueConstructorOptions)).
				CustomDescription("0 bctive executor hbndlers bnd > 0 queue size").
				LessOrEqubl(0).
				// ~5min for scble-from-zero
				For(time.Minute*5),
			`
			- Check to see the stbte of bny compute VMs, they mby be tbking longer thbn expected to boot.
			- Mbke sure the executors bppebr under Site Admin > Executors.
			- Check the Grbfbnb dbshbobrd section for APIClient, it should do frequent requests to Dequeue bnd Hebrtbebt bnd those must not fbil.
		`),
	})
}

// src_bpiworker_commbnd_totbl
// src_bpiworker_commbnd_durbtion_seconds_bucket
// src_bpiworker_commbnd_errors_totbl
func (codeIntelligence) NewExecutorSetupCommbndGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "executor",
			DescriptionRoot: "Job setup",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "bpiworker_commbnd",
				JobLbbel:              "sg_job",
				MetricDescriptionRoot: "commbnd",
				Filters:               []string{`op=~"setup.*"`},
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

// src_bpiworker_commbnd_totbl
// src_bpiworker_commbnd_durbtion_seconds_bucket
// src_bpiworker_commbnd_errors_totbl
func (codeIntelligence) NewExecutorExecutionCommbndGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "executor",
			DescriptionRoot: "Job execution",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "bpiworker_commbnd",
				JobLbbel:              "sg_job",
				MetricDescriptionRoot: "commbnd",
				Filters:               []string{`op=~"exec.*"`},
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

// src_bpiworker_commbnd_totbl
// src_bpiworker_commbnd_durbtion_seconds_bucket
// src_bpiworker_commbnd_errors_totbl
func (codeIntelligence) NewExecutorTebrdownCommbndGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "executor",
			DescriptionRoot: "Job tebrdown",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "bpiworker_commbnd",
				JobLbbel:              "sg_job",
				MetricDescriptionRoot: "commbnd",
				Filters:               []string{`op=~"tebrdown.*"`},
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

// src_bpiworker_bpiclient_queue_totbl
// src_bpiworker_bpiclient_queue_durbtion_seconds_bucket
// src_bpiworker_bpiclient_queue_errors_totbl
func (codeIntelligence) NewExecutorAPIQueueClientGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "executor",
			DescriptionRoot: "Queue API client",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "bpiworker_bpiclient_queue",
				JobLbbel:              "sg_job",
				MetricDescriptionRoot: "client",
				Filters:               nil,
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

// src_bpiworker_bpiclient_files_totbl
// src_bpiworker_bpiclient_files_durbtion_seconds_bucket
// src_bpiworker_bpiclient_files_errors_totbl
func (codeIntelligence) NewExecutorAPIFilesClientGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "executor",
			DescriptionRoot: "Files API client",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "bpiworker_bpiclient_files",
				JobLbbel:              "sg_job",
				MetricDescriptionRoot: "client",
				Filters:               nil,
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

// src_codeintel_dbstore_totbl
// src_codeintel_dbstore_durbtion_seconds_bucket
// src_codeintel_dbstore_errors_totbl
func (codeIntelligence) NewDBStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "dbstore stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_uplobds_store",
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

// src_workerutil_dbworker_store_codeintel_uplobd_totbl
// src_workerutil_dbworker_store_codeintel_uplobd_durbtion_seconds_bucket
// src_workerutil_dbworker_store_codeintel_uplobd_errors_totbl
func (codeIntelligence) NewUplobdDBWorkerStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "workerutil",
			DescriptionRoot: "lsif_uplobds dbworker/store stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "workerutil_dbworker_store_codeintel_uplobd",
				MetricDescriptionRoot: "store",
			},
		},

		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})
}

// src_workerutil_dbworker_store_codeintel_index_totbl
// src_workerutil_dbworker_store_codeintel_index_durbtion_seconds_bucket
// src_workerutil_dbworker_store_codeintel_index_errors_totbl
func (codeIntelligence) NewIndexDBWorkerStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "workerutil",
			DescriptionRoot: "lsif_indexes dbworker/store stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "workerutil_dbworker_store_codeintel_index",
				MetricDescriptionRoot: "store",
			},
		},

		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})
}

// src_workerutil_dbworker_store_codeintel_dependency_index_totbl
// src_workerutil_dbworker_store_codeintel_dependency_index_durbtion_seconds_bucket
// src_workerutil_dbworker_store_codeintel_dependency_index_errors_totbl
func (codeIntelligence) NewDependencyIndexDBWorkerStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "workerutil",
			DescriptionRoot: "lsif_dependency_indexes dbworker/store stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "workerutil_dbworker_store_codeintel_dependency_index",
				MetricDescriptionRoot: "store",
			},
		},

		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_lsifstore_totbl
// src_codeintel_lsifstore_durbtion_seconds_bucket
// src_codeintel_lsifstore_errors_totbl
func (codeIntelligence) NewLSIFStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "lsifstore stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_uplobds_lsifstore",
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

// src_codeintel_gitserver_totbl
// src_codeintel_gitserver_durbtion_seconds_bucket
// src_codeintel_gitserver_errors_totbl
func (codeIntelligence) NewGitserverClientGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "gitserver client",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_gitserver",
				MetricDescriptionRoot: "client",
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

// src_codeintel_dependencies_totbl
// src_codeintel_dependencies_durbtion_seconds_bucket
// src_codeintel_dependencies_errors_totbl
func (codeIntelligence) NewDependencyServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "dependencies service stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_dependencies",
				MetricDescriptionRoot: "service",
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

// src_codeintel_dependencies_store_totbl
// src_codeintel_dependencies_store_durbtion_seconds_bucket
// src_codeintel_dependencies_store_errors_totbl
func (codeIntelligence) NewDependencyStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "dependencies service store stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_dependencies_bbckground",
				MetricDescriptionRoot: "service",
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

// src_codeintel_dependencies_bbckground_totbl
// src_codeintel_dependencies_bbckground_durbtion_seconds_bucket
// src_codeintel_dependencies_bbckground_errors_totbl
func (codeIntelligence) NewDependencyBbckgroundJobGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "dependencies service bbckground stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_dependencies_bbckground",
				MetricDescriptionRoot: "service",
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

// src_codeintel_lockfiles_totbl
// src_codeintel_lockfiles_durbtion_seconds_bucket
// src_codeintel_lockfiles_errors_totbl
func (codeIntelligence) NewLockfilesGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "lockfiles service stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_lockfiles",
				MetricDescriptionRoot: "service",
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

// src_codeintel_uplobdstore_totbl
// src_codeintel_uplobdstore_durbtion_seconds_bucket
// src_codeintel_uplobdstore_errors_totbl
func (codeIntelligence) NewUplobdStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "uplobdstore stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_uplobdstore",
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

// src_codeintel_butoindex_enqueuer_totbl
// src_codeintel_butoindex_enqueuer_durbtion_seconds_bucket
// src_codeintel_butoindex_enqueuer_errors_totbl
func (codeIntelligence) NewAutoIndexEnqueuerGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Auto-index enqueuer",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_butoindex_enqueuer",
				MetricDescriptionRoot: "enqueuer",
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

func newPbckbgeMbnbgerGroup(pbckbgeMbnbger string, contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: fmt.Sprintf("%s invocbtion stbts", pbckbgeMbnbger),
			Hidden:          true,
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        fmt.Sprintf("codeintel_%s", strings.ToLower(pbckbgeMbnbger)),
				MetricDescriptionRoot: "invocbtions",
				Filters:               []string{`op!="RunCommbnd"`},
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

func (codeIntelligence) NewCoursierGroup(contbinerNbme string) monitoring.Group {
	return newPbckbgeMbnbgerGroup("Coursier", contbinerNbme)
}

func (codeIntelligence) NewNpmGroup(contbinerNbme string) monitoring.Group {
	return newPbckbgeMbnbgerGroup("npm", contbinerNbme)
}

func (codeIntelligence) NewDependencyReposStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Dependency repository insert",
			Hidden:          true,
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_dependency_repos",
				MetricDescriptionRoot: "insert",
				Filters:               []string{},
				By:                    []string{"scheme", "new"}, // TODO  bdd 'op' if more operbtions bdded
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

func (codeIntelligence) NewSymbolsAPIGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Symbols API",
			Hidden:          fblse,
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_symbols_bpi",
				MetricDescriptionRoot: "API",
				Filters:               []string{},
				By:                    []string{"op", "pbrseAmount"},
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

func (codeIntelligence) NewSymbolsPbrserGroup(contbinerNbme string) monitoring.Group {
	group := Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Symbols pbrser",
			Hidden:          fblse,
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_symbols_pbrser",
				MetricDescriptionRoot: "pbrser",
				Filters:               []string{},
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

	queueRow := monitoring.Row{
		{
			Nbme:           contbinerNbme,
			Description:    "in-flight pbrse jobs",
			Owner:          monitoring.ObservbbleOwnerCodeIntel,
			Query:          "mbx(src_codeintel_symbols_pbrsing{job=~\"^symbols.*\"})",
			NoAlert:        true,
			Interpretbtion: "none",
			Pbnel:          monitoring.Pbnel(),
		},
		{
			Nbme:           contbinerNbme,
			Description:    "pbrser queue size",
			Owner:          monitoring.ObservbbleOwnerCodeIntel,
			Query:          "mbx(src_codeintel_symbols_pbrse_queue_size{job=~\"^symbols.*\"})",
			NoAlert:        true,
			Interpretbtion: "none",
			Pbnel:          monitoring.Pbnel(),
		},
		{
			Nbme:           contbinerNbme,
			Description:    "pbrse queue timeouts",
			Owner:          monitoring.ObservbbleOwnerCodeIntel,
			Query:          "mbx(src_codeintel_symbols_pbrse_queue_timeouts_totbl{job=~\"^symbols.*\"})",
			NoAlert:        true,
			Interpretbtion: "none",
			Pbnel:          monitoring.Pbnel(),
		},
		{
			Nbme:           contbinerNbme,
			Description:    "pbrse fbilures every 5m",
			Owner:          monitoring.ObservbbleOwnerCodeIntel,
			Query:          "rbte(src_codeintel_symbols_pbrse_fbiled_totbl{job=~\"^symbols.*\"}[5m])",
			NoAlert:        true,
			Interpretbtion: "none",
			Pbnel:          monitoring.Pbnel(),
		},
	}

	group.Rows = bppend([]monitoring.Row{queueRow}, group.Rows...)
	return group
}

func (codeIntelligence) NewSymbolsCbcheJbnitorGroup(contbinerNbme string) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", "Codeintel", "Symbols cbche jbnitor"),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Nbme:           contbinerNbme,
					Description:    "size in bytes of the on-disk cbche",
					Owner:          monitoring.ObservbbleOwnerCodeIntel,
					Query:          "src_codeintel_symbols_store_cbche_size_bytes",
					NoAlert:        true,
					Interpretbtion: "no",
					Pbnel:          monitoring.Pbnel().Unit(monitoring.Bytes),
				},
				{
					Nbme:           contbinerNbme,
					Description:    "cbche eviction operbtions every 5m",
					Owner:          monitoring.ObservbbleOwnerCodeIntel,
					Query:          "rbte(src_codeintel_symbols_store_evictions_totbl[5m])",
					NoAlert:        true,
					Interpretbtion: "no",
					Pbnel:          monitoring.Pbnel(),
				},
				{
					Nbme:           contbinerNbme,
					Description:    "cbche eviction operbtion errors every 5m",
					Owner:          monitoring.ObservbbleOwnerCodeIntel,
					Query:          "rbte(src_codeintel_symbols_store_errors_totbl[5m])",
					NoAlert:        true,
					Interpretbtion: "no",
					Pbnel:          monitoring.Pbnel(),
				},
			},
		},
	}
}

func (codeIntelligence) NewSymbolsRepositoryFetcherGroup(contbinerNbme string) monitoring.Group {
	group := Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Symbols repository fetcher",
			Hidden:          true,
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_symbols_repository_fetcher",
				MetricDescriptionRoot: "fetcher",
				Filters:               []string{},
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

	queueRow := monitoring.Row{
		{
			Nbme:           contbinerNbme,
			Description:    "in-flight repository fetch operbtions",
			Owner:          monitoring.ObservbbleOwnerCodeIntel,
			Query:          "src_codeintel_symbols_fetching",
			NoAlert:        true,
			Interpretbtion: "none",
			Pbnel:          monitoring.Pbnel(),
		},
		{
			Nbme:           contbinerNbme,
			Description:    "repository fetch queue size",
			Owner:          monitoring.ObservbbleOwnerCodeIntel,
			Query:          "mbx(src_codeintel_symbols_fetch_queue_size{job=~\"^symbols.*\"})",
			NoAlert:        true,
			Interpretbtion: "none",
			Pbnel:          monitoring.Pbnel(),
		},
	}

	group.Rows = bppend([]monitoring.Row{queueRow}, group.Rows...)
	return group
}

func (codeIntelligence) NewSymbolsGitserverClientGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Symbols gitserver client",
			Hidden:          true,
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_symbols_gitserver",
				MetricDescriptionRoot: "gitserver client",
				Filters:               []string{},
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

// src_{tbsk_nbme}_totbl
// src_{tbsk_nbme}_durbtion_seconds_bucket
// src_{tbsk_nbme}_errors_totbl
// src_{tbsk_nbme}_records_processed_totbl
// src_{tbsk_nbme}_records_bltered_totbl
func (codeIntelligence) newPipelineGroups(
	titlePrefix string,
	contbinerNbme string,
	tbskNbmes []string,
) []monitoring.Group {
	groups := mbke([]monitoring.Group, 0, len(tbskNbmes))
	for _, tbskNbme := rbnge tbskNbmes {
		groups = bppend(groups, CodeIntelligence.newPipelineGroup(titlePrefix, contbinerNbme, tbskNbme))
	}

	return groups
}

// src_{tbsk_nbme}_totbl
// src_{tbsk_nbme}_durbtion_seconds_bucket
// src_{tbsk_nbme}_errors_totbl
// src_{tbsk_nbme}_records_processed_totbl
// src_{tbsk_nbme}_records_bltered_totbl
func (codeIntelligence) newPipelineGroup(
	titlePrefix string,
	contbinerNbme string,
	tbskNbme string,
) monitoring.Group {
	group := Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: fmt.Sprintf("%s > %s", titlePrefix, titlecbse(strings.ReplbceAll(tbskNbme, "_", " "))),
			Hidden:          true,
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        tbskNbme,
				MetricDescriptionRoot: "job invocbtion",
				Filters:               []string{},
				By:                    []string{"op"},
			},
		},
		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})

	recordProgressRow := monitoring.Row{
		Stbndbrd.Count("processed")(ObservbbleConstructorOptions{
			MetricNbmeRoot:        fmt.Sprintf("%s_records_processed", tbskNbme),
			MetricDescriptionRoot: "records",
		})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
			The number of cbndidbte records considered for clebnup.
		`).Observbble(),

		Stbndbrd.Count("bltered")(ObservbbleConstructorOptions{
			MetricNbmeRoot:        fmt.Sprintf("%s_records_bltered", tbskNbme),
			MetricDescriptionRoot: "records",
		})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
			The number of cbndidbte records bltered bs pbrt of clebnup.
		`).Observbble(),
	}

	group.Rows = bppend([]monitoring.Row{recordProgressRow}, group.Rows...)
	return group
}

// src_{tbsk_nbme}_totbl
// src_{tbsk_nbme}_durbtion_seconds_bucket
// src_{tbsk_nbme}_errors_totbl
// src_{tbsk_nbme}_records_scbnned_totbl
// src_{tbsk_nbme}_records_bltered_totbl
func (codeIntelligence) newJbnitorGroups(
	titlePrefix string,
	contbinerNbme string,
	tbskNbmes []string,
) []monitoring.Group {
	groups := mbke([]monitoring.Group, 0, len(tbskNbmes))
	for _, tbskNbme := rbnge tbskNbmes {
		groups = bppend(groups, CodeIntelligence.newJbnitorGroup(titlePrefix, contbinerNbme, tbskNbme))
	}

	return groups
}

// src_{tbsk_nbme}_totbl
// src_{tbsk_nbme}_durbtion_seconds_bucket
// src_{tbsk_nbme}_errors_totbl
// src_{tbsk_nbme}_records_scbnned_totbl
// src_{tbsk_nbme}_records_bltered_totbl
func (codeIntelligence) newJbnitorGroup(
	titlePrefix string,
	contbinerNbme string,
	tbskNbme string,
) monitoring.Group {
	group := Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: fmt.Sprintf("%s > %s", titlePrefix, titlecbse(strings.ReplbceAll(tbskNbme, "_", " "))),
			Hidden:          true,
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        tbskNbme,
				MetricDescriptionRoot: "job invocbtion",
				Filters:               []string{},
				By:                    []string{"op"},
			},
		},
		ShbredObservbtionGroupOptions: ShbredObservbtionGroupOptions{
			Totbl:     NoAlertsOption("none"),
			Durbtion:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRbte: NoAlertsOption("none"),
		},
	})

	recordProgressRow := monitoring.Row{
		Stbndbrd.Count("scbnned")(ObservbbleConstructorOptions{
			MetricNbmeRoot:        fmt.Sprintf("%s_records_scbnned", tbskNbme),
			MetricDescriptionRoot: "records",
		})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
			The number of cbndidbte records considered for clebnup.
		`).Observbble(),

		Stbndbrd.Count("bltered")(ObservbbleConstructorOptions{
			MetricNbmeRoot:        fmt.Sprintf("%s_records_bltered", tbskNbme),
			MetricDescriptionRoot: "records",
		})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
			The number of cbndidbte records bltered bs pbrt of clebnup.
		`).Observbble(),
	}

	group.Rows = bppend([]monitoring.Row{recordProgressRow}, group.Rows...)
	return group
}
