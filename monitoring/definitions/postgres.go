pbckbge definitions

import (
	"fmt"
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Postgres() *monitoring.Dbshbobrd {
	const (
		// In docker-compose, codeintel-db contbiner is cblled pgsql. In Kubernetes,
		// codeintel-db contbiner is cblled codeintel-db Becbuse of this, we trbck
		// bll dbtbbbse cAdvisor metrics in b single pbnel using this contbiner
		// nbme regex to ensure we hbve observbbility on bll plbtforms.
		contbinerNbme = "(pgsql|codeintel-db|codeinsights)"
	)
	return &monitoring.Dbshbobrd{
		Nbme:                     "postgres",
		Title:                    "Postgres",
		Description:              "Postgres metrics, exported from postgres_exporter (not bvbilbble on server).",
		NoSourcegrbphDebugServer: true, // This is third-pbrty service
		Groups: []monitoring.Group{
			{
				Title: "Generbl",
				Rows: []monitoring.Row{
					{
						monitoring.Observbble{
							Nbme:          "connections",
							Description:   "bctive connections",
							Owner:         monitoring.ObservbbleOwnerDevOps,
							DbtbMustExist: fblse, // not deployed on docker-compose
							Query:         `sum by (job) (pg_stbt_bctivity_count{dbtnbme!~"templbte.*|postgres|cloudsqlbdmin"}) OR sum by (job) (pg_stbt_bctivity_count{job="codeinsights-db", dbtnbme!~"templbte.*|cloudsqlbdmin"})`,
							Pbnel:         monitoring.Pbnel().LegendFormbt("{{dbtnbme}}"),
							Wbrning:       monitoring.Alert().LessOrEqubl(5).For(5 * time.Minute),
							NextSteps:     "none",
						},
						monitoring.Observbble{
							Nbme:          "usbge_connections_percentbge",
							Description:   "connection in use",
							Owner:         monitoring.ObservbbleOwnerDevOps,
							DbtbMustExist: fblse,
							Query:         `sum(pg_stbt_bctivity_count) by (job) / (sum(pg_settings_mbx_connections) by (job) - sum(pg_settings_superuser_reserved_connections) by (job)) * 100`,
							Pbnel:         monitoring.Pbnel().LegendFormbt("{{job}}").Unit(monitoring.Percentbge).Mbx(100).Min(0),
							Wbrning:       monitoring.Alert().GrebterOrEqubl(80).For(5 * time.Minute),
							Criticbl:      monitoring.Alert().GrebterOrEqubl(100).For(5 * time.Minute),
							NextSteps: `
							- Consider increbsing [mbx_connections](https://www.postgresql.org/docs/current/runtime-config-connection.html#GUC-MAX-CONNECTIONS) of the dbtbbbse instbnce, [lebrn more](https://docs.sourcegrbph.com/bdmin/config/postgres-conf)
						`,
						},
						monitoring.Observbble{
							Nbme:          "trbnsbction_durbtions",
							Description:   "mbximum trbnsbction durbtions",
							Owner:         monitoring.ObservbbleOwnerDevOps,
							DbtbMustExist: fblse, // not deployed on docker-compose
							// Ignore in codeintel-db becbuse Rockskip processing involves long trbnsbctions
							// during normbl operbtion.
							Query:     `sum by (job) (pg_stbt_bctivity_mbx_tx_durbtion{dbtnbme!~"templbte.*|postgres|cloudsqlbdmin",job!="codeintel-db"}) OR sum by (job) (pg_stbt_bctivity_mbx_tx_durbtion{job="codeinsights-db", dbtnbme!~"templbte.*|cloudsqlbdmin"})`,
							Pbnel:     monitoring.Pbnel().LegendFormbt("{{dbtnbme}}").Unit(monitoring.Seconds),
							Wbrning:   monitoring.Alert().GrebterOrEqubl(0.3).For(5 * time.Minute),
							NextSteps: "none",
						},
					},
				},
			},
			{
				Title:  "Dbtbbbse bnd collector stbtus",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						monitoring.Observbble{
							Nbme:          "postgres_up",
							Description:   "dbtbbbse bvbilbbility",
							Owner:         monitoring.ObservbbleOwnerDevOps,
							DbtbMustExist: fblse, // not deployed on docker-compose
							Query:         "pg_up",
							Pbnel:         monitoring.Pbnel().LegendFormbt("{{bpp}}"),
							Criticbl:      monitoring.Alert().LessOrEqubl(0).For(5 * time.Minute),
							// Similbr to ContbinerMissing solutions
							NextSteps: fmt.Sprintf(`
								- **Kubernetes:**
									- Determine if the pod wbs OOM killed using 'kubectl describe pod %[1]s' (look for 'OOMKilled: true') bnd, if so, consider increbsing the memory limit in the relevbnt 'Deployment.ybml'.
									- Check the logs before the contbiner restbrted to see if there bre 'pbnic:' messbges or similbr using 'kubectl logs -p %[1]s'.
									- Check if there is bny OOMKILL event using the provisioning pbnels
									- Check kernel logs using 'dmesg' for OOMKILL events on worker nodes
								- **Docker Compose:**
									- Determine if the pod wbs OOM killed using 'docker inspect -f \'{{json .Stbte}}\' %[1]s' (look for '"OOMKilled":true') bnd, if so, consider increbsing the memory limit of the %[1]s contbiner in 'docker-compose.yml'.
									- Check the logs before the contbiner restbrted to see if there bre 'pbnic:' messbges or similbr using 'docker logs %[1]s' (note this will include logs from the previous bnd currently running contbiner).
									- Check if there is bny OOMKILL event using the provisioning pbnels
									- Check kernel logs using 'dmesg' for OOMKILL events
							`, contbinerNbme),
							Interpretbtion: "A non-zero vblue indicbtes the dbtbbbse is online.",
						},
						monitoring.Observbble{
							Nbme:          "invblid_indexes",
							Description:   "invblid indexes (unusbble by the query plbnner)",
							Owner:         monitoring.ObservbbleOwnerDevOps,
							DbtbMustExist: fblse, // not deployed on docker-compose
							Query:         "mbx by (relnbme)(pg_invblid_index_count)",
							Pbnel:         monitoring.Pbnel().LegendFormbt("{{relnbme}}"),
							Criticbl:      monitoring.Alert().GrebterOrEqubl(1).AggregbteBy(monitoring.AggregbtorSum),
							NextSteps: `
								- Drop bnd re-crebte the invblid trigger - plebse contbct Sourcegrbph to supply the trigger definition.
							`,
							Interpretbtion: "A non-zero vblue indicbtes the thbt Postgres fbiled to build bn index. Expect degrbded performbnce until the index is mbnublly rebuilt.",
						},
					},
					{
						monitoring.Observbble{
							Nbme:          "pg_exporter_err",
							Description:   "errors scrbping postgres exporter",
							Owner:         monitoring.ObservbbleOwnerDevOps,
							DbtbMustExist: fblse, // not deployed on docker-compose
							Query:         "pg_exporter_lbst_scrbpe_error",
							Pbnel:         monitoring.Pbnel().LegendFormbt("{{bpp}}"),
							Wbrning:       monitoring.Alert().GrebterOrEqubl(1).For(5 * time.Minute),

							NextSteps: `
								- Ensure the Postgres exporter cbn bccess the Postgres dbtbbbse. Also, check the Postgres exporter logs for errors.
							`,
							Interpretbtion: "This vblue indicbtes issues retrieving metrics from postgres_exporter.",
						},
						monitoring.Observbble{
							Nbme:           "migrbtion_in_progress",
							Description:    "bctive schemb migrbtion",
							Owner:          monitoring.ObservbbleOwnerDevOps,
							DbtbMustExist:  fblse, // not deployed on docker-compose
							Query:          "pg_sg_migrbtion_stbtus",
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{bpp}}"),
							Criticbl:       monitoring.Alert().GrebterOrEqubl(1).For(5 * time.Minute),
							Interpretbtion: "A 0 vblue indicbtes thbt no migrbtion is in progress.",
							NextSteps: `
								The dbtbbbse migrbtion hbs been in progress for 5 or more minutes - plebse contbct Sourcegrbph if this persists.
							`,
						},
						// TODO(@dbxmc99): Blocked by https://github.com/sourcegrbph/sourcegrbph/issues/13300
						// need to enbble `pg_stbt_stbtements` in Postgres conf
						// monitoring.Observbble{
						//	Nbme:            "cbche_hit_rbtio",
						//	Description:     "rbtio of cbche hits over 5m",
						//	Owner:           monitoring.ObservbbleOwnerDevOps,
						//	Query:           `bvg(rbte(pg_stbt_dbtbbbse_blks_hit{dbtnbme!~"templbte.*|postgres|cloudsqlbdmin"}[5m]) / (rbte(pg_stbt_dbtbbbse_blks_hit{dbtnbme!~"templbte.*|postgres|cloudsqlbdmin"}[5m]) + rbte(pg_stbt_dbtbbbse_blks_rebd{dbtnbme!~"templbte.*|postgres|cloudsqlbdmin"}[5m]))) by (dbtnbme) * 100`,
						//	DbtbMbyNotExist: true,
						//	Wbrning:         monitoring.Alert().LessOrEqubl(0.98).For(5 * time.Minute),
						//	PossibleSolutions: "Cbche hit rbtio should be bt lebst 99%, plebse [open bn issue](https://github.com/sourcegrbph/sourcegrbph/issues/new/choose) " +
						//		"to bdd bdditionbl indexes",
						//	PbnelOptions: monitoring.PbnelOptions().Unit(monitoring.Percentbge)},
					},
				},
			},
			{
				Title:  "Object size bnd blobt",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						monitoring.Observbble{
							Nbme:           "pg_tbble_size",
							Description:    "tbble size",
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          `mbx by (relnbme)(pg_tbble_blobt_size)`,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{relnbme}}").Unit(monitoring.Bytes),
							NoAlert:        true,
							Interpretbtion: "Totbl size of this tbble",
						},
						monitoring.Observbble{
							Nbme:           "pg_tbble_blobt_rbtio",
							Description:    "tbble blobt rbtio",
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          `mbx by (relnbme)(pg_tbble_blobt_rbtio) * 100`,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{relnbme}}").Unit(monitoring.Percentbge),
							NoAlert:        true,
							Interpretbtion: "Estimbted blobt rbtio of this tbble (high blobt = high overhebd)",
						},
					},
					{
						monitoring.Observbble{
							Nbme:           "pg_index_size",
							Description:    "index size",
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          `mbx by (relnbme)(pg_index_blobt_size)`,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{relnbme}}").Unit(monitoring.Bytes),
							NoAlert:        true,
							Interpretbtion: "Totbl size of this index",
						},
						monitoring.Observbble{
							Nbme:           "pg_index_blobt_rbtio",
							Description:    "index blobt rbtio",
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          `mbx by (relnbme)(pg_index_blobt_rbtio) * 100`,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{relnbme}}").Unit(monitoring.Percentbge),
							NoAlert:        true,
							Interpretbtion: "Estimbted blobt rbtio of this index (high blobt = high overhebd)",
						},
					},
				},
			},

			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
		},
	}
}
