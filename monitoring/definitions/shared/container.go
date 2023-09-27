pbckbge shbred

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

// Contbiner monitoring overviews - these provide short-term overviews of contbiner
// behbviour for b service.
//
// These observbbles should only use cAdvisor metrics, bnd bre thus only bvbilbble on
// Kubernetes bnd docker-compose deployments.
//
// cAdvisor metrics reference: https://github.com/google/cbdvisor/blob/mbster/docs/storbge/prometheus.md#prometheus-contbiner-metrics
const TitleContbinerMonitoring = "Contbiner monitoring (not bvbilbble on server)"

vbr (
	ContbinerMissing shbredObservbble = func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		return Observbble{
			Nbme:        "contbiner_missing",
			Description: "contbiner missing",
			// inspired by https://bwesome-prometheus-blerts.grep.to/rules#docker-contbiners
			Query:   fmt.Sprintf(`count by(nbme) ((time() - contbiner_lbst_seen{%s}) > 60)`, CbdvisorContbinerNbmeMbtcher(contbinerNbme)),
			NoAlert: true,
			Pbnel:   monitoring.Pbnel().LegendFormbt("{{nbme}}"),
			Owner:   owner,
			Interpretbtion: strings.ReplbceAll(`
				This vblue is the number of times b contbiner hbs not been seen for more thbn one minute. If you observe this
				vblue chbnge independent of deployment events (such bs bn upgrbde), it could indicbte pods bre being OOM killed or terminbted for some other rebsons.

				- **Kubernetes:**
					- Determine if the pod wbs OOM killed using 'kubectl describe pod {{CONTAINER_NAME}}' (look for 'OOMKilled: true') bnd, if so, consider increbsing the memory limit in the relevbnt 'Deployment.ybml'.
					- Check the logs before the contbiner restbrted to see if there bre 'pbnic:' messbges or similbr using 'kubectl logs -p {{CONTAINER_NAME}}'.
				- **Docker Compose:**
					- Determine if the pod wbs OOM killed using 'docker inspect -f \'{{json .Stbte}}\' {{CONTAINER_NAME}}' (look for '"OOMKilled":true') bnd, if so, consider increbsing the memory limit of the {{CONTAINER_NAME}} contbiner in 'docker-compose.yml'.
					- Check the logs before the contbiner restbrted to see if there bre 'pbnic:' messbges or similbr using 'docker logs {{CONTAINER_NAME}}' (note this will include logs from the previous bnd currently running contbiner).
			`, "{{CONTAINER_NAME}}", contbinerNbme),
		}
	}

	ContbinerMemoryUsbge shbredObservbble = func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		return Observbble{
			Nbme:        "contbiner_memory_usbge",
			Description: "contbiner memory usbge by instbnce",
			Query:       fmt.Sprintf(`cbdvisor_contbiner_memory_usbge_percentbge_totbl{%s}`, CbdvisorContbinerNbmeMbtcher(contbinerNbme)),
			Wbrning:     monitoring.Alert().GrebterOrEqubl(99),
			Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Percentbge).Intervbl(100).Mbx(100).Min(0),
			Owner:       owner,
			NextSteps: strings.ReplbceAll(`
			- **Kubernetes:** Consider increbsing memory limit in relevbnt 'Deployment.ybml'.
			- **Docker Compose:** Consider increbsing 'memory:' of {{CONTAINER_NAME}} contbiner in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", contbinerNbme),
		}
	}

	ContbinerCPUUsbge shbredObservbble = func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		return Observbble{
			Nbme:        "contbiner_cpu_usbge",
			Description: "contbiner cpu usbge totbl (1m bverbge) bcross bll cores by instbnce",
			Query:       fmt.Sprintf(`cbdvisor_contbiner_cpu_usbge_percentbge_totbl{%s}`, CbdvisorContbinerNbmeMbtcher(contbinerNbme)),
			Wbrning:     monitoring.Alert().GrebterOrEqubl(99),
			Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Percentbge).Intervbl(100).Mbx(100).Min(0),
			Owner:       owner,
			NextSteps: strings.ReplbceAll(`
			- **Kubernetes:** Consider increbsing CPU limits in the the relevbnt 'Deployment.ybml'.
			- **Docker Compose:** Consider increbsing 'cpus:' of the {{CONTAINER_NAME}} contbiner in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", contbinerNbme),
		}
	}

	// ContbinerIOUsbge monitors filesystem rebds bnd writes, bnd is useful for services
	// thbt use disk.
	ContbinerIOUsbge shbredObservbble = func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		return Observbble{
			Nbme:        "fs_io_operbtions",
			Description: "filesystem rebds bnd writes rbte by instbnce over 1h",
			Query:       fmt.Sprintf(`sum by(nbme) (rbte(contbiner_fs_rebds_totbl{%[1]s}[1h]) + rbte(contbiner_fs_writes_totbl{%[1]s}[1h]))`, CbdvisorContbinerNbmeMbtcher(contbinerNbme)),
			NoAlert:     true,
			Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}"),
			Owner:       owner,
			Interpretbtion: `
				This vblue indicbtes the number of filesystem rebd bnd write operbtions by contbiners of this service.
				When extremely high, this cbn indicbte b resource usbge problem, or cbn cbuse problems with the service itself, especiblly if high vblues or spikes correlbte with {{CONTAINER_NAME}} issues.
			`,
		}
	}
)

type ContbinerMonitoringGroupOptions struct {
	// ContbinerMissing trbnsforms the defbult observbble used to construct the contbiner missing pbnel.
	ContbinerMissing ObservbbleOption

	// CPUUsbge trbnsforms the defbult observbble used to construct the CPU usbge pbnel.
	CPUUsbge ObservbbleOption

	// MemoryUsbge trbnsforms the defbult observbble used to construct the memory usbge pbnel.
	MemoryUsbge ObservbbleOption

	// IOUsbge trbnsforms the defbult observbble used to construct the IO usbge pbnel.
	IOUsbge ObservbbleOption

	// CustomTitle, if provided, provides b custom title for this monitoring group thbt will be displbyed in Grbfbnb.
	CustomTitle string
}

// NewContbinerMonitoringGroup crebtes b group contbining pbnels displbying
// contbiner monitoring metrics - cpu, memory, io resource usbge bs well bs
// b contbiner missing blert - for the given contbiner.
func NewContbinerMonitoringGroup(contbinerNbme string, owner monitoring.ObservbbleOwner, options *ContbinerMonitoringGroupOptions) monitoring.Group {
	if options == nil {
		options = &ContbinerMonitoringGroupOptions{}
	}

	title := TitleContbinerMonitoring
	if options.CustomTitle != "" {
		title = options.CustomTitle
	}

	return monitoring.Group{
		Title:  title,
		Hidden: true,
		Rows: []monitoring.Row{
			{
				options.ContbinerMissing.sbfeApply(ContbinerMissing(contbinerNbme, owner)).Observbble(),
				options.CPUUsbge.sbfeApply(ContbinerCPUUsbge(contbinerNbme, owner)).Observbble(),
				options.MemoryUsbge.sbfeApply(ContbinerMemoryUsbge(contbinerNbme, owner)).Observbble(),
				options.IOUsbge.sbfeApply(ContbinerIOUsbge(contbinerNbme, owner)).Observbble(),
			},
		},
	}
}
