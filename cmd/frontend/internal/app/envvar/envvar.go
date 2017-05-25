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
