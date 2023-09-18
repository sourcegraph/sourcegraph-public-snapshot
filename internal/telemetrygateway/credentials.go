package telemetrygateway

import (
	"context"

	"google.golang.org/grpc/credentials"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

type perRPCCredentials struct {
	conf     conftypes.SiteConfigQuerier
	insecure bool
}

var _ credentials.PerRPCCredentials = (*perRPCCredentials)(nil)

func (p *perRPCCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	licenseKey := p.conf.SiteConfig().LicenseKey
	if licenseKey != "" {
		keyPartial := licenseKey
		if len(licenseKey) > 60 {
			keyPartial = licenseKey[len(licenseKey)-60:]
		}
		return map[string]string{
			"Authorization": "LicenseKeyPartial " + keyPartial,
		}, nil
	}
	return map[string]string{}, nil
}

func (p *perRPCCredentials) RequireTransportSecurity() bool {
	return !p.insecure
}
