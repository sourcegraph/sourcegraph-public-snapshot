package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

var FrontendInternalAPIErrorResponses sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
	return Observable{
		Name:        "frontend_internal_api_error_responses",
		Description: "frontend-internal API error responses every 5m by route",
		Query:       fmt.Sprintf(`sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="%[1]s",code!~"2.."}[5m])) / ignoring(category) group_left sum(increase(src_frontend_internal_request_duration_seconds_count{job="%[1]s"}[5m]))`, containerName),
		Warning:     monitoring.Alert().GreaterOrEqual(2).For(5 * time.Minute),
		Panel:       monitoring.Panel().LegendFormat("{{category}}").Unit(monitoring.Percentage),
		Owner:       owner,
		NextSteps: strings.ReplaceAll(`
			- **Single-container deployments:** Check 'docker logs $CONTAINER_ID' for logs starting with 'repo-updater' that indicate requests to the frontend service are failing.
			- **Kubernetes:**
				- Confirm that 'kubectl get pods' shows the 'frontend' pods are healthy.
				- Check 'kubectl logs {{CONTAINER_NAME}}' for logs indicate request failures to 'frontend' or 'frontend-internal'.
			- **Docker Compose:**
				- Confirm that 'docker ps' shows the 'frontend-internal' container is healthy.
				- Check 'docker logs {{CONTAINER_NAME}}' for logs indicating request failures to 'frontend' or 'frontend-internal'.
		`, "{{CONTAINER_NAME}}", containerName),
	}
}

type FrontendInternalAPIERrorResponseMonitoringOptions struct {
	// ErrorResponses transforms the default observable used to construct the error responses panel.
	ErrorResponses ObservableOption
}

// NewProvisioningIndicatorsGroup creates a group containing panels displaying
// internal API error response metrics for the given container.
func NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName string, owner monitoring.ObservableOwner, options *FrontendInternalAPIERrorResponseMonitoringOptions) monitoring.Group {
	if options == nil {
		options = &FrontendInternalAPIERrorResponseMonitoringOptions{}
	}

	return monitoring.Group{
		Title:  "Internal service requests",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				options.ErrorResponses.safeApply(FrontendInternalAPIErrorResponses(containerName, owner)).Observable(),
			},
		},
	}
}
