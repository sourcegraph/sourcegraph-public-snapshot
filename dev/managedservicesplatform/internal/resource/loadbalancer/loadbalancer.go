package loadbalancer

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeglobaladdress"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeglobalforwardingrule"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeregionnetworkendpointgroup"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslcertificate"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computetargethttpsproxy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/googlesecretsmanager"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
	ExternalAddress computeglobaladdress.ComputeGlobalAddress
}

type Config struct {
	ProjectID string
	Region    string

	TargetService cloudrunv2service.CloudRunV2Service
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
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	// Endpoint group represents the Cloud Run service.
	endpointGroup := computeregionnetworkendpointgroup.NewComputeRegionNetworkEndpointGroup(scope,
		id.ResourceID("endpoint_group"),
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
		id.ResourceID("backend_service"),
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
		id.ResourceID("url_map"),
		&computeurlmap.ComputeUrlMapConfig{
			Name:           pointers.Ptr(id.DisplayName()),
			Project:        pointers.Ptr(config.ProjectID),
			DefaultService: backendService.Id(),
		})

	// Create an SSL certificate from a secret in the shared secrets project
	//
	// TODO(@bobheadxi): Provision our own certificates with
	// computesslcertificate.NewComputeSslCertificate, see sourcegraph/controller
	sslCert := computesslcertificate.NewComputeSslCertificate(scope,
		id.ResourceID("origin-cert"),
		&computesslcertificate.ComputeSslCertificateConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: pointers.Ptr(config.ProjectID),

			PrivateKey: &gsmsecret.Get(scope, id.SubID("secret-origin-private-key"), gsmsecret.DataConfig{
				Secret:    googlesecretsmanager.SecretSourcegraphWildcardKey,
				ProjectID: googlesecretsmanager.ProjectID,
			}).Value,
			Certificate: &gsmsecret.Get(scope, id.SubID("secret-origin-cert"), gsmsecret.DataConfig{
				Secret:    googlesecretsmanager.SecretSourcegraphWildcardCert,
				ProjectID: googlesecretsmanager.ProjectID,
			}).Value,

			Lifecycle: &cdktf.TerraformResourceLifecycle{
				CreateBeforeDestroy: pointers.Ptr(true),
			},
		})

	// Set up an HTTPS proxy to route incoming HTTPS requests to our target's
	// URL map, which handles load balancing for a service.
	httpsProxy := computetargethttpsproxy.NewComputeTargetHttpsProxy(scope,
		id.ResourceID("https-proxy"),
		&computetargethttpsproxy.ComputeTargetHttpsProxyConfig{
			Name:    pointers.Ptr(id.DisplayName()),
			Project: pointers.Ptr(config.ProjectID),
			// target the URL map
			UrlMap: urlMap.Id(),
			// via our SSL configuration
			SslCertificates: pointers.Ptr([]*string{
				sslCert.Id(),
			}),
			SslPolicy: computesslpolicy.NewComputeSslPolicy(
				scope,
				id.ResourceID("ssl-policy"),
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
		id.ResourceID("external-address"),
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
		id.ResourceID("forwarding-rule"),
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
	}
}
