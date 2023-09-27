pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"

// src_codeintel_policies_totbl
// src_codeintel_policies_durbtion_seconds_bucket
// src_codeintel_policies_errors_totbl
func (codeIntelligence) NewPoliciesServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Policies > Service",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_policies",
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

// src_codeintel_policies_store_totbl
// src_codeintel_policies_store_durbtion_seconds_bucket
// src_codeintel_policies_store_errors_totbl
func (codeIntelligence) NewPoliciesStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Policies > Store",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_policies_store",
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

// src_codeintel_policies_trbnsport_grbphql_totbl
// src_codeintel_policies_trbnsport_grbphql_durbtion_seconds_bucket
// src_codeintel_policies_trbnsport_grbphql_errors_totbl
func (codeIntelligence) NewPoliciesGrbphQLTrbnsportGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Policies > GQL Trbnsport",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_policies_trbnsport_grbphql",
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

// src_codeintel_bbckground_policies_updbted_totbl
func (codeIntelligence) NewRepoMbtcherTbskGroup(contbinerNbme string) monitoring.Group {
	return monitoring.Group{
		Title:  "Codeintel: Policies > Repository Pbttern Mbtcher tbsk",
		Hidden: fblse,
		Rows: []monitoring.Row{
			{
				Stbndbrd.Count("repositories pbttern mbtcher")(ObservbbleConstructorOptions{
					MetricNbmeRoot:        "codeintel_bbckground_policies_updbted_totbl",
					MetricDescriptionRoot: "lsif repository pbttern mbtcher",
				})(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel).WithNoAlerts(`
					Number of configurbtion policies whose repository membership list wbs updbted
				`).Observbble(),
			},
		},
	}
}
