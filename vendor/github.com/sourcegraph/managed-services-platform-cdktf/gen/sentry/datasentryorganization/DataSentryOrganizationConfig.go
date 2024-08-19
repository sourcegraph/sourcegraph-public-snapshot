package datasentryorganization

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type DataSentryOrganizationConfig struct {
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
	// The unique URL slug for this organization.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/data-sources/organization#slug DataSentryOrganization#slug}
	Slug *string `field:"required" json:"slug" yaml:"slug"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/data-sources/organization#id DataSentryOrganization#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
}

