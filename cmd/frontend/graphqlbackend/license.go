package graphqlbackend

import (
	"context"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

type LicenseResolver interface {
	EnterpriseLicenseHasFeature(ctx context.Context, args *EnterpriseLicenseHasFeatureArgs) (bool, error)
	LicenseInfo(ctx context.Context, args *LicenseInfoArgs) (*LicenseInfoResolver, error)
}

type EnterpriseLicenseHasFeatureArgs struct {
	Feature string
}

type LicenseInfoArgs struct {
	LicenseKey *string
}

type LicenseInfoResolver struct {
	Info *licensing.Info
}

func (r LicenseInfoResolver) Plan() string {
	return cases.Title(language.English).String(string(r.Info.Plan()))
}
func (r LicenseInfoResolver) ExpiresAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.Info.ExpiresAt}
}
func (r LicenseInfoResolver) UserCount() int32 { return int32(r.Info.UserCount) }
func (r LicenseInfoResolver) UserCountRestricted() bool {
	return !r.Info.HasTag(licensing.TrueUpUserCountTag)
}
func (r LicenseInfoResolver) Features() []LicenseFeatureResolver {
	plan := r.Info.Plan()
	allFeatures := licensing.AllFeatures[:]
	var featureResolvers []LicenseFeatureResolver
	for _, feature := range allFeatures {
		featureResolvers = append(featureResolvers, LicenseFeatureResolver{
			feature: feature,
			enabled: plan.HasFeature(feature, false),
		})
	}
	return featureResolvers
}

type LicenseFeatureResolver struct {
	feature licensing.Feature
	enabled bool
}

func (r LicenseFeatureResolver) Name() string  { return r.feature.DisplayName() }
func (r LicenseFeatureResolver) Enabled() bool { return r.enabled }
