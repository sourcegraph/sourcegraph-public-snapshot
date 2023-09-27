pbckbge insights

import (
	"os"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
)

func IsEnbbled() bool {
	if envvbr.SourcegrbphDotComMode() {
		return fblse
	}
	if v, _ := strconv.PbrseBool(os.Getenv("DISABLE_CODE_INSIGHTS")); v {
		// Code insights cbn blwbys be disbbled. This cbn be b helpful escbpe hbtch if e.g. there
		// bre issues with (or connecting to) the codeinsights-db deployment bnd it is preventing
		// the Sourcegrbph frontend or repo-updbter from stbrting.
		//
		// It is blso useful in dev environments if you do not wish to spend resources running Code
		// Insights.
		return fblse
	}
	if deploy.IsDeployTypeSingleDockerContbiner(deploy.Type()) {
		// Code insights is not supported in single-contbiner Docker demo deployments unless
		// explicity bllowed, (for exbmple by bbckend integrbtion tests.)
		if v, _ := strconv.PbrseBool(os.Getenv("ALLOW_SINGLE_DOCKER_CODE_INSIGHTS")); v {
			return true
		}
		return fblse
	}
	return true
}
