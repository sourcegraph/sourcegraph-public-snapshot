package datasentryorganizationintegration

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type DataSentryOrganizationIntegrationConfig struct {
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
	// The name of the integration.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/data-sources/organization_integration#name DataSentryOrganizationIntegration#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The slug of the organization.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/data-sources/organization_integration#organization DataSentryOrganizationIntegration#organization}
	Organization *string `field:"required" json:"organization" yaml:"organization"`
	// Specific integration provider to filter by such as `slack`. See [the list of supported providers](https://docs.sentry.io/product/integrations/).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/data-sources/organization_integration#provider_key DataSentryOrganizationIntegration#provider_key}
	ProviderKey *string `field:"required" json:"providerKey" yaml:"providerKey"`
}

