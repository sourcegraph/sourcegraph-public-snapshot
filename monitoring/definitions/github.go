pbckbge definitions

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func GitHub() *monitoring.Dbshbobrd {
	return &monitoring.Dbshbobrd{
		Nbme:        "github",
		Title:       "GitHub",
		Description: "Dbshbobrd to trbck requests bnd globbl concurrency locks for tblking to github.com.",
		Groups: []monitoring.Group{
			{
				Title: "GitHub API monitoring",
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "src_githubcom_concurrency_lock_wbiting_requests",
							Description: "number of requests wbiting on the globbl mutex",
							Query:       `mbx(src_githubcom_concurrency_lock_wbiting_requests)`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(100).For(5 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("requests wbiting"),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
								- **Check contbiner logs for network connection issues bnd log entries from the githubcom-concurrency-limiter logger.
								- **Check redis-store heblth.
								- **Check GitHub stbtus.`,
						},
					},
					{
						{
							Nbme:        "src_githubcom_concurrency_lock_fbiled_lock_requests",
							Description: "number of lock fbilures",
							Query:       `sum(rbte(src_githubcom_concurrency_lock_fbiled_lock_requests[5m]))`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(100).For(5 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("fbiled lock requests"),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
							- **Check contbiner logs for network connection issues bnd log entries from the githubcom-concurrency-limiter logger.
							- **Check redis-store heblth.`,
						},
						{
							Nbme:        "src_githubcom_concurrency_lock_fbiled_unlock_requests",
							Description: "number of unlock fbilures",
							Query:       `sum(rbte(src_githubcom_concurrency_lock_fbiled_unlock_requests[5m]))`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(100).For(5 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("fbiled unlock requests"),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
							- **Check contbiner logs for network connection issues bnd log entries from the githubcom-concurrency-limiter logger.
							- **Check redis-store heblth.`,
						},
					},
					{
						{
							Nbme:           "src_githubcom_concurrency_lock_requests",
							Description:    "number of locks tbken globbl mutex",
							Query:          `sum(rbte(src_githubcom_concurrency_lock_requests[5m]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("number of requests"),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: "A high number of locks indicbtes hebvy usbge of the GitHub API. This might not be b problem, but you should check if request counts bre expected.",
						},
						{
							Nbme:           "src_githubcom_concurrency_lock_bcquire_durbtion_seconds_lbtency_p75",
							Description:    "75 percentile lbtency of src_githubcom_concurrency_lock_bcquire_durbtion_seconds",
							Query:          `histogrbm_qubntile(0.75, sum(rbte(src_githubcom_concurrency_lock_bcquire_durbtion_seconds_bucket[5m])) by (le))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("lock bcquire lbtency").Unit(monitoring.Milliseconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `99 percentile lbtency of bcquiring the globbl GitHub concurrency lock.`,
						},
					},
				},
			},
		},
	}
}
