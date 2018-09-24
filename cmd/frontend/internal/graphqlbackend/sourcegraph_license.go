package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/licensing"
)

// currentSourcegraphLicenseInfo returns the resolver for the GraphQL type SourcegraphLicenseInfo.
func currentSourcegraphLicenseInfo(ctx context.Context) (*sourcegraphLicenseInfoResolver, error) {
	info, err := licensing.GetConfiguredSourcegraphLicenseInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &sourcegraphLicenseInfoResolver{info: *info}, nil
}

// sourcegraphLicenseInfoResolver implements the GraphQL type SourcegraphLicenseInfo.
type sourcegraphLicenseInfoResolver struct {
	// LicenseInfo describes the Sourcegraph license, if any. If there is no Sourcegraph license,
	// fallback values are used.
	info licensing.SourcegraphLicenseInfo
}

// Plan implements the GraphQL type SourcegraphLicenseInfo.
func (r sourcegraphLicenseInfoResolver) Plan() string { return r.info.Plan }

// UserCount implements the GraphQL type SourcegraphLicenseInfo.
func (sourcegraphLicenseInfoResolver) UserCount(ctx context.Context) (int32, error) {
	count, err := db.Users.Count(ctx, nil)
	return int32(count), err
}

// MaxUserCount implements the GraphQL type SourcegraphLicenseInfo.
func (r sourcegraphLicenseInfoResolver) MaxUserCount(ctx context.Context) (*int32, error) {
	if r.info.MaxUserCount == nil {
		return nil, nil
	}
	n2 := int32(*r.info.MaxUserCount)
	return &n2, nil
}

// ExpiresAt implements the GraphQL type SourcegraphLicenseInfo.
func (r sourcegraphLicenseInfoResolver) ExpiresAt() *string {
	if r.info.Expiry == nil {
		return nil
	}
	s := r.info.Expiry.Format(time.RFC3339)
	return &s
}
