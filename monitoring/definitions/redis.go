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
						{Name: "redis-store_up",
							Description: "determination if redis-store is currently alive", Owner: monitoring.ObservableOwnerDevOps,
							Query:         `redis_up{app="redis-store"}`,
							DataMustExist: true,
							Panel:         monitoring.Panel().LegendFormat("{{app}}"),
							Critical:      monitoring.Alert().LessOrEqual(0, nil).For(1 * time.Minute),
							PossibleSolutions: `
							-- Ensure redis-store is  running`},
					}},
			},
			{
				Title:  "Redis Cache",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{Name: "redis-cache_up",
							Description: "determination if redis-store is currently alive", Owner: monitoring.ObservableOwnerDevOps,
							Query:         `redis_up{app="redis-cache"}`,
							Panel:         monitoring.Panel().LegendFormat("{{app}}"),
							DataMustExist: true,
							Critical:      monitoring.Alert().LessOrEqual(0, nil).For(1 * time.Minute),
							PossibleSolutions: `
							-- Ensure redis-cache is running`},
					}}},
		},

		NoSourcegraphDebugServer: false,
	}
}
