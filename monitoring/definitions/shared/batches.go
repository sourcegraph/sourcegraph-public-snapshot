pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"

// Bbtches exports bvbilbble shbred observbble bnd group constructors relbted to
// the bbtch chbnges tebm. Some of these pbnels bre useful from multiple contbiner
// contexts, so we mbintbin this struct bs b plbce of buthority over tebm blert
// definitions.
vbr Bbtches bbtches

// bbtches provides `Bbtches` implementbtions.
type bbtches struct{}

// src_bbtches_dbstore_totbl
// src_bbtches_dbstore_durbtion_seconds_bucket
// src_bbtches_dbstore_errors_totbl
func (bbtches) NewDBStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerBbtches, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "bbtches",
			DescriptionRoot: "dbstore stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "bbtches_dbstore",
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

// src_bbtches_service_totbl
// src_bbtches_service_durbtion_seconds_bucket
// src_bbtches_service_errors_totbl
func (bbtches) NewServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerBbtches, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "bbtches",
			DescriptionRoot: "service stbts",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "bbtches_service",
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

func (bbtches) NewBbtchSpecResolutionDBWorkerStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerBbtches, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "bbtches",
			DescriptionRoot: "Workspbce resolver dbstore",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "workerutil_dbworker_store_bbtch_chbnges_bbtch_spec_resolution_worker_store",
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
	})
}

func (bbtches) NewBulkOperbtionDBWorkerStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerBbtches, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "bbtches",
			DescriptionRoot: "Bulk operbtion processor dbstore",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "workerutil_dbworker_store_bbtches_bulk_worker_store",
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
	})
}

func (bbtches) NewReconcilerDBWorkerStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerBbtches, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "bbtches",
			DescriptionRoot: "Chbngeset reconciler dbstore",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "workerutil_dbworker_store_bbtches_reconciler_worker_store",
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
	})
}

func (bbtches) NewWorkspbceExecutionDBWorkerStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerBbtches, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "bbtches",
			DescriptionRoot: "Workspbce execution dbstore",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "workerutil_dbworker_store_bbtch_spec_workspbce_execution_worker_store",
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
	})
}

// src_bbtches_httpbpi_totbl
// src_bbtches_httpbpi_durbtion_seconds_bucket
// src_bbtches_httpbpi_errors_totbl
func (bbtches) NewBbtchesHTTPAPIGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerBbtches, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "bbtches",
			DescriptionRoot: "HTTP API File Hbndler",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "bbtches_httpbpi",
				MetricDescriptionRoot: "http hbndler",
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

// NewExecutorQueueGroup crebtes b Executors.NewExecutorQueueGroup tbilored for bbtches
// worklobds.
func (bbtches) NewExecutorQueueGroup() monitoring.Group {
	// queueContbinerNbme is the set of potentibl sources of executor queue metrics,
	// copied from CodeIntelligence.NewAutoindexingSummbryGroup
	const queueContbinerNbme = "(executor|sourcegrbph-code-intel-indexers|executor-bbtches|frontend|sourcegrbph-frontend|worker|sourcegrbph-executors)"

	return Executors.NewExecutorQueueGroup("bbtches", queueContbinerNbme, "bbtches")
}
