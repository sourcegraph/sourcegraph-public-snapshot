package envvar

import (
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var deploymentOnPrem, _ = strconv.ParseBool(env.Get("DEPLOYMENT_ON_PREM", "false", "true if the frontend is running in a customer's datacenter instead of in Sourcegraph's cloud."))

// DeploymentOnPrem returns true when the frontend is running in a customer's datacenter
// instead of in Sourcegraph's cloud.
func DeploymentOnPrem() bool {
	return deploymentOnPrem
}

var debugMode, _ = strconv.ParseBool(env.Get("DEBUG", "false", "debug mode"))

// DebugMode is true if and only if the application is running in debug mode. In
// this mode, the application should display more verbose and informative errors
// in the UI. It should also show all features (as possible). Debug should NEVER
// be true in production.
func DebugMode() bool { return debugMode }
