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
