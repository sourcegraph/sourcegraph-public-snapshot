package telemetry

import "strings"

// eventFeature defines the feature associated with an event. Values should
// be in camelCase, e.g. 'myFeature'
//
// This is a private type, requiring the values to be declared in-package or
// be const strings. This roughly prevents arbitrary string values (potentially
// unsafe) from being cast to this type.
//
// 🚨 DO NOT EXPORT - this is intentionally unexported to avoid casting that
// may expose unsafe strings.
type eventFeature string

const (
	// FeatureExample is a value for testing - do not use.
	FeatureExample eventFeature = "exampleFeature"
)

// eventAction defines the action associated with an event. Values should
// be in camelCase, e.g. 'myAction'
//
// This is a private type, requiring the values to be declared in-package or
// be const strings. This roughly prevents arbitrary string values (potentially
// unsafe) from being cast to this type. The telemetry.Action() constructor is
// available as a fallback - see the relevant docstring for more details.
//
// 🚨 DO NOT EXPORT - this is intentionally unexported to avoid casting that
// may expose unsafe strings.
type eventAction string

const (
	ActionExample eventAction = "exampleAction"

	// ActionFailed, ActionSucceeded, ActionAttempted, and so on are some common
	// actions that can be used to denote the result of an event of a particular
	// eventFeature.
	ActionFailed    eventAction = "failed"
	ActionSucceeded eventAction = "succeeded"
	ActionAttempted eventAction = "attempted"
)

// SafeAction is an escape hatch for constructing eventAction from variable strings
// for known string enums. Where possible, prefer to use a constant string or a
// predefined action constant in the internal/telemetry package instead.
//
// 🚨 SECURITY: Use with care, as variable strings can accidentally contain data
// sensitive to standalone Sourcegraph instances.
func SafeAction(parts ...string) eventAction {
	return eventAction(strings.Join(parts, "."))
}
