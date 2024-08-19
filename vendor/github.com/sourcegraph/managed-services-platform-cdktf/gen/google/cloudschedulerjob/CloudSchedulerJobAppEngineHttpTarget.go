package cloudschedulerjob


type CloudSchedulerJobAppEngineHttpTarget struct {
	// The relative URI.
	//
	// The relative URL must begin with "/" and must be a valid HTTP relative URL.
	// It can contain a path, query string arguments, and \# fragments.
	// If the relative URL is empty, then the root path "/" will be used.
	// No spaces are allowed, and the maximum length allowed is 2083 characters
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#relative_uri CloudSchedulerJob#relative_uri}
	RelativeUri *string `field:"required" json:"relativeUri" yaml:"relativeUri"`
	// app_engine_routing block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#app_engine_routing CloudSchedulerJob#app_engine_routing}
	AppEngineRouting *CloudSchedulerJobAppEngineHttpTargetAppEngineRouting `field:"optional" json:"appEngineRouting" yaml:"appEngineRouting"`
	// HTTP request body.
	//
	// A request body is allowed only if the HTTP method is POST or PUT.
	// It will result in invalid argument error to set a body on a job with an incompatible HttpMethod.
	//
	// A base64-encoded string.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#body CloudSchedulerJob#body}
	Body *string `field:"optional" json:"body" yaml:"body"`
	// HTTP request headers. This map contains the header field names and values. Headers can be set when the job is created.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#headers CloudSchedulerJob#headers}
	Headers *map[string]*string `field:"optional" json:"headers" yaml:"headers"`
	// Which HTTP method to use for the request.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#http_method CloudSchedulerJob#http_method}
	HttpMethod *string `field:"optional" json:"httpMethod" yaml:"httpMethod"`
}

