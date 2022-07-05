package shared

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Kubernetes monitoring overviews.
//
// These observables should only use metrics exported by Kubernetes, or use Prometheus
// metrics in a way that only applies in Kubernetes deployments.
const TitleKubernetesMonitoring = "Kubernetes monitoring (only available on Kubernetes)"

var (
	KubernetesPodsAvailable sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "pods_available_percentage",
			Description: "percentage pods available",
			// the 'app' label is only available in Kubernetes deloyments - it indicates the pod.
			Query:    fmt.Sprintf(`sum by(app) (up{app=~".*%[1]s"}) / count by (app) (up{app=~".*%[1]s"}) * 100`, containerName),
			Critical: monitoring.Alert().LessOrEqual(90).For(10 * time.Minute),
			Panel:    monitoring.Panel().LegendFormat("{{name}}").Unit(monitoring.Percentage).Max(100).Min(0),
			Owner:    owner,
			// Solutions similar to the ContainerMissing solutions.
			NextSteps: fmt.Sprintf(`
				- Determine if the pod was OOM killed using 'kubectl describe pod %[1]s' (look for 'OOMKilled: true') and, if so, consider increasing the memory limit in the relevant 'Deployment.yaml'.
				- Check the logs before the container restarted to see if there are 'panic:' messages or similar using 'kubectl logs -p %[1]s'.
			`, containerName),
		}
	}
)

type KubernetesMonitoringOptions struct {
	// PodsAvailable transforms the default observable used to construct the pods available panel.
	PodsAvailable ObservableOption
}

// NewProvisioningIndicatorsGroup creates a group containing panels displaying
// provisioning indication metrics - long and short term usage for both CPU and
// memory usage - for the given container.
func NewKubernetesMonitoringGroup(containerName string, owner monitoring.ObservableOwner, options *KubernetesMonitoringOptions) monitoring.Group {
	if options == nil {
		options = &KubernetesMonitoringOptions{}
	}

	return monitoring.Group{
		Title:  TitleKubernetesMonitoring,
		Hidden: true,
		Rows: []monitoring.Row{
			{
				options.PodsAvailable.safeApply(KubernetesPodsAvailable(containerName, owner)).Observable(),
			},
		},
	}
}
