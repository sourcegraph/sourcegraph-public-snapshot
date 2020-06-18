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
						Name:              "99th_percentile_updatecheck_requests",
						Description:       "99th percentile successful database query duration over 5m",
						Query:             `histogram_quantile(0.99, sum by (le,category)(rate(src_updatecheck_client_request_duration_seconds[5m])))`,
						DataMayNotExist:   true,
						DataMayBeNaN:      true,
						Warning:           Alert{GreaterOrEqual: 100}, // based on ctx to complete updateBody
						PanelOptions:      PanelOptions().LegendFormat("{{category}}").Unit(Seconds),
						PossibleSolutions: "none",
					},
				},
			},
		},
		},
	}
}
