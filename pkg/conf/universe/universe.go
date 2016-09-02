package universe

import (
	"context"
	"os"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/betautil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sqs/pbtypes"
)

// Enabled tells if universe should be used because repo is a universe enabled
// repository (see isUniverseRepo below) OR because the user in the context is
// part of the universe beta. It also performs feature flag checking, such that
// this function is all you need to do for checking universe.
func Enabled(ctx context.Context, repo string) bool {
	if !feature.Features.Universe {
		return false
	}
	if EnabledExcludingBeta(repo) {
		return true
	}
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return false
	}
	info, err := cl.Auth.Identify(ctx, &pbtypes.Void{})
	if err != nil {
		return false
	}
	if info.UID == 0 {
		return false
	}
	user, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{UID: info.UID})
	if err != nil {
		return false
	}
	return user.InBeta(betautil.Universe)
}

// EnabledExcludingBeta is just like Enabled except it excludes users who are
// in the beta program. It should only be used for operations which would
// otherwise affect users not in the universe beta (e.g. data altering
// operations). Effectively it just checks if the given repo is a universe
// repo.
func EnabledExcludingBeta(repo string) bool {
	return repoChecker(feature.Features.Universe, os.Getenv("SG_UNIVERSE_REPO"), repo)
}

func repoChecker(on bool, enabled, repo string) bool {
	if !on {
		return false
	}
	if enabled == "all" {
		return true
	}
	if enabled == "" {
		// Go testing repos.
		enabled = "github.com/sourcegraph/sourcegraph"
		enabled += ",github.com/slimsag/mux"
		enabled += ",github.com/slimsag/context"
		enabled += ",github.com/slimsag/docker"
		enabled += ",github.com/slimsag/kubernetes"
		enabled += ",github.com/slimsag/prometheus"

		// Java testing repos.
		enabled += ",github.com/slimsag/RxJava"
		enabled += ",github.com/slimsag/guava"
		enabled += ",github.com/slimsag/joda-time"

		// JavaScript testing repos.
		enabled += ",github.com/sgtest/javascript-nodejs-sample-0"
		enabled += ",github.com/sgtest/javascript-nodejs-xrefs-0"
		enabled += ",github.com/sgtest/minimal_nodejs_stdlib"
		enabled += ",github.com/sgtest/js-misc"
		enabled += ",github.com/sgtest/js-misc"
		enabled += ",github.com/sgtest/javascript-es6-tests"
	}
	for _, e := range strings.Split(enabled, ",") {
		if repo == e {
			return true
		}
	}
	return false
}
