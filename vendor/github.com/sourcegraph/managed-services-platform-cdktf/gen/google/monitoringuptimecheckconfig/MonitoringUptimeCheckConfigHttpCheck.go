package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigHttpCheck struct {
	// accepted_response_status_codes block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#accepted_response_status_codes MonitoringUptimeCheckConfig#accepted_response_status_codes}
	AcceptedResponseStatusCodes interface{} `field:"optional" json:"acceptedResponseStatusCodes" yaml:"acceptedResponseStatusCodes"`
	// auth_info block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#auth_info MonitoringUptimeCheckConfig#auth_info}
	AuthInfo *MonitoringUptimeCheckConfigHttpCheckAuthInfo `field:"optional" json:"authInfo" yaml:"authInfo"`
	// The request body associated with the HTTP POST request.
	//
	// If 'content_type' is 'URL_ENCODED', the body passed in must be URL-encoded. Users can provide a 'Content-Length' header via the 'headers' field or the API will do so. If the 'request_method' is 'GET' and 'body' is not empty, the API will return an error. The maximum byte size is 1 megabyte. Note - As with all bytes fields JSON representations are base64 encoded. e.g. 'foo=bar' in URL-encoded form is 'foo%3Dbar' and in base64 encoding is 'Zm9vJTI1M0RiYXI='.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#body MonitoringUptimeCheckConfig#body}
	Body *string `field:"optional" json:"body" yaml:"body"`
	// The content type to use for the check. Possible values: ["TYPE_UNSPECIFIED", "URL_ENCODED", "USER_PROVIDED"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#content_type MonitoringUptimeCheckConfig#content_type}
	ContentType *string `field:"optional" json:"contentType" yaml:"contentType"`
	// A user provided content type header to use for the check.
	//
	// The invalid configurations outlined in the 'content_type' field apply to custom_content_type', as well as the following 1. 'content_type' is 'URL_ENCODED' and 'custom_content_type' is set. 2. 'content_type' is 'USER_PROVIDED' and 'custom_content_type' is not set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#custom_content_type MonitoringUptimeCheckConfig#custom_content_type}
	CustomContentType *string `field:"optional" json:"customContentType" yaml:"customContentType"`
	// The list of headers to send as part of the uptime check request.
	//
	// If two headers have the same key and different values, they should be entered as a single header, with the value being a comma-separated list of all the desired values as described in [RFC 2616 (page 31)](https://www.w3.org/Protocols/rfc2616/rfc2616.txt). Entering two separate headers with the same key in a Create call will cause the first to be overwritten by the second. The maximum number of headers allowed is 100.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#headers MonitoringUptimeCheckConfig#headers}
	Headers *map[string]*string `field:"optional" json:"headers" yaml:"headers"`
	// Boolean specifying whether to encrypt the header information.
	//
	// Encryption should be specified for any headers related to authentication that you do not wish to be seen when retrieving the configuration. The server will be responsible for encrypting the headers. On Get/List calls, if 'mask_headers' is set to 'true' then the headers will be obscured with '******'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#mask_headers MonitoringUptimeCheckConfig#mask_headers}
	MaskHeaders interface{} `field:"optional" json:"maskHeaders" yaml:"maskHeaders"`
	// The path to the page to run the check against.
	//
	// Will be combined with the host (specified within the MonitoredResource) and port to construct the full URL. If the provided path does not begin with '/', a '/' will be prepended automatically. Optional (defaults to '/').
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#path MonitoringUptimeCheckConfig#path}
	Path *string `field:"optional" json:"path" yaml:"path"`
	// ping_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#ping_config MonitoringUptimeCheckConfig#ping_config}
	PingConfig *MonitoringUptimeCheckConfigHttpCheckPingConfig `field:"optional" json:"pingConfig" yaml:"pingConfig"`
	// The port to the page to run the check against.
	//
	// Will be combined with 'host' (specified within the ['monitored_resource'](#nested_monitored_resource)) and path to construct the full URL. Optional (defaults to 80 without SSL, or 443 with SSL).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#port MonitoringUptimeCheckConfig#port}
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// The HTTP request method to use for the check.
	//
	// If set to 'METHOD_UNSPECIFIED' then 'request_method' defaults to 'GET'. Default value: "GET" Possible values: ["METHOD_UNSPECIFIED", "GET", "POST"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#request_method MonitoringUptimeCheckConfig#request_method}
	RequestMethod *string `field:"optional" json:"requestMethod" yaml:"requestMethod"`
	// service_agent_authentication block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#service_agent_authentication MonitoringUptimeCheckConfig#service_agent_authentication}
	ServiceAgentAuthentication *MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthentication `field:"optional" json:"serviceAgentAuthentication" yaml:"serviceAgentAuthentication"`
	// If true, use HTTPS instead of HTTP to run the check.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#use_ssl MonitoringUptimeCheckConfig#use_ssl}
	UseSsl interface{} `field:"optional" json:"useSsl" yaml:"useSsl"`
	// Boolean specifying whether to include SSL certificate validation as a part of the Uptime check.
	//
	// Only applies to checks where 'monitored_resource' is set to 'uptime_url'. If 'use_ssl' is 'false', setting 'validate_ssl' to 'true' has no effect.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#validate_ssl MonitoringUptimeCheckConfig#validate_ssl}
	ValidateSsl interface{} `field:"optional" json:"validateSsl" yaml:"validateSsl"`
}

