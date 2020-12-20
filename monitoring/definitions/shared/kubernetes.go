package shared

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Kubernetes monitoring overviews

var (
	KubernetesPodsAvailable sharedObservable = func(containerName string, owner monitoring.ObservableOwner) monitoring.Observable {
		return monitoring.Observable{
			Name:              "pods_available_percentage",
			Description:       "percentage pods available",
			Query:             fmt.Sprintf(`sum by(app) (up{app=~".*%[1]s"}) / count by (app) (up{app=~".*%[1]s"}) * 100`, containerName),
			Critical:          monitoring.Alert().LessOrEqual(90).For(10 * time.Minute),
			DataMayNotExist:   true,
			PanelOptions:      monitoring.PanelOptions().LegendFormat("{{name}}").Unit(monitoring.Percentage).Max(100).Min(0),
			Owner:             owner,
			PossibleSolutions: "none",
		}
	}
)
