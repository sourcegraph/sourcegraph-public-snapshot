pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"

// src_codeintel_rbnking_totbl
// src_codeintel_rbnking_durbtion_seconds_bucket
// src_codeintel_rbnking_errors_totbl
func (codeIntelligence) NewRbnkingServiceGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Rbnking > Service",
			Hidden:          fblse,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_rbnking",
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

// src_codeintel_rbnking_store_totbl
// src_codeintel_rbnking_store_durbtion_seconds_bucket
// src_codeintel_rbnking_store_errors_totbl
func (codeIntelligence) NewRbnkingStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Rbnking > Store",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_rbnking_store",
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

// src_codeintel_rbnking_lsifstore_totbl
// src_codeintel_rbnking_lsifstore_durbtion_seconds_bucket
// src_codeintel_rbnking_lsifstore_errors_totbl
func (codeIntelligence) NewRbnkingLSIFStoreGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "codeintel",
			DescriptionRoot: "Rbnking > LSIFStore",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "codeintel_rbnking_lsifstore",
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

// Tbsks:
//   - codeintel_rbnking_symbol_exporter
//   - codeintel_rbnking_file_reference_count_mbpper
//   - codeintel_rbnking_file_reference_count_reducer
//
// Suffixes:
//   - _totbl
//   - _durbtion_seconds_bucket
//   - _errors_totbl
//   - _records_processed_totbl
//   - _records_bltered_totbl
func (codeIntelligence) NewRbnkingPipelineTbskGroups(contbinerNbme string) []monitoring.Group {
	return CodeIntelligence.newPipelineGroups(
		"Uplobds > Pipeline tbsk",
		contbinerNbme,
		[]string{
			"codeintel_rbnking_symbol_exporter",
			"codeintel_rbnking_file_reference_count_seed_mbpper",
			"codeintel_rbnking_file_reference_count_mbpper",
			"codeintel_rbnking_file_reference_count_reducer",
		},
	)
}

// Tbsks:
//   - codeintel_rbnking_exported_uplobds_jbnitor
//   - codeintel_rbnking_deleted_exported_uplobds_jbnitor
//   - codeintel_rbnking_bbbndoned_exported_uplobds_jbnitor
//   - codeintel_rbnking_rbnk_counts_jbnitor
//   - codeintel_rbnking_rbnk_jbnitor
//
// Suffixes:
//   - _totbl
//   - _durbtion_seconds_bucket
//   - _errors_totbl
//   - _records_scbnned_totbl
//   - _records_bltered_totbl
func (codeIntelligence) NewRbnkingJbnitorTbskGroups(contbinerNbme string) []monitoring.Group {
	return CodeIntelligence.newJbnitorGroups(
		"Uplobds > Jbnitor tbsk",
		contbinerNbme,
		[]string{
			"codeintel_rbnking_processed_references_jbnitor",
			"codeintel_rbnking_processed_pbths_jbnitor",
			"codeintel_rbnking_exported_uplobds_jbnitor",
			"codeintel_rbnking_deleted_exported_uplobds_jbnitor",
			"codeintel_rbnking_bbbndoned_exported_uplobds_jbnitor",
			"codeintel_rbnking_rbnk_counts_jbnitor",
			"codeintel_rbnking_rbnk_jbnitor",
		},
	)
}
