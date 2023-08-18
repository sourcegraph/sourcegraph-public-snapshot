package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

type LicenseResolver struct{}

var _ graphqlbackend.LicenseResolver = LicenseResolver{}

func (LicenseResolver) EnterpriseLicenseHasFeature(ctx context.Context, args *graphqlbackend.EnterpriseLicenseHasFeatureArgs) (bool, error) {
	if err := licensing.Check(licensing.BasicFeature(args.Feature)); err != nil {
		if licensing.IsFeatureNotActivated(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (LicenseResolver) LicenseInfo(ctx context.Context, args *graphqlbackend.LicenseInfoArgs) (*graphqlbackend.LicenseInfoResolver, error) {
	var (
		info *licensing.Info
		err  error
	)
	if args.LicenseKey != nil {
		info, _, err = licensing.GetLicenseInfoFromKey(*args.LicenseKey)
	} else {
		info, err = licensing.GetConfiguredProductLicenseInfo()
	}
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.LicenseInfoResolver{Info: info}, nil
}
