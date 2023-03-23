package cody

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func IsCodyExperimentalFeatureFlagEnabled(ctx context.Context, db database.DB) (bool, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return false, nil
	}

	userFlags, err := db.FeatureFlags().GetUserFlags(ctx, a.UID)
	if err != nil {
		return false, err
	}

	return userFlags["cody-experimental"], nil
}
