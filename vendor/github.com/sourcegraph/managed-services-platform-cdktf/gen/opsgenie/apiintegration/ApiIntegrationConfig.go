package apiintegration

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ApiIntegrationConfig struct {
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
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#name ApiIntegration#name}.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#allow_write_access ApiIntegration#allow_write_access}.
	AllowWriteAccess interface{} `field:"optional" json:"allowWriteAccess" yaml:"allowWriteAccess"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#enabled ApiIntegration#enabled}.
	Enabled interface{} `field:"optional" json:"enabled" yaml:"enabled"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#headers ApiIntegration#headers}.
	Headers *map[string]*string `field:"optional" json:"headers" yaml:"headers"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#id ApiIntegration#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#ignore_responders_from_payload ApiIntegration#ignore_responders_from_payload}.
	IgnoreRespondersFromPayload interface{} `field:"optional" json:"ignoreRespondersFromPayload" yaml:"ignoreRespondersFromPayload"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#owner_team_id ApiIntegration#owner_team_id}.
	OwnerTeamId *string `field:"optional" json:"ownerTeamId" yaml:"ownerTeamId"`
	// responders block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#responders ApiIntegration#responders}
	Responders interface{} `field:"optional" json:"responders" yaml:"responders"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#suppress_notifications ApiIntegration#suppress_notifications}.
	SuppressNotifications interface{} `field:"optional" json:"suppressNotifications" yaml:"suppressNotifications"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#type ApiIntegration#type}.
	Type *string `field:"optional" json:"type" yaml:"type"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/resources/api_integration#webhook_url ApiIntegration#webhook_url}.
	WebhookUrl *string `field:"optional" json:"webhookUrl" yaml:"webhookUrl"`
}

