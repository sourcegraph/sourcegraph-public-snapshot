package computeurlmap

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ComputeUrlMapConfig struct {
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
	// Provided by the client when the resource is created. The
	// name must be 1-63 characters long, and comply with RFC1035. Specifically, the
	// name must be 1-63 characters long and match the regular expression
	// '[a-z]([-a-z0-9]*[a-z0-9])?' which means the first character must be a lowercase
	// letter, and all following characters must be a dash, lowercase letter, or digit,
	// except the last character, which cannot be a dash.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#name ComputeUrlMap#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// default_route_action block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#default_route_action ComputeUrlMap#default_route_action}
	DefaultRouteAction *ComputeUrlMapDefaultRouteAction `field:"optional" json:"defaultRouteAction" yaml:"defaultRouteAction"`
	// The backend service or backend bucket to use when none of the given rules match.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#default_service ComputeUrlMap#default_service}
	DefaultService *string `field:"optional" json:"defaultService" yaml:"defaultService"`
	// default_url_redirect block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#default_url_redirect ComputeUrlMap#default_url_redirect}
	DefaultUrlRedirect *ComputeUrlMapDefaultUrlRedirect `field:"optional" json:"defaultUrlRedirect" yaml:"defaultUrlRedirect"`
	// An optional description of this resource. Provide this property when you create the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#description ComputeUrlMap#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// header_action block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#header_action ComputeUrlMap#header_action}
	HeaderAction *ComputeUrlMapHeaderAction `field:"optional" json:"headerAction" yaml:"headerAction"`
	// host_rule block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#host_rule ComputeUrlMap#host_rule}
	HostRule interface{} `field:"optional" json:"hostRule" yaml:"hostRule"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#id ComputeUrlMap#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// path_matcher block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#path_matcher ComputeUrlMap#path_matcher}
	PathMatcher interface{} `field:"optional" json:"pathMatcher" yaml:"pathMatcher"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#project ComputeUrlMap#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// test block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#test ComputeUrlMap#test}
	Test interface{} `field:"optional" json:"test" yaml:"test"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_url_map#timeouts ComputeUrlMap#timeouts}
	Timeouts *ComputeUrlMapTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}

