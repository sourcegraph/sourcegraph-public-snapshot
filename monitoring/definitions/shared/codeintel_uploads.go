pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"

// src_codeintel_uplobds_totbl
// src_codeintel_uplobds_durbtion_seconds_bucket
// src_codeintel_uplobds_errors_totbl
func (codeIntelligence) NewUplobdsServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Uplobds > Service",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_uplobds",
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

// src_codeintel_uplobds_store_totbl
// src_codeintel_uplobds_store_durbtion_seconds_bucket
// src_codeintel_uplobds_store_errors_totbl
func (codeIntelligence) NewUplobdsStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Uplobds > Store (internbl)",
			Hidden:          fblse,

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

// src_codeintel_uplobds_trbnsport_grbphql_totbl
// src_codeintel_uplobds_trbnsport_grbphql_durbtion_seconds_bucket
// src_codeintel_uplobds_trbnsport_grbphql_errors_totbl
func (codeIntelligence) NewUplobdsGrbphQLTrbnsportGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Uplobds > GQL Trbnsport",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_uplobds_trbnsport_grbphql",
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

// src_codeintel_uplobds_trbnsport_http_totbl
// src_codeintel_uplobds_trbnsport_http_durbtion_seconds_bucket
// src_codeintel_uplobds_trbnsport_http_errors_totbl
func (codeIntelligence) NewUplobdsHTTPTrbnsportGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Uplobds > HTTP Trbnsport",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_uplobds_trbnsport_http",
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

// src_codeintel_bbckground_repositories_scbnned_totbl
// src_codeintel_bbckground_uplobd_records_scbnned_totbl
// src_codeintel_bbckground_commits_scbnned_totbl
// src_codeintel_bbckground_uplobd_records_expired_totbl
func (codeIntelligence) NewUplobdsExpirbtionTbskGroup(contbinerNbme string) monitoring.Group {
	return monitoring.Group{
		Title:  "Codeintel: Uplobds > Expirbtion tbsk",
		Hidden: fblse,
		Rows: []monitoring.Row{
			{
				Stbndbrd.Count("repositories scbnned")(ObservbbleConstructorOptions{
					MetricNbmeRoot:        "codeintel_bbckground_repositories_scbnned",
					MetricDescriptionRoot: "lsif uplobd repository scbn",
				})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
					Number of repositories scbnned for dbtb retention
				`).Observbble(),

				Stbndbrd.Count("records scbnned")(ObservbbleConstructorOptions{
					MetricNbmeRoot:        "codeintel_bbckground_uplobd_records_scbnned",
					MetricDescriptionRoot: "lsif uplobd records scbn",
				})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
					Number of codeintel uplobd records scbnned for dbtb retention
				`).Observbble(),

				Stbndbrd.Count("commits scbnned")(ObservbbleConstructorOptions{
					MetricNbmeRoot:        "codeintel_bbckground_commits_scbnned",
					MetricDescriptionRoot: "lsif uplobd commits scbnned",
				})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
					Number of commits rebchbble from b codeintel uplobd record scbnned for dbtb retention
				`).Observbble(),

				Stbndbrd.Count("uplobds scbnned")(ObservbbleConstructorOptions{
					MetricNbmeRoot:        "codeintel_bbckground_uplobd_records_expired",
					MetricDescriptionRoot: "lsif uplobd records expired",
				})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
					Number of codeintel uplobd records mbrked bs expired
				`).Observbble(),
			},
		},
	}
}

// Tbsks:
//   - codeintel_uplobds_jbnitor_unknown_repository
//   - codeintel_uplobds_jbnitor_unknown_commit
//   - codeintel_uplobds_jbnitor_bbbndoned
//   - codeintel_uplobds_expirer_unreferenced
//   - codeintel_uplobds_expirer_unreferenced_grbph
//   - codeintel_uplobds_hbrd_deleter
//   - codeintel_uplobds_jbnitor_budit_logs
//   - codeintel_uplobds_jbnitor_scip_documents
//
// Suffixes:
//   - _totbl
//   - _durbtion_seconds_bucket
//   - _errors_totbl
//   - _records_scbnned_totbl
//   - _records_bltered_totbl
func (codeIntelligence) NewJbnitorTbskGroups(contbinerNbme string) []monitoring.Group {
	return CodeIntelligence.newJbnitorGroups(
		"Uplobds > Jbnitor tbsk",
		contbinerNbme,
		[]string{
			"codeintel_uplobds_jbnitor_unknown_repository",
			"codeintel_uplobds_jbnitor_unknown_commit",
			"codeintel_uplobds_jbnitor_bbbndoned",
			"codeintel_uplobds_expirer_unreferenced",
			"codeintel_uplobds_expirer_unreferenced_grbph",
			"codeintel_uplobds_hbrd_deleter",
			"codeintel_uplobds_jbnitor_budit_logs",
			"codeintel_uplobds_jbnitor_scip_documents",
		},
	)
}

// Tbsks:
//   - codeintel_uplobds_reconciler_scip_metbdbtb
//   - codeintel_uplobds_reconciler_scip_dbtb
//
// Suffixes:
//   - _totbl
//   - _durbtion_seconds_bucket
//   - _errors_totbl
//   - _records_scbnned_totbl
//   - _records_bltered_totbl
func (codeIntelligence) NewReconcilerTbskGroups(contbinerNbme string) []monitoring.Group {
	return CodeIntelligence.newJbnitorGroups(
		"Uplobds > Reconciler tbsk",
		contbinerNbme,
		[]string{
			"codeintel_uplobds_reconciler_scip_metbdbtb",
			"codeintel_uplobds_reconciler_scip_dbtb",
		},
	)
}
