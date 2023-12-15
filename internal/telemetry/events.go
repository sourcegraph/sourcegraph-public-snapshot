package telemetry

import "strings"

// eventFeature defines the feature associated with an event. Values should
// be in camelCase, e.g. 'myFeature'
//
// This is a private type, requiring the values to be declared in-package
// and preventing strings from being cast to this type.
type eventFeature string

// Shared event names in Sourcegraph's backend services.
const (
	FeatureExample eventFeature = "exampleFeature"

	// FeatureSourcegraphOperator collects all events related to Sourcegraph
	// Operatores.
	FeatureSourcegraphOperator eventFeature = "sourcegraphOperator"
)

// eventAction defines the action associated with an event. Values should
// be in camelCase, e.g. 'myAction'
//
// This is a private type, requiring the values to be declared in-package
// and preventing strings from being cast to this type.
type eventAction string

const (
	ActionExample eventAction = "exampleAction"

	ActionFailed    eventAction = "failed"
	ActionSucceeded eventAction = "succeeded"
	ActionAttempted eventAction = "attempted"
)

// Action is an escape hatch for constructing eventAction from variable strings.
// where possible, prefer to use a constant string or a predefined action constant
// in the internal/telemetry package instead.
//
// 🚨 SECURITY: Use with care, as variable strings can accidentally contain data
// sensitive to standalone Sourcegraph instances.
func Action(parts ...string) eventAction {
	return eventAction(strings.Join(parts, "."))
}
