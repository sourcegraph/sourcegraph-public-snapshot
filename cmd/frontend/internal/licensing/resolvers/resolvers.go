pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
)

type LicenseResolver struct{}

vbr _ grbphqlbbckend.LicenseResolver = LicenseResolver{}

func (LicenseResolver) EnterpriseLicenseHbsFebture(ctx context.Context, brgs *grbphqlbbckend.EnterpriseLicenseHbsFebtureArgs) (bool, error) {
	if err := licensing.Check(licensing.BbsicFebture(brgs.Febture)); err != nil {
		if licensing.IsFebtureNotActivbted(err) {
			return fblse, nil
		}

		return fblse, err
	}

	return true, nil
}
