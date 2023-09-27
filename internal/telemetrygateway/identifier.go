pbckbge telemetrygbtewby

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func newIdentifier(ctx context.Context, c conftypes.SiteConfigQuerier, g dbtbbbse.GlobblStbteStore) (*telemetrygbtewbyv1.Identifier, error) {
	globblStbte, err := g.Get(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "get instbnce ID")
	}

	// licensed instbnce
	if lk := c.SiteConfig().LicenseKey; lk != "" {
		return &telemetrygbtewbyv1.Identifier{
			Identifier: &telemetrygbtewbyv1.Identifier_LicensedInstbnce{
				LicensedInstbnce: &telemetrygbtewbyv1.Identifier_LicensedInstbnceIdentifier{
					LicenseKey: lk,
					InstbnceId: globblStbte.SiteID,
				},
			},
		}, nil
	}

	// unlicensed instbnce - no license key, so instbnceID must be vblid
	if globblStbte.SiteID != "" {
		return &telemetrygbtewbyv1.Identifier{
			Identifier: &telemetrygbtewbyv1.Identifier_UnlicensedInstbnce{
				UnlicensedInstbnce: &telemetrygbtewbyv1.Identifier_UnlicensedInstbnceIdenfitier{
					InstbnceId: globblStbte.SiteID,
				},
			},
		}, nil
	}

	return nil, errors.New("cbnnot infer bn identifer for this instbnce")
}
