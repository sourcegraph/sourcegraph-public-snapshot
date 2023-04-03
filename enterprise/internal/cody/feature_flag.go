package cody

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

func IsCodyExperimentalFeatureFlagEnabled(ctx context.Context) bool {
	return featureflag.FromContext(ctx).GetBoolOr("cody-experimental", false)
}
