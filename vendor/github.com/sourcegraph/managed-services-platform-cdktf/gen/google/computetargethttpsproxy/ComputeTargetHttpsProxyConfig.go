package computetargethttpsproxy

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ComputeTargetHttpsProxyConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// Name of the resource.
	//
	// Provided by the client when the resource is
	// created. The name must be 1-63 characters long, and comply with
	// RFC1035. Specifically, the name must be 1-63 characters long and match
	// the regular expression '[a-z]([-a-z0-9]*[a-z0-9])?' which means the
	// first character must be a lowercase letter, and all following
	// characters must be a dash, lowercase letter, or digit, except the last
	// character, which cannot be a dash.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#name ComputeTargetHttpsProxy#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// A reference to the UrlMap resource that defines the mapping from URL to the BackendService.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#url_map ComputeTargetHttpsProxy#url_map}
	UrlMap *string `field:"required" json:"urlMap" yaml:"urlMap"`
	// URLs to certificate manager certificate resources that are used to authenticate connections between users and the load balancer.
	//
	// Currently, you may specify up to 15 certificates. Certificate manager certificates do not apply when the load balancing scheme is set to INTERNAL_SELF_MANAGED.
	// sslCertificates and certificateManagerCertificates fields can not be defined together.
	// Accepted format is '//certificatemanager.googleapis.com/projects/{project}/locations/{location}/certificates/{resourceName}' or just the self_link 'projects/{project}/locations/{location}/certificates/{resourceName}'
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#certificate_manager_certificates ComputeTargetHttpsProxy#certificate_manager_certificates}
	CertificateManagerCertificates *[]*string `field:"optional" json:"certificateManagerCertificates" yaml:"certificateManagerCertificates"`
	// A reference to the CertificateMap resource uri that identifies a certificate map associated with the given target proxy.
	//
	// This field can only be set for global target proxies.
	// Accepted format is '//certificatemanager.googleapis.com/projects/{project}/locations/{location}/certificateMaps/{resourceName}'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#certificate_map ComputeTargetHttpsProxy#certificate_map}
	CertificateMap *string `field:"optional" json:"certificateMap" yaml:"certificateMap"`
	// An optional description of this resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#description ComputeTargetHttpsProxy#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Specifies how long to keep a connection open, after completing a response, while there is no matching traffic (in seconds).
	//
	// If an HTTP keepalive is
	// not specified, a default value (610 seconds) will be used. For Global
	// external HTTP(S) load balancer, the minimum allowed value is 5 seconds and
	// the maximum allowed value is 1200 seconds. For Global external HTTP(S)
	// load balancer (classic), this option is not available publicly.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#http_keep_alive_timeout_sec ComputeTargetHttpsProxy#http_keep_alive_timeout_sec}
	HttpKeepAliveTimeoutSec *float64 `field:"optional" json:"httpKeepAliveTimeoutSec" yaml:"httpKeepAliveTimeoutSec"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#id ComputeTargetHttpsProxy#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#project ComputeTargetHttpsProxy#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// This field only applies when the forwarding rule that references this target proxy has a loadBalancingScheme set to INTERNAL_SELF_MANAGED.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#proxy_bind ComputeTargetHttpsProxy#proxy_bind}
	ProxyBind interface{} `field:"optional" json:"proxyBind" yaml:"proxyBind"`
	// Specifies the QUIC override policy for this resource.
	//
	// This determines
	// whether the load balancer will attempt to negotiate QUIC with clients
	// or not. Can specify one of NONE, ENABLE, or DISABLE. If NONE is
	// specified, Google manages whether QUIC is used. Default value: "NONE" Possible values: ["NONE", "ENABLE", "DISABLE"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#quic_override ComputeTargetHttpsProxy#quic_override}
	QuicOverride *string `field:"optional" json:"quicOverride" yaml:"quicOverride"`
	// A URL referring to a networksecurity.ServerTlsPolicy resource that describes how the proxy should authenticate inbound traffic. serverTlsPolicy only applies to a global TargetHttpsProxy attached to globalForwardingRules with the loadBalancingScheme set to INTERNAL_SELF_MANAGED or EXTERNAL or EXTERNAL_MANAGED. For details which ServerTlsPolicy resources are accepted with INTERNAL_SELF_MANAGED and which with EXTERNAL, EXTERNAL_MANAGED loadBalancingScheme consult ServerTlsPolicy documentation. If left blank, communications are not encrypted.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#server_tls_policy ComputeTargetHttpsProxy#server_tls_policy}
	ServerTlsPolicy *string `field:"optional" json:"serverTlsPolicy" yaml:"serverTlsPolicy"`
	// URLs to SslCertificate resources that are used to authenticate connections between users and the load balancer.
	//
	// Currently, you may specify up to 15 SSL certificates. sslCertificates do not apply when the load balancing scheme is set to INTERNAL_SELF_MANAGED.
	// sslCertificates and certificateManagerCertificates can not be defined together.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#ssl_certificates ComputeTargetHttpsProxy#ssl_certificates}
	SslCertificates *[]*string `field:"optional" json:"sslCertificates" yaml:"sslCertificates"`
	// A reference to the SslPolicy resource that will be associated with the TargetHttpsProxy resource.
	//
	// If not set, the TargetHttpsProxy
	// resource will not have any SSL policy configured.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#ssl_policy ComputeTargetHttpsProxy#ssl_policy}
	SslPolicy *string `field:"optional" json:"sslPolicy" yaml:"sslPolicy"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy#timeouts ComputeTargetHttpsProxy#timeouts}
	Timeouts *ComputeTargetHttpsProxyTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}

