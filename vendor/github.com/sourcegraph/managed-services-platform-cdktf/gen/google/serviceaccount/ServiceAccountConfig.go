package serviceaccount

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ServiceAccountConfig struct {
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
	// The account id that is used to generate the service account email address and a stable unique id.
	//
	// It is unique within a project, must be 6-30 characters long, and match the regular expression [a-z]([-a-z0-9]*[a-z0-9]) to comply with RFC1035. Changing this forces a new service account to be created.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account#account_id ServiceAccount#account_id}
	AccountId *string `field:"required" json:"accountId" yaml:"accountId"`
	// If set to true, skip service account creation if a service account with the same email already exists.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account#create_ignore_already_exists ServiceAccount#create_ignore_already_exists}
	CreateIgnoreAlreadyExists interface{} `field:"optional" json:"createIgnoreAlreadyExists" yaml:"createIgnoreAlreadyExists"`
	// A text description of the service account. Must be less than or equal to 256 UTF-8 bytes.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account#description ServiceAccount#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Whether the service account is disabled. Defaults to false.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account#disabled ServiceAccount#disabled}
	Disabled interface{} `field:"optional" json:"disabled" yaml:"disabled"`
	// The display name for the service account. Can be updated without creating a new resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account#display_name ServiceAccount#display_name}
	DisplayName *string `field:"optional" json:"displayName" yaml:"displayName"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account#id ServiceAccount#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The ID of the project that the service account will be created in. Defaults to the provider project configuration.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account#project ServiceAccount#project}
	Project *string `field:"optional" json:"project" yaml:"project"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/service_account#timeouts ServiceAccount#timeouts}
	Timeouts *ServiceAccountTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}

