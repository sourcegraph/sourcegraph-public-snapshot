package shared

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Golang monitoring overviews

var (
	GoGoroutines sharedObservable = func(containerName string, owner monitoring.ObservableOwner) monitoring.Observable {
		return monitoring.Observable{
			Name:              "go_goroutines",
			Description:       "maximum active goroutines",
			Query:             fmt.Sprintf(`max by(instance) (go_goroutines{job=~".*%s"})`, containerName),
			DataMayNotExist:   true,
			Warning:           monitoring.Alert().GreaterOrEqual(10000).For(10 * time.Minute),
			PanelOptions:      monitoring.PanelOptions().LegendFormat("{{name}}"),
			Owner:             owner,
			PossibleSolutions: "none",
		}
	}

	GoGcDuration sharedObservable = func(containerName string, owner monitoring.ObservableOwner) monitoring.Observable {
		return monitoring.Observable{
			Name:              "go_gc_duration_seconds",
			Description:       "maximum go garbage collection duration",
			Query:             fmt.Sprintf(`max by(instance) (go_gc_duration_seconds{job=~".*%s"})`, containerName),
			DataMayNotExist:   true,
			Warning:           monitoring.Alert().GreaterOrEqual(2),
			PanelOptions:      monitoring.PanelOptions().LegendFormat("{{name}}").Unit(monitoring.Seconds),
			Owner:             owner,
			PossibleSolutions: "none",
		}
	}
)
