package key

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type KeyConfig struct {
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
	// The name of the key.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/key#name Key#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The slug of the organization the key should be created for.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/key#organization Key#organization}
	Organization *string `field:"required" json:"organization" yaml:"organization"`
	// The slug of the project the key should be created for.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/key#project Key#project}
	Project *string `field:"required" json:"project" yaml:"project"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/key#id Key#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Number of events that can be reported within the rate limit window.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/key#rate_limit_count Key#rate_limit_count}
	RateLimitCount *float64 `field:"optional" json:"rateLimitCount" yaml:"rateLimitCount"`
	// Length of time that will be considered when checking the rate limit.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/key#rate_limit_window Key#rate_limit_window}
	RateLimitWindow *float64 `field:"optional" json:"rateLimitWindow" yaml:"rateLimitWindow"`
}

