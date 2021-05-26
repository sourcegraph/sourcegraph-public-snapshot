package featureflag

import "time"

type FeatureFlag struct {
	Name string

	// A feature flag is one of the following types.
	// Exactly one of the following will be set.
	Bool    *FeatureFlagBool
	Rollout *FeatureFlagRollout

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type FeatureFlagBool struct {
	Value bool
}

type FeatureFlagRollout struct {
	// Rollout is an integer between 0 and 10000, representing the percent of
	// users for which this feature flag will evaluate to 'true' in increments
	// of 0.01%
	Rollout int
}

type FeatureFlagOverride struct {
	UserID   *int32
	OrgID    *int32
	FlagName string
	Value    bool
}
