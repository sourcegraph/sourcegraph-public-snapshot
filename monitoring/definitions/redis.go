package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Redis() *monitoring.Dashboard {
	const (
		redisCache = "redis-cache"
		redisStore = "redis-store"
	)

	return &monitoring.Dashboard{
		Name:                     "redis",
		Title:                    "Redis",
		Description:              "Metrics from both redis databases.",
		NoSourcegraphDebugServer: true, // This is third-party service
		Groups: []monitoring.Group{
			{
				Title:  "Redis Store",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:          "redis-store_up",
							Description:   "redis-store availability",
							Owner:         monitoring.ObservableOwnerDevOps,
							Query:         `redis_up{app="redis-store"}`,
							Panel:         monitoring.Panel().LegendFormat("{{app}}"),
							DataMustExist: false, // not deployed on docker-compose
							Critical:      monitoring.Alert().Less(1).For(10 * time.Second),
							NextSteps: `
								- Ensure redis-store is running
							`,
							Interpretation: "A value of 1 indicates the service is currently running",
						},
					},
				},
			},
			{
				Title:  "Redis Cache",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						{
							Name:          "redis-cache_up",
							Description:   "redis-cache availability",
							Owner:         monitoring.ObservableOwnerDevOps,
							Query:         `redis_up{app="redis-cache"}`,
							Panel:         monitoring.Panel().LegendFormat("{{app}}"),
							DataMustExist: false, // not deployed on docker-compose

							Critical: monitoring.Alert().Less(1).For(10 * time.Second),
							NextSteps: `
								- Ensure redis-cache is running
							`,
							Interpretation: "A value of 1 indicates the service is currently running",
						},
					},
				},
			},
			shared.NewProvisioningIndicatorsGroup(redisCache, monitoring.ObservableOwnerDevOps, nil),
			shared.NewProvisioningIndicatorsGroup(redisStore, monitoring.ObservableOwnerDevOps, nil),
			shared.NewKubernetesMonitoringGroup(redisCache, monitoring.ObservableOwnerDevOps, nil),
			shared.NewKubernetesMonitoringGroup(redisStore, monitoring.ObservableOwnerDevOps, nil),
		},
	}
}
