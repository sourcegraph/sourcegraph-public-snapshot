package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

// GetConfiguredSourcegraphLicenseInfo is called to obtain the Sourcegraph license info when
// creating the GraphQL resolver for the GraphQL type SourcegraphLicenseInfo.
//
// Exactly 1 of its return values must be non-nil.
//
// It is overridden in non-OSS builds to return information about the actual Sourcegraph license in
// use.
var GetConfiguredSourcegraphLicenseInfo = func(ctx context.Context) (*SourcegraphLicenseInfo, error) {
	// Stub value for OSS builds (where the Sourcegraph license isn't used).
	return &SourcegraphLicenseInfo{PlanValue: "OSS"}, nil
}

// SourcegraphLicenseInfo implements the GraphQL type SourcegraphLicenseInfo.
type SourcegraphLicenseInfo struct {
	PlanValue         string
	MaxUserCountValue *uint
	ExpiresAtValue    *time.Time
}

// Enterprise is the name of the Enterprise plan.
const Enterprise = "Enterprise"

// Plan implements the GraphQL type SourcegraphLicenseInfo.
func (r SourcegraphLicenseInfo) Plan() string { return r.PlanValue }

// UserCount implements the GraphQL type SourcegraphLicenseInfo.
func (SourcegraphLicenseInfo) UserCount(ctx context.Context) (int32, error) {
	count, err := db.Users.Count(ctx, nil)
	return int32(count), err
}

// MaxUserCount implements the GraphQL type SourcegraphLicenseInfo.
func (r SourcegraphLicenseInfo) MaxUserCount(ctx context.Context) (*int32, error) {
	if r.MaxUserCountValue == nil {
		return nil, nil
	}
	n2 := int32(*r.MaxUserCountValue)
	return &n2, nil
}

// ExpiresAt implements the GraphQL type SourcegraphLicenseInfo.
func (r SourcegraphLicenseInfo) ExpiresAt() *string {
	if r.ExpiresAtValue == nil {
		return nil
	}
	s := r.ExpiresAtValue.Format(time.RFC3339)
	return &s
}

// IsExpired reports whether the license has expired.
func (r SourcegraphLicenseInfo) IsExpired() bool {
	return r.ExpiresAtValue != nil && r.ExpiresAtValue.Before(time.Now())
}
