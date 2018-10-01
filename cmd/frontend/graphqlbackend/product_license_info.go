package graphqlbackend

import (
	"context"
	"time"
)

// GetConfiguredProductLicenseInfo is called to obtain the product subscription info when creating
// the GraphQL resolver for the GraphQL type ProductLicenseInfo.
//
// Exactly 1 of its return values must be non-nil.
//
// It is overridden in non-OSS builds to return information about the actual product subscription in
// use.
var GetConfiguredProductLicenseInfo = func(ctx context.Context) (*ProductLicenseInfo, error) {
	return nil, nil // OSS builds have no license
}

// ProductLicenseInfo implements the GraphQL type ProductLicenseInfo.
type ProductLicenseInfo struct {
	PlanValue      string
	UserCountValue *uint
	ExpiresAtValue *time.Time
}

// Plan implements the GraphQL type ProductLicenseInfo.
func (r ProductLicenseInfo) Plan() string { return r.PlanValue }

// UserCount implements the GraphQL type ProductLicenseInfo.
func (r ProductLicenseInfo) UserCount(ctx context.Context) (*int32, error) {
	if r.UserCountValue == nil {
		return nil, nil
	}
	n2 := int32(*r.UserCountValue)
	return &n2, nil
}

// ExpiresAt implements the GraphQL type ProductLicenseInfo.
func (r ProductLicenseInfo) ExpiresAt() *string {
	if r.ExpiresAtValue == nil {
		return nil
	}
	s := r.ExpiresAtValue.Format(time.RFC3339)
	return &s
}
