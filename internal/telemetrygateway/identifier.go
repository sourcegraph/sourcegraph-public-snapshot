package telemetrygateway

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newIdentifier(ctx context.Context, c conftypes.SiteConfigQuerier, g database.GlobalStateStore) (*telemetrygatewayv1.Identifier, error) {
	globalState, err := g.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get instance ID")
	}

	// licensed instance
	if lk := c.SiteConfig().LicenseKey; lk != "" {
		return &telemetrygatewayv1.Identifier{
			Identifier: &telemetrygatewayv1.Identifier_LicensedInstance{
				LicensedInstance: &telemetrygatewayv1.Identifier_LicensedInstanceIdentifier{
					LicenseKey:  lk,
					InstanceId:  globalState.SiteID,
					ExternalUrl: c.SiteConfig().ExternalURL,
				},
			},
		}, nil
	}

	// unlicensed instance - no license key, so instanceID must be valid
	if globalState.SiteID != "" {
		return &telemetrygatewayv1.Identifier{
			Identifier: &telemetrygatewayv1.Identifier_UnlicensedInstance{
				UnlicensedInstance: &telemetrygatewayv1.Identifier_UnlicensedInstanceIdentifier{
					InstanceId:  globalState.SiteID,
					ExternalUrl: c.SiteConfig().ExternalURL,
				},
			},
		}, nil
	}

	return nil, errors.New("cannot infer an identifer for this instance")
}
