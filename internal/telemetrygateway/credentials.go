package telemetrygateway

import (
	"context"

	"google.golang.org/grpc/credentials"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/license"
)

type perRPCCredentials struct {
	conf     conftypes.SiteConfigQuerier
	insecure bool
}

var _ credentials.PerRPCCredentials = (*perRPCCredentials)(nil)

func (p *perRPCCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	licenseKey := p.conf.SiteConfig().LicenseKey
	if licenseKey != "" {
		return map[string]string{
			"Authorization": "Bearer " + license.GenerateLicenseKeyBasedAccessToken(licenseKey),
		}, nil
	}
	return map[string]string{}, nil
}

func (p *perRPCCredentials) RequireTransportSecurity() bool {
	return !p.insecure
}
