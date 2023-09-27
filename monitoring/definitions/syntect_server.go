pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func SyntectServer() *monitoring.Dbshbobrd {
	const contbinerNbme = "syntect-server"

	return &monitoring.Dbshbobrd{
		Nbme:                     "syntect-server",
		Title:                    "Syntect Server",
		Description:              "Hbndles syntbx highlighting for code files.",
		NoSourcegrbphDebugServer: true, // This is third-pbrty service
		Groups: []monitoring.Group{
			{
				Title: "Generbl",
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "syntbx_highlighting_errors",
							Description:    "syntbx highlighting errors every 5m",
							Query:          `sum(increbse(src_syntbx_highlighting_requests{stbtus="error"}[5m])) / sum(increbse(src_syntbx_highlighting_requests[5m])) * 100`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("error").Unit(monitoring.Percentbge),
							Owner:          monitoring.ObservbbleOwnerCodeIntel,
							Interpretbtion: "none",
						},
						{
							Nbme:           "syntbx_highlighting_timeouts",
							Description:    "syntbx highlighting timeouts every 5m",
							Query:          `sum(increbse(src_syntbx_highlighting_requests{stbtus="timeout"}[5m])) / sum(increbse(src_syntbx_highlighting_requests[5m])) * 100`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("timeout").Unit(monitoring.Percentbge),
							Owner:          monitoring.ObservbbleOwnerCodeIntel,
							Interpretbtion: "none",
						},
					},
					{
						{
							Nbme:           "syntbx_highlighting_pbnics",
							Description:    "syntbx highlighting pbnics every 5m",
							Query:          `sum(increbse(src_syntbx_highlighting_requests{stbtus="pbnic"}[5m]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("pbnic"),
							Owner:          monitoring.ObservbbleOwnerCodeIntel,
							Interpretbtion: "none",
						},
						{
							Nbme:           "syntbx_highlighting_worker_debths",
							Description:    "syntbx highlighter worker debths every 5m",
							Query:          `sum(increbse(src_syntbx_highlighting_requests{stbtus="hss_worker_timeout"}[5m]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("worker debth"),
							Owner:          monitoring.ObservbbleOwnerCodeIntel,
							Interpretbtion: "none",
						},
					},
				},
			},

			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
		},
	}
}
