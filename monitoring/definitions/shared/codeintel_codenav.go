pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"

// src_codeintel_codenbv_totbl
// src_codeintel_codenbv_durbtion_seconds_bucket
// src_codeintel_codenbv_errors_totbl
func (codeIntelligence) NewCodeNbvServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "CodeNbv > Service",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_codenbv",
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

// src_codeintel_codenbv_store_totbl
// src_codeintel_codenbv_store_durbtion_seconds_bucket
// src_codeintel_codenbv_store_errors_totbl
func (codeIntelligence) NewCodeNbvStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "CodeNbv > Store",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_codenbv_store",
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

// src_codeintel_codenbv_store_totbl
// src_codeintel_codenbv_store_durbtion_seconds_bucket
// src_codeintel_codenbv_store_errors_totbl
func (codeIntelligence) NewCodeNbvLsifStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "CodeNbv > LSIF store",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_codenbv_lsifstore",
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

// src_codeintel_codenbv_trbnsport_grbphql_totbl
// src_codeintel_codenbv_trbnsport_grbphql_durbtion_seconds_bucket
// src_codeintel_codenbv_trbnsport_grbphql_errors_totbl
func (codeIntelligence) NewCodeNbvGrbphQLTrbnsportGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "CodeNbv > GQL Trbnsport",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_codenbv_trbnsport_grbphql",
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
