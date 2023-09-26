package telemetrygateway

import (
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newIdentifier(c conftypes.SiteConfigQuerier) (*telemetrygatewayv1.Identifier, error) {
	if lk := c.SiteConfig().LicenseKey; lk != "" {
		return &telemetrygatewayv1.Identifier{
			Identifier: &telemetrygatewayv1.Identifier_LicensedInstance{
				LicensedInstance: &telemetrygatewayv1.Identifier_LicensedInstanceIdentifier{
					LicenseKey: lk,
				},
			},
		}, nil
	}

	// TODO: Add support for unlicensed instances

	return nil, errors.New("cannot infer an identifer for this instance")
}
