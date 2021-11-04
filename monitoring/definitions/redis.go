package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Redis() *monitoring.Container {

	return &monitoring.Container{
		Name:        "redis",
		Title:       "Redis",
		Description: "Metrics from both redis databases.",
		Templates:   nil,
		Groups: []monitoring.Group{
			{Title: "Redis Store",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name: "redis-store_up",
							Description: "redis-store up", Owner: monitoring.ObservableOwnerDevOps,
							Query:         `redis_up{app="redis-store"}`,
							DataMustExist: true,
							Panel:         monitoring.Panel().LegendFormat("{{app}}"),
							Critical:      monitoring.Alert().LessOrEqual(1, nil).For(10 * time.Second),
							PossibleSolutions: `
								- Ensure redis-store is running.
							`},
					}},
			},
			{
				Title:  "Redis Cache",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name: "redis-cache_up",
							Description: "redis-cache availability",
							Owner: monitoring.ObservableOwnerDevOps,
							Query:         `redis_up{app="redis-cache"}`,
							Panel:         monitoring.Panel().LegendFormat("{{app}}"),
							DataMustExist: true,
							Critical:      monitoring.Alert().LessOrEqual(1, nil).For(10 * time.Second),
							PossibleSolutions: `
								- Ensure redis-cache is running
							`},
					}}},
		},

		NoSourcegraphDebugServer: false,
	}
}
