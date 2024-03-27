package loadbalancer

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudrunv2service"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeglobaladdress"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeglobalforwardingrule"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computemanagedsslcertificate"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeregionnetworkendpointgroup"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesecuritypolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslcertificate"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesslpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computetargethttpsproxy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/cloudflare/datacloudflareipranges"

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

	// CloudflareProxied is true when environments[].domain.cloudflare.proxied is true
	CloudflareProxied bool
	// Production is true if environments[].category is `internal` or `external` but not `test`
	Production bool
	// EnableLogging enables always-on logging for the backend.
	EnableLogging bool
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
			Name:    config.TargetService.Name(),
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
			Name:    config.TargetService.Name(),
			Project: pointers.Ptr(config.ProjectID),

			Protocol: pointers.Ptr("HTTP"),
			PortName: pointers.Ptr("http"),

			SecurityPolicy: func() *string {
				if config.CloudflareProxied && config.Production {
					return cloudArmorAllowOnlyCloudflareEdge(scope, id.Group("cf-policy"), config)
				}
				return nil
			}(),
			LogConfig: &computebackendservice.ComputeBackendServiceLogConfig{
				// https://cloud.google.com/load-balancing/docs/https/https-logging-monitoring#viewing_logs
				Enable:     config.EnableLogging,
				SampleRate: pointers.Float64(1), // always-sample when logging is enabled
			},

			Backend: []*computebackendservice.ComputeBackendServiceBackend{{
				Group: endpointGroup.Id(),
			}},
			// Removing the security policy is an in place operation
			// As such we need to update the backend to remove the security policy
			// before the security policy can be destroyed.
			Lifecycle: &cdktf.TerraformResourceLifecycle{
				CreateBeforeDestroy: pointers.Ptr(true),
			},
		})

	// Enable routing requests to the backend service working serving traffic
	// for load balancing
	urlMap := computeurlmap.NewComputeUrlMap(scope,
		id.TerraformID("url_map"),
		&computeurlmap.ComputeUrlMapConfig{
			Name:           config.TargetService.Name(),
			Project:        pointers.Ptr(config.ProjectID),
			DefaultService: backendService.Id(),
		})

	// Set up an HTTPS proxy to route incoming HTTPS requests to our target's
	// URL map, which handles load balancing for a service.
	httpsProxy := computetargethttpsproxy.NewComputeTargetHttpsProxy(scope,
		id.TerraformID("https-proxy"),
		&computetargethttpsproxy.ComputeTargetHttpsProxyConfig{
			Name:    pointers.Stringf("%s-https", *config.TargetService.Name()),
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
					Name:    pointers.Stringf("%s-https", *config.TargetService.Name()),
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
			Name:    config.TargetService.Name(),
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

// cloudArmorAllowOnlyCloudflareEdge defines a security policy which only allows traffic from Cloudflare edge IPs
func cloudArmorAllowOnlyCloudflareEdge(scope constructs.Construct, id resourceid.ID, config Config) *string {
	securityPolicy := computesecuritypolicy.NewComputeSecurityPolicy(scope, id.TerraformID("security-policy"), &computesecuritypolicy.ComputeSecurityPolicyConfig{
		Name:    pointers.Stringf("%s-security-policy", *config.TargetService.Name()),
		Project: pointers.Ptr(config.ProjectID),
		// Rules are evaluated from highest priority (lowest numerically) to lowest priority (highest numerically) in order.
		// It is necessary to have a default rule with priority 2147483647 and CIDR match * to handle all traffic not matched by higher priority rules.
		// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_security_policy#rule
		Rule: []computesecuritypolicy.ComputeSecurityPolicyRule{
			{
				// Default rule for all unmatched traffic (not from CF IPs)
				// Allowlist for CF IPs defined using Override below
				Description: pointers.Ptr("Deny All"),
				Action:      pointers.Ptr("deny(403)"), // deny with status 403
				Priority:    pointers.Ptr(2147483647.0),
				Match: &computesecuritypolicy.ComputeSecurityPolicyRuleMatch{
					VersionedExpr: pointers.Ptr("SRC_IPS_V1"),
					Config: &computesecuritypolicy.ComputeSecurityPolicyRuleMatchConfig{
						SrcIpRanges: pointers.Ptr(pointers.Slice([]string{"*"})),
					},
				},
			},
		},
	})

	ips := datacloudflareipranges.NewDataCloudflareIpRanges(scope, id.TerraformID("cf-ip-ranges"), &datacloudflareipranges.DataCloudflareIpRangesConfig{})

	// Hack because Sort returns *[]*string but Chunklist wants *[]interface{}
	var sortedIPs []any
	for _, val := range *cdktf.Fn_Sort(ips.Ipv4CidrBlocks()) {
		sortedIPs = append(sortedIPs, any(val))
	}

	// Add Dynamic block override to add new rules for chunks of 10 CIDRs (max per rule)
	// We need to use cdktf.Fn_Chunklist because we don't know the set of Cloudflare IP ranges ahead of time
	securityPolicy.AddOverride(pointers.Ptr("dynamic.rule"), &map[string]any{
		"for_each": cdktf.Fn_Chunklist(pointers.Ptr(sortedIPs), pointers.Ptr(10.0)),
		"content": &map[string]any{
			"description": "Allow Cloudflare edge nodes ${rule.key}",
			"action":      "allow",
			"priority":    "${4000 + rule.key}",
			"match": map[string]any{
				"versioned_expr": "SRC_IPS_V1",
				"config": map[string]any{
					"src_ip_ranges": "${rule.value}",
				},
			},
		},
	})

	return securityPolicy.Name()
}
