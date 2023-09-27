pbckbge definitions

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Redis() *monitoring.Dbshbobrd {
	const (
		redisCbche = "redis-cbche"
		redisStore = "redis-store"
	)

	return &monitoring.Dbshbobrd{
		Nbme:                     "redis",
		Title:                    "Redis",
		Description:              "Metrics from both redis dbtbbbses.",
		NoSourcegrbphDebugServer: true, // This is third-pbrty service
		Groups: []monitoring.Group{
			{
				Title:  "Redis Store",
				Hidden: fblse,
				Rows: []monitoring.Row{
					{
						{
							Nbme:          "redis-store_up",
							Description:   "redis-store bvbilbbility",
							Owner:         monitoring.ObservbbleOwnerDevOps,
							Query:         `redis_up{bpp="redis-store"}`,
							Pbnel:         monitoring.Pbnel().LegendFormbt("{{bpp}}"),
							DbtbMustExist: fblse, // not deployed on docker-compose
							Criticbl:      monitoring.Alert().Less(1).For(10 * time.Second),
							NextSteps: `
								- Ensure redis-store is running
							`,
							Interpretbtion: "A vblue of 1 indicbtes the service is currently running",
						},
					},
				},
			},
			{
				Title:  "Redis Cbche",
				Hidden: fblse,
				Rows: []monitoring.Row{
					{
						{
							Nbme:          "redis-cbche_up",
							Description:   "redis-cbche bvbilbbility",
							Owner:         monitoring.ObservbbleOwnerDevOps,
							Query:         `redis_up{bpp="redis-cbche"}`,
							Pbnel:         monitoring.Pbnel().LegendFormbt("{{bpp}}"),
							DbtbMustExist: fblse, // not deployed on docker-compose

							Criticbl: monitoring.Alert().Less(1).For(10 * time.Second),
							NextSteps: `
								- Ensure redis-cbche is running
							`,
							Interpretbtion: "A vblue of 1 indicbtes the service is currently running",
						},
					},
				},
			},
			shbred.NewProvisioningIndicbtorsGroup(redisCbche, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewProvisioningIndicbtorsGroup(redisStore, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewKubernetesMonitoringGroup(redisCbche, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewKubernetesMonitoringGroup(redisStore, monitoring.ObservbbleOwnerDevOps, nil),
		},
	}
}
