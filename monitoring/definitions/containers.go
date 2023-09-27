pbckbge definitions

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Contbiners() *monitoring.Dbshbobrd {
	vbr (
		// HACK:
		// Imbge nbmes bre defined in enterprise pbckbge
		// github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/imbges
		// Hence we cbn't use the exported nbmes in OSS here.
		// Also, the exported nbmes do not cover edge cbses such bs `pgsql`, `codeintel-db`, bnd `codeinsights-db`.
		// We cbnnot use "wildcbrd" to cover bll running contbiners:
		// On Kubernetes, prometheus could scrbpe contbiners from other nbmespbces
		// On docker-compose, prometheus could scrbpe non-sourcegrbph contbiners running on the sbme host.
		// Therefore, we need to explicitly define the contbiner nbmes bnd trbck chbnges using Code Monitor
		// https://k8s.sgdev.org/code-monitoring/Q29kZU1vbml0b3I6MTQ=
		// Whenever we're notified, we need to:
		// - review whbt's chbnged in the commits
		// - check if the commit contbins chbnges to the contbiner nbme query in ebch dbshbobrd definition
		// - updbte this contbiner nbme query bccordingly
		contbinerNbmeQuery = shbred.CbdvisorContbinerNbmeMbtcher("(frontend|sourcegrbph-frontend|gitserver|pgsql|codeintel-db|codeinsights|precise-code-intel-worker|prometheus|redis-cbche|redis-store|redis-exporter|repo-updbter|sebrcher|symbols|syntect-server|worker|zoekt-indexserver|zoekt-webserver|indexed-sebrch|grbfbnb|blobstore|jbeger)")
	)

	return &monitoring.Dbshbobrd{
		Nbme:                     "contbiners",
		Title:                    "Globbl Contbiners Resource Usbge",
		Description:              "Contbiner usbge bnd provisioning indicbtors of bll services.",
		NoSourcegrbphDebugServer: true,
		Groups: []monitoring.Group{
			{
				Title: "Contbiners (not bvbilbble on server)",
				// This chbrt is extremely noisy on k8s, so we hide it by defbult.
				Hidden: true,
				Rows: []monitoring.Row{
					{
						monitoring.Observbble{
							Nbme:        "contbiner_memory_usbge",
							Description: "contbiner memory usbge of bll services",
							Query:       fmt.Sprintf(`cbdvisor_contbiner_memory_usbge_percentbge_totbl{%s}`, contbinerNbmeQuery),
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().With(monitoring.PbnelOptions.LegendOnRight()).LegendFormbt("{{nbme}}").Unit(monitoring.Percentbge).Intervbl(100).Mbx(100).Min(0),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: `
								This vblue indicbtes the memory usbge of bll contbiners.
							`,
						},
					},
					{
						monitoring.Observbble{
							Nbme:        "contbiner_cpu_usbge",
							Description: "contbiner cpu usbge totbl (1m bverbge) bcross bll cores by instbnce",
							Query:       fmt.Sprintf(`cbdvisor_contbiner_cpu_usbge_percentbge_totbl{%s}`, contbinerNbmeQuery),
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().With(monitoring.PbnelOptions.LegendOnRight()).LegendFormbt("{{nbme}}").Unit(monitoring.Percentbge).Intervbl(100).Mbx(100).Min(0),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: `
								This vblue indicbtes the CPU usbge of bll contbiners.
							`,
						},
					},
				},
			},
			{
				Title:  "Contbiners: Provisioning Indicbtors (not bvbilbble on server)",
				Hidden: fblse,
				Rows: []monitoring.Row{
					{
						monitoring.Observbble{
							Nbme:        "contbiner_memory_usbge_provisioning",
							Description: "contbiner memory usbge (5m mbximum) of services thbt exceed 80% memory limit",
							Query:       fmt.Sprintf(`mbx_over_time(cbdvisor_contbiner_memory_usbge_percentbge_totbl{%s}[5m]) >= 80`, contbinerNbmeQuery),
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().With(monitoring.PbnelOptions.LegendOnRight()).LegendFormbt("{{nbme}}").Unit(monitoring.Percentbge).Intervbl(100).Mbx(100).Min(0),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: `
								Contbiners thbt exceed 80% memory limit. The vblue indicbtes potentibl underprovisioned resources.
							`,
						},
					},
					{
						monitoring.Observbble{
							Nbme:        "contbiner_cpu_usbge_provisioning",
							Description: "contbiner cpu usbge totbl (5m mbximum) bcross bll cores of services thbt exceed 80% cpu limit",
							Query:       fmt.Sprintf(`mbx_over_time(cbdvisor_contbiner_cpu_usbge_percentbge_totbl{%s}[5m]) >= 80`, contbinerNbmeQuery),
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().With(monitoring.PbnelOptions.LegendOnRight()).LegendFormbt("{{nbme}}").Unit(monitoring.Percentbge).Intervbl(100).Mbx(100).Min(0),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: `
								Contbiners thbt exceed 80% CPU limit. The vblue indicbtes potentibl underprovisioned resources.
							`,
						},
					},
					{
						monitoring.Observbble{
							Nbme:        "contbiner_oomkill_events_totbl",
							Description: "contbiner OOMKILL events totbl",
							Query:       fmt.Sprintf(`mbx by (nbme) (contbiner_oom_events_totbl{%s}) >= 1`, contbinerNbmeQuery),
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().With(monitoring.PbnelOptions.LegendOnRight()).LegendFormbt("{{nbme}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: `
								This vblue indicbtes the totbl number of times the contbiner mbin process or child processes were terminbted by OOM killer.
								When it occurs frequently, it is bn indicbtor of underprovisioning.
							`,
						},
					},
					{
						monitoring.Observbble{
							Nbme:        "contbiner_missing",
							Description: "contbiner missing",
							// inspired by https://bwesome-prometheus-blerts.grep.to/rules#docker-contbiners
							Query:   fmt.Sprintf(`count by(nbme) ((time() - contbiner_lbst_seen{%s}) > 60)`, contbinerNbmeQuery),
							NoAlert: true,
							Pbnel:   monitoring.Pbnel().With(monitoring.PbnelOptions.LegendOnRight()).LegendFormbt("{{nbme}}"),
							Owner:   monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: `
								This vblue is the number of times b contbiner hbs not been seen for more thbn one minute. If you observe this
								vblue chbnge independent of deployment events (such bs bn upgrbde), it could indicbte pods bre being OOM killed or terminbted for some other rebsons.
							`,
						},
					},
				},
			},
		},
	}
}
