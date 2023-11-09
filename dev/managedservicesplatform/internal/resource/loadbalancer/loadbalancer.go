package loadbalancer

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeglobaladdress"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeglobalforwardingrule"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computemanagedsslcertificate"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeregionnetworkendpointgroup"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslcertificate"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computetargethttpsproxy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
	ExternalAddress computeglobaladdress.ComputeGlobalAddress
}

type Config struct {
	ProjectID string
	Region    string

	// TargetService should be the Cloud Run Service to point to.
	TargetService cloudrunv2service.CloudRunV2Service

	// SSLCertificate must be either computesslcertificate.ComputeSslCertificate
	// or computemanagedsslcertificate.ComputeManagedSslCertificate. It's used
	// by the loadbalancer's HTTPS proxy.
	SSLCertificate SSLCertificate
}

type SSLCertificate interface {
	Id() *string
}

// New instantiates a set of resources for a load-balancer backend that routes
// requests to a Cloud Run service:
//
//	ExternalAddress (Output)
//	  -> ForwardingRule
//	    -> HTTPSProxy
//	      -> URLMap
//	        -> BackendService
//	          -> NetworkEndpointGroup
//	            -> CloudRun (TargetService)
//
// Typically some other frontend will then be placed in front of URLMap, e.g.
// resource/cloudflare.
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	switch config.SSLCertificate.(type) {
	case computesslcertificate.ComputeSslCertificate, computemanagedsslcertificate.ComputeManagedSslCertificate:
		// ok
	default:
		return nil, errors.Newf("SSLCertificate must be either ComputeSslCertificate or ComputeManagedSslCertificate, got %T",
			config.SSLCertificate)
	}

	// Endpoint group represents the Cloud Run service.
	endpointGroup := computeregionnetworkendpointgroup.NewComputeRegionNetworkEndpointGroup(scope,
		id.TerraformID("endpoint_group"),
		&computeregionnetworkendpointgroup.ComputeRegionNetworkEndpointGroupConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: pointers.Ptr(config.ProjectID),
			Region:  pointers.Ptr(config.Region),

			NetworkEndpointType: pointers.Ptr("SERVERLESS"),
			CloudRun: &computeregionnetworkendpointgroup.ComputeRegionNetworkEndpointGroupCloudRun{
				Service: config.TargetService.Name(),
			},
		})

	// Set up a group of virtual machines that will serve traffic for load balancing
	backendService := computebackendservice.NewComputeBackendService(scope,
		id.TerraformID("backend_service"),
		&computebackendservice.ComputeBackendServiceConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: pointers.Ptr(config.ProjectID),

			Protocol: pointers.Ptr("HTTP"),
			PortName: pointers.Ptr("http"),

			// TODO: Parameterize with cloudflaresecuritypolicy as needed
			SecurityPolicy: nil,

			Backend: []*computebackendservice.ComputeBackendServiceBackend{{
				Group: endpointGroup.Id(),
			}},
		})

	// Enable routing requests to the backend service working serving traffic
	// for load balancing
	urlMap := computeurlmap.NewComputeUrlMap(scope,
		id.TerraformID("url_map"),
		&computeurlmap.ComputeUrlMapConfig{
			Name:           pointers.Ptr(id.DisplayName()),
			Project:        pointers.Ptr(config.ProjectID),
			DefaultService: backendService.Id(),
		})

	// Set up an HTTPS proxy to route incoming HTTPS requests to our target's
	// URL map, which handles load balancing for a service.
	httpsProxy := computetargethttpsproxy.NewComputeTargetHttpsProxy(scope,
		id.TerraformID("https-proxy"),
		&computetargethttpsproxy.ComputeTargetHttpsProxyConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: pointers.Ptr(config.ProjectID),
			// target the URL map
			UrlMap: urlMap.Id(),
			// via our SSL configuration
			SslCertificates: pointers.Ptr([]*string{
				config.SSLCertificate.Id(),
			}),
			SslPolicy: computesslpolicy.NewComputeSslPolicy(
				scope,
				id.TerraformID("ssl-policy"),
				&computesslpolicy.ComputeSslPolicyConfig{
					Name:    pointers.Ptr(id.DisplayName()),
					Project: pointers.Ptr(config.ProjectID),

					Profile:       pointers.Ptr("MODERN"),
					MinTlsVersion: pointers.Ptr("TLS_1_2"),
				},
			).Id(),
		})

	// Set up an external address to receive traffic
	externalAddress := computeglobaladdress.NewComputeGlobalAddress(
		scope,
		id.TerraformID("external-address"),
		&computeglobaladdress.ComputeGlobalAddressConfig{
			Name:        pointers.Ptr(id.DisplayName()),
			Project:     pointers.Ptr(config.ProjectID),
			AddressType: pointers.Ptr("EXTERNAL"),
			IpVersion:   pointers.Ptr("IPV4"),
		},
	)

	// Forward traffic from the external address to the HTTPS proxy that then
	// routes request to our target
	_ = computeglobalforwardingrule.NewComputeGlobalForwardingRule(scope,
		id.TerraformID("forwarding-rule"),
		&computeglobalforwardingrule.ComputeGlobalForwardingRuleConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: pointers.Ptr(config.ProjectID),

			IpAddress: externalAddress.Address(),
			PortRange: pointers.Ptr("443"),

			Target:              httpsProxy.Id(),
			LoadBalancingScheme: pointers.Ptr("EXTERNAL"),
		})

	return &Output{
		ExternalAddress: externalAddress,
	}, nil
}
