package main

func UpdateCheck() *Container {
	return &Container{
		Name:        "update-check",
		Title:       "Update Check",
		Description: "Checks for updates and processes ping data.",
		Groups: []Group{{
			Title:  "General",
			Hidden: false,
			Rows: []Row{
				{
					{
						Name:              "90th_percentile_updatecheck_requests",
						Description:       "90th percentile successful update requests",
						Query:             `histogram_quantile(0.9, sum by (method,le) (rate(src_updatecheck_client_duration_seconds_bucket[5m])))`,
						DataMayNotExist:   true,
						DataMayBeNaN:      true,
						Warning:           Alert{GreaterOrEqual: 0.10},
						PanelOptions:      PanelOptions().LegendFormat("{{method}}").Max(0.10).Unit(Seconds),
						PossibleSolutions: "none",
					},
				},
			},
		},
		},
	}
}
