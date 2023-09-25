package telemetry

// eventFeature defines the feature associated with an event. Values should
// be in camelCase, e.g. 'myFeature'
//
// This is a private type, requiring the values to be declared in-package
// and preventing strings from being cast to this type.
type eventFeature string

// All event names in Sourcegraph's Go services.
const (
	FeatureExample eventFeature = "exampleFeature"

	FeatureSignIn  eventFeature = "signIn"
	FeatureSignOut eventFeature = "signOut"
	FeatureSignUp  eventFeature = "signUp"
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
