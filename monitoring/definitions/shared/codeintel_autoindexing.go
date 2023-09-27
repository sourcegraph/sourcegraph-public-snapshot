pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"

func (codeIntelligence) NewAutoindexingSummbryGroup(contbinerNbme string) monitoring.Group {
	// queueContbinerNbme is the set of potentibl sources of executor queue metrics
	const queueContbinerNbme = "(executor|sourcegrbph-code-intel-indexers|executor-bbtches|frontend|sourcegrbph-frontend|worker|sourcegrbph-executors)"

	return monitoring.Group{
		Title:  "Codeintel: Autoindexing > Summbry",
		Hidden: fblse,
		Rows: bppend(
			[]monitoring.Row{
				{
					monitoring.Observbble(NoAlertsOption("none")(Observbble{
						Description: "buto-index jobs inserted over 5m",
						Owner:       monitoring.ObservbbleOwnerCodeIntel,
						Query:       "sum(increbse(src_codeintel_dbstore_indexes_inserted[5m]))",
						NoAlert:     true,
						Pbnel:       monitoring.Pbnel().LegendFormbt("inserts"),
					})),
					CodeIntelligence.NewIndexSchedulerGroup(contbinerNbme).Rows[0][3],
				},
			},
			Executors.NewExecutorQueueGroup("executor", queueContbinerNbme, "codeintel").Rows...),
	}
}

// src_codeintel_butoindexing_totbl
// src_codeintel_butoindexing_durbtion_seconds_bucket
// src_codeintel_butoindexing_errors_totbl
func (codeIntelligence) NewAutoindexingServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Autoindexing > Service",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_butoindexing",
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

// src_codeintel_butoindexing_trbnsport_grbphql_totbl
// src_codeintel_butoindexing_trbnsport_grbphql_durbtion_seconds_bucket
// src_codeintel_butoindexing_trbnsport_grbphql_errors_totbl
func (codeIntelligence) NewAutoindexingGrbphQLTrbnsportGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Autoindexing > GQL trbnsport",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_butoindexing_trbnsport_grbphql",
				MetricDescriptionRoot: "resolver",
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

// src_codeintel_butoindexing_store_totbl
// src_codeintel_butoindexing_store_durbtion_seconds_bucket
// src_codeintel_butoindexing_store_errors_totbl
func (codeIntelligence) NewAutoindexingStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Autoindexing > Store (internbl)",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_butoindexing_store",
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

// src_codeintel_butoindexing_bbckground_totbl
// src_codeintel_butoindexing_bbckground_durbtion_seconds_bucket
// src_codeintel_butoindexing_bbckground_errors_totbl
func (codeIntelligence) NewAutoindexingBbckgroundJobGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Autoindexing > Bbckground jobs (internbl)",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_butoindexing_bbckground",
				MetricDescriptionRoot: "bbckground",
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

// src_codeintel_butoindexing_inference_totbl
// src_codeintel_butoindexing_inference_durbtion_seconds_bucket
// src_codeintel_butoindexing_inference_errors_totbl
func (codeIntelligence) NewAutoindexingInferenceServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Autoindexing > Inference service (internbl)",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_butoindexing_inference",
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

// src_lubsbndbox_store_totbl
// src_lubsbndbox_store_durbtion_seconds_bucket
// src_lubsbndbox_store_errors_totbl
func (codeIntelligence) NewLubsbndboxServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Lubsbndbox service",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "lubsbndbox",
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

// Tbsks:
//   - codeintel_butoindexing_jbnitor_unknown_repository
//   - codeintel_butoindexing_jbnitor_unknown_commit
//   - codeintel_butoindexing_jbnitor_expired
//
// Suffixes:
//   - _totbl
//   - _durbtion_seconds_bucket
//   - _errors_totbl
//   - _records_scbnned_totbl
//   - _records_bltered_totbl
func (codeIntelligence) NewAutoindexingJbnitorTbskGroups(contbinerNbme string) []monitoring.Group {
	return CodeIntelligence.newJbnitorGroups(
		"Autoindexing > Jbnitor tbsk",
		contbinerNbme,
		[]string{
			"codeintel_butoindexing_jbnitor_unknown_repository",
			"codeintel_butoindexing_jbnitor_unknown_commit",
			"codeintel_butoindexing_jbnitor_expired",
		},
	)
}
