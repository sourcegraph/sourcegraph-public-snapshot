package privategoogleaccess

import (
	"fmt"
	"net/netip"

	"github.com/aws/constructs-go/constructs/v10"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computenetwork"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/dnsmanagedzone"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/dnsrecordset"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	ProjectID string
	Network   computenetwork.ComputeNetwork
	Spec      spec.EnvironmentPrivateGoogleAccessSpec
}

type Output struct{}

type privateGoogleAccessDomain struct {
	id     string // dash-delimited, all-lowercase
	domain string // e.g. 'run.app.' - see https://cloud.google.com/vpc/docs/configure-private-google-access#domain-options
}

// See private.googleapis.com entry in:
// https://cloud.google.com/vpc/docs/configure-private-google-access#config-options
const privateGoogleIPCIDR = "199.36.153.8/30"

// New sets up resources required to route relevant traffic to Private Google
// Access.
//
// This should only be created once, hence why it does not have accept
// a resourceid.ID
func New(scope constructs.Construct, config Config) (*Output, error) {
	id := resourceid.New("privategoogleaccess") // top-level because this resource is a singleton

	var domains []privateGoogleAccessDomain
	if pointers.DerefZero(config.Spec.CloudRunApps) {
		domains = append(domains, privateGoogleAccessDomain{
			id:     "cloudrun",
			domain: "run.app",
		})
	}

	// See the following guides for how all this works:
	// - https://cloud.google.com/vpc/docs/configure-private-google-access#config-domain
	// - https://cloud.google.com/run/docs/securing/ingress#settings
	pgaIPs, err := getPrivateGoogleIPs()
	if err != nil {
		return nil, errors.Wrap(err, "getPrivateGoogleIPs")
	}

	// TTL to use for various records
	recordTTL := 86400

	// Only add *.run.app for now, but structure to enable other domains:
	// - https://cloud.google.com/vpc/docs/configure-private-google-access#domain-options
	// - https://github.com/sourcegraph/managed-services/issues/1093
	for _, pga := range domains {
		id := id.Group(pga.id)
		dnsName := fmt.Sprintf("%s.", pga.domain)

		zone := dnsmanagedzone.NewDnsManagedZone(scope, id.TerraformID("dns_zone"), &dnsmanagedzone.DnsManagedZoneConfig{
			Project:     pointers.Ptr(config.ProjectID),
			Name:        pointers.Ptr(pga.id),
			DnsName:     pointers.Stringf(dnsName),
			Description: pointers.Stringf("Private DNS zone for routing %s URLs to Private Google Access", pga.id),
			Visibility:  pointers.Ptr("private"),
			PrivateVisibilityConfig: &dnsmanagedzone.DnsManagedZonePrivateVisibilityConfig{
				Networks: &dnsmanagedzone.DnsManagedZonePrivateVisibilityConfigNetworks{
					NetworkUrl: config.Network.Id(),
				},
			},
		})

		// See private.googleapis.com references in the following guides:
		// - https://cloud.google.com/vpc/docs/configure-private-google-access#config-domain
		// - https://cloud.google.com/vpc/docs/configure-private-google-access#domain-options
		_ = dnsrecordset.NewDnsRecordSet(scope, id.TerraformID("a_record"), &dnsrecordset.DnsRecordSetConfig{
			Project:     pointers.Ptr(config.ProjectID),
			ManagedZone: zone.Name(),

			Name:    pointers.Ptr(dnsName),
			Type:    pointers.Ptr("A"),
			Ttl:     pointers.Float64(recordTTL),
			Rrdatas: pointers.Ptr(pointers.Slice(pgaIPs)),
		})

		_ = dnsrecordset.NewDnsRecordSet(scope, id.TerraformID("cname"), &dnsrecordset.DnsRecordSetConfig{
			Project:     pointers.Ptr(config.ProjectID),
			ManagedZone: zone.Name(),

			Name:    pointers.Stringf("*.%s", dnsName),
			Type:    pointers.Ptr("CNAME"),
			Ttl:     pointers.Float64(recordTTL),
			Rrdatas: pointers.Ptr([]*string{pointers.Ptr(dnsName)}),
		})
	}

	return &Output{}, nil
}

func getPrivateGoogleIPs() ([]string, error) {
	p, err := netip.ParsePrefix(privateGoogleIPCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "netip.ParsePrefix")
	}

	var addresses []string
	addr := p.Masked().Addr()
	for {
		if !p.Contains(addr) {
			break
		}
		addresses = append(addresses, addr.String())
		addr = addr.Next()
	}

	return addresses, nil
}
