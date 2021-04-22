package graphqlbackend

import "context"

type LicenseResolver interface {
	EnterpriseLicenseHasFeature(ctx context.Context, args *EnterpriseLicenseHasFeatureArgs) (bool, error)
}

type EnterpriseLicenseHasFeatureArgs struct {
	Feature string
}
