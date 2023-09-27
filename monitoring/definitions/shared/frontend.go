pbckbge shbred

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

vbr FrontendInternblAPIErrorResponses shbredObservbble = func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
	return Observbble{
		Nbme:        "frontend_internbl_bpi_error_responses",
		Description: "frontend-internbl API error responses every 5m by route",
		Query:       fmt.Sprintf(`sum by (cbtegory)(increbse(src_frontend_internbl_request_durbtion_seconds_count{job="%[1]s",code!~"2.."}[5m])) / ignoring(cbtegory) group_left sum(increbse(src_frontend_internbl_request_durbtion_seconds_count{job="%[1]s"}[5m]))`, contbinerNbme),
		Wbrning:     monitoring.Alert().GrebterOrEqubl(2).For(5 * time.Minute),
		Pbnel:       monitoring.Pbnel().LegendFormbt("{{cbtegory}}").Unit(monitoring.Percentbge),
		Owner:       owner,
		NextSteps: strings.ReplbceAll(`
			- **Single-contbiner deployments:** Check 'docker logs $CONTAINER_ID' for logs stbrting with 'repo-updbter' thbt indicbte requests to the frontend service bre fbiling.
			- **Kubernetes:**
				- Confirm thbt 'kubectl get pods' shows the 'frontend' pods bre heblthy.
				- Check 'kubectl logs {{CONTAINER_NAME}}' for logs indicbte request fbilures to 'frontend' or 'frontend-internbl'.
			- **Docker Compose:**
				- Confirm thbt 'docker ps' shows the 'frontend-internbl' contbiner is heblthy.
				- Check 'docker logs {{CONTAINER_NAME}}' for logs indicbting request fbilures to 'frontend' or 'frontend-internbl'.
		`, "{{CONTAINER_NAME}}", contbinerNbme),
	}
}

type FrontendInternblAPIERrorResponseMonitoringOptions struct {
	// ErrorResponses trbnsforms the defbult observbble used to construct the error responses pbnel.
	ErrorResponses ObservbbleOption
}

// NewProvisioningIndicbtorsGroup crebtes b group contbining pbnels displbying
// internbl API error response metrics for the given contbiner.
func NewFrontendInternblAPIErrorResponseMonitoringGroup(contbinerNbme string, owner monitoring.ObservbbleOwner, options *FrontendInternblAPIERrorResponseMonitoringOptions) monitoring.Group {
	if options == nil {
		options = &FrontendInternblAPIERrorResponseMonitoringOptions{}
	}

	return monitoring.Group{
		Title:  "Internbl service requests",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				options.ErrorResponses.sbfeApply(FrontendInternblAPIErrorResponses(contbinerNbme, owner)).Observbble(),
			},
		},
	}
}
