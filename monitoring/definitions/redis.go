package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func RedisStore() *monitoring.Container {
	const (
		containerName = "redis-store"
	)

	return &monitoring.Container{
		Name:        "redis-store",
		Title:       "Redis Store",
		Description: "Holds data that cannot be easily recomputed, like sessions.",
		Templates:   nil,
		Groups: []monitoring.Group{
			{Title: "Redis Up",
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
					},
				}},
		},

		NoSourcegraphDebugServer: false,
	}
}

func RedisCache() *monitoring.Container {
	const (
		containerName = "redis-cache"
	)

	return &monitoring.Container{
		Name:        "redis-cache",
		Title:       "Redis Cache",
		Description: "Holds data that can be easily recomputed.",
		Groups: []monitoring.Group{
			{Title: "Redis Cache Up",
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
					},
				}},
		},
		NoSourcegraphDebugServer: false,
	}
}
