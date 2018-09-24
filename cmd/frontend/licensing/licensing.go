package licensing

import (
	"context"
	"time"
)

// SourcegraphLicenseInfo describes a Sourcegraph license.
type SourcegraphLicenseInfo struct {
	Plan         string
	MaxUserCount *uint
	Expiry       *time.Time
}

// GetConfiguredSourcegraphLicenseInfo is called to obtain the Sourcegraph license info when
// creating the GraphQL resolver for the GraphQL type SourcegraphLicenseInfo.
//
// Exactly 1 of its return values must be non-nil.
//
// It is overridden in non-OSS builds to return information about the actual Sourcegraph license in
// use.
var GetConfiguredSourcegraphLicenseInfo = func(ctx context.Context) (*SourcegraphLicenseInfo, error) {
	// Stub value for OSS builds (where the Sourcegraph license isn't used).
	return &SourcegraphLicenseInfo{Plan: "OSS"}, nil
}
