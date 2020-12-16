package graphqlbackend

import "context"

type LicenseResolver interface {
	EnterpriseLicenseHasFeature(ctx context.Context, args *EnterpriseLicenseHasFeatureArgs) (bool, error)
}

type EnterpriseLicenseHasFeatureArgs struct {
	Feature string
}

type defaultLicenseResolver struct{}

var DefaultLicenseResolver = defaultLicenseResolver{}

func (defaultLicenseResolver) EnterpriseLicenseHasFeature(ctx context.Context, args *EnterpriseLicenseHasFeatureArgs) (bool, error) {
	// By definition, no enterprise features are available in the open source
	// version of Sourcegraph.
	return false, nil
}
