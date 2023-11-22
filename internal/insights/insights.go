package insights

import (
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
)

func IsEnabled() bool {
	if envvar.SourcegraphDotComMode() {
		return false
	}
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_INSIGHTS")); v {
		// Code insights can always be disabled. This can be a helpful escape hatch if e.g. there
		// are issues with (or connecting to) the codeinsights-db deployment and it is preventing
		// the Sourcegraph frontend or repo-updater from starting.
		//
		// It is also useful in dev environments if you do not wish to spend resources running Code
		// Insights.
		return false
	}
	if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) {
		// Code insights is not supported in single-container Docker demo deployments unless
		// explicity allowed, (for example by backend integration tests.)
		if v, _ := strconv.ParseBool(os.Getenv("ALLOW_SINGLE_DOCKER_CODE_INSIGHTS")); v {
			return true
		}
		return false
	}
	return true
}
