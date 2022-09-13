// Package shared contains shared declarations between dashboards. In general, you should NOT be making
// changes to this package: we use a declarative style for monitoring intentionally, so you should err
// on the side of repeating yourself and NOT writing shared or programatically generated monitoring.
//
// When editing this package or introducing any shared declarations, you should abide strictly by the
// following rules:
//
//  1. Do NOT declare a shared definition unless 5+ dashboards will use it. Sharing dashboard
//     declarations means the codebase becomes more complex and non-declarative which we want to avoid
//     so repeat yourself instead if it applies to less than 5 dashboards.
//
//  2. ONLY declare shared Observables. Introducing shared Rows or Groups prevents individual dashboard
//     maintainers from holistically considering both the layout of dashboards as well as the
//     metrics and alerts defined within them -- which we do not want.
//
//  3. Use the sharedObservable type and do NOT parameterize more than just the container name. It may
//     be tempting to pass an alerting threshold as an argument, or parameterize whether a critical
//     alert is defined -- but this makes reasoning about alerts at a high level much more difficult.
//     If you have a need for this, it is a strong signal you should NOT be using the shared definition
//     anymore and should instead copy it and apply your modifications.
//
// Learn more about monitoring in https://handbook.sourcegraph.com/engineering/observability/monitoring_pillars
package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// sharedObservable defines the type all shared observable variables should have in this package.
type sharedObservable func(containerName string, owner monitoring.ObservableOwner) Observable

// Observable is a variant of normal Observables that offer convenience functions for
// customizing shared observables.
type Observable monitoring.Observable

// Observable is a convenience adapter that casts this SharedObservable as an normal Observable.
func (o Observable) Observable() monitoring.Observable { return monitoring.Observable(o) }

// WithWarning overrides this Observable's warning-level alert with the given alert.
func (o Observable) WithWarning(a *monitoring.ObservableAlertDefinition) Observable {
	o.Warning = a
	if a != nil {
		o.NoAlert = false
	}
	return o
}

// WithCritical overrides this Observable's critical-level alert with the given alert.
func (o Observable) WithCritical(a *monitoring.ObservableAlertDefinition) Observable {
	o.Critical = a
	if a != nil {
		o.NoAlert = false
	}
	return o
}

// WithNoAlerts disables alerting on this Observable and sets the given interpretation instead.
func (o Observable) WithNoAlerts(interpretation string) Observable {
	o.Warning = nil
	o.Critical = nil
	o.NoAlert = true
	o.NextSteps = ""
	o.Interpretation = interpretation
	return o
}

// ObservableOption is a function that transforms an observable.
type ObservableOption func(observable Observable) Observable

func (f ObservableOption) safeApply(observable Observable) Observable {
	if f == nil {
		return observable
	}

	return f(observable)
}

// and creates a chained ObservableOption that first invokes the receiver,
// and the the argument on the result of invoking the receiver.
func (f ObservableOption) and(m ObservableOption) ObservableOption {
	return func(observable Observable) Observable {
		return m.safeApply(f.safeApply(observable))
	}
}

// WarningOption creates an ObservableOption that overrides this Observable's
// warning-level alert with the given alert.
func WarningOption(a *monitoring.ObservableAlertDefinition, possibleSolution string) ObservableOption {
	return func(observable Observable) Observable {
		observable = observable.WithWarning(a)
		observable.NextSteps = possibleSolution
		return observable
	}
}

// CriticalOption creates an ObservableOption that overrides this Observable's
// critical-level alert with the given alert.
func CriticalOption(a *monitoring.ObservableAlertDefinition, possibleSolution string) ObservableOption {
	return func(observable Observable) Observable {
		observable = observable.WithCritical(a)
		observable.NextSteps = possibleSolution
		return observable
	}
}

// NoAlertsOption creates an ObservableOption that disables alerting on this
// Observable and sets the given interpretation instead.
func NoAlertsOption(interpretation string) ObservableOption {
	return func(observable Observable) Observable {
		return observable.WithNoAlerts(interpretation)
	}
}

// CadvisorContainerNameMatcher generates Prometheus matchers that capture metrics that match the
// given container name while excluding some irrelevant series.
func CadvisorContainerNameMatcher(containerName string) string {
	// Name must start with the container name exactly.
	//
	// In docker-compose:
	// - `name` is just the container name
	// - suffix could be replica in docker-compose ('-0', '-1')
	//
	// In Kubernetes:
	// - a `metric_relabel_configs` generates a `name` with the format `CONTAINERNAME-PODNAME`,
	//   because cAdvisor does not consistently generate a name in all container runtimes.
	//   See https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/deploy-sourcegraph%24+target_label:+name&patternType=literal
	// - because of above, suffix could be pod name in Kubernetes
	return fmt.Sprintf(`name=~"^%s.*"`, containerName)
}

// CadvisorPodNameMatcher generates Prometheus matchers that capture metrics that match the
// given pod name.
func CadvisorPodNameMatcher(podName string) string {
	// The regex handles values with arbitrary prefixes and suffixes around the core pod
	// name.
	return fmt.Sprintf("container_label_io_kubernetes_pod_name=~`.*%s.*`", podName)
}

func titlecase(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
