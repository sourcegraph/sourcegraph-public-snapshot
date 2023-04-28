package cody

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

// IsCodyEnabled determines if cody is enabled for the actor in the given context.
// If it is an unauthenticated request, cody is disabled.
// If CodyRestrictUsersFeatureFlag is set, the cody-experimental featureflag
// will determine access.
// Otherwise, all authenticated users are granted access.
func IsCodyEnabled(ctx context.Context) bool {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return false
	}
	if conf.Get().ExperimentalFeatures.CodyRestrictUsersFeatureFlag {
		return featureflag.FromContext(ctx).GetBoolOr("cody-experimental", false)
	}
	return true
}
