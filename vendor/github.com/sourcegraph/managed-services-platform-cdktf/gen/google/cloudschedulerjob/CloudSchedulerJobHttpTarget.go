package cloudschedulerjob


type CloudSchedulerJobHttpTarget struct {
	// The full URI path that the request will be sent to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#uri CloudSchedulerJob#uri}
	Uri *string `field:"required" json:"uri" yaml:"uri"`
	// HTTP request body.
	//
	// A request body is allowed only if the HTTP method is POST, PUT, or PATCH.
	// It is an error to set body on a job with an incompatible HttpMethod.
	//
	// A base64-encoded string.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#body CloudSchedulerJob#body}
	Body *string `field:"optional" json:"body" yaml:"body"`
	// This map contains the header field names and values.
	//
	// Repeated headers are not supported, but a header value can contain commas.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#headers CloudSchedulerJob#headers}
	Headers *map[string]*string `field:"optional" json:"headers" yaml:"headers"`
	// Which HTTP method to use for the request.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#http_method CloudSchedulerJob#http_method}
	HttpMethod *string `field:"optional" json:"httpMethod" yaml:"httpMethod"`
	// oauth_token block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#oauth_token CloudSchedulerJob#oauth_token}
	OauthToken *CloudSchedulerJobHttpTargetOauthToken `field:"optional" json:"oauthToken" yaml:"oauthToken"`
	// oidc_token block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#oidc_token CloudSchedulerJob#oidc_token}
	OidcToken *CloudSchedulerJobHttpTargetOidcToken `field:"optional" json:"oidcToken" yaml:"oidcToken"`
}

