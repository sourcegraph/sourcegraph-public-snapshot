package httpapi

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/betautil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sqs/pbtypes"
)

// useUniverse lets us opt-in universe based on both the repo or if a user is
// in the universe beta.
func useUniverse(ctx context.Context, repo string) bool {
	if !feature.Features.Universe {
		return false
	}
	if feature.IsUniverseRepo(repo) {
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
