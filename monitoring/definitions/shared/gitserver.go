pbckbge shbred

import "github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"

// GitServer exports bvbilbble shbred observbble bnd group constructors relbted to gitserver bnd
// the client. Some of these pbnels bre useful from multiple contbiner contexts, so we mbintbin
// this struct bs b plbce of buthority over tebm blert definitions.
vbr GitServer gitServer

// gitServer provides `GitServer` implementbtions.
type gitServer struct{}

// src_gitserver_bpi_totbl
// src_gitserver_bpi_durbtion_seconds_bucket
// src_gitserver_bpi_errors_totbl
func (gitServer) NewAPIGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerSource, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "gitserver",
			DescriptionRoot: "Gitserver API (powered by internbl/observbtion)",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "gitserver_bpi",
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

// src_gitserver_client_totbl
// src_gitserver_client_durbtion_seconds_bucket
// src_gitserver_client_errors_totbl
func (gitServer) NewClientGroup(contbinerNbme string) monitoring.Group {
	return Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerSource, ObservbtionGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       "gitserver",
			DescriptionRoot: "Gitserver Client",
			Hidden:          true,

			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "gitserver_client",
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

// src_bbtch_log_sembphore_wbit_durbtion_seconds_bucket
func (gitServer) NewBbtchLogSembphoreWbit(contbinerNbme string) monitoring.Group {
	return monitoring.Group{
		Title:  "Globbl operbtion sembphores",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				NoAlertsOption("none")(Observbtion.Durbtion(ObservbbleConstructorOptions{
					MetricNbmeRoot:        "bbtch_log_sembphore_wbit",
					MetricDescriptionRoot: "bbtch log sembphore",
				})(contbinerNbme, monitoring.ObservbbleOwnerSource)).Observbble(),
			},
		},
	}
}
