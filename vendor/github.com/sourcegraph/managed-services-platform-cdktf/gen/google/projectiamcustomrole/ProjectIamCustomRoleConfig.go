package projectiamcustomrole

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type ProjectIamCustomRoleConfig struct {
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
	// The names of the permissions this role grants when bound in an IAM policy.
	//
	// At least one permission must be specified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/project_iam_custom_role#permissions ProjectIamCustomRole#permissions}
	Permissions *[]*string `field:"required" json:"permissions" yaml:"permissions"`
	// The camel case role id to use for this role. Cannot contain - characters.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/project_iam_custom_role#role_id ProjectIamCustomRole#role_id}
	RoleId *string `field:"required" json:"roleId" yaml:"roleId"`
	// A human-readable title for the role.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/project_iam_custom_role#title ProjectIamCustomRole#title}
	Title *string `field:"required" json:"title" yaml:"title"`
	// A human-readable description for the role.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/project_iam_custom_role#description ProjectIamCustomRole#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/project_iam_custom_role#id ProjectIamCustomRole#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The project that the service account will be created in. Defaults to the provider project configuration.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/project_iam_custom_role#project ProjectIamCustomRole#project}
	Project *string `field:"optional" json:"project" yaml:"project"`
	// The current launch stage of the role. Defaults to GA.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/project_iam_custom_role#stage ProjectIamCustomRole#stage}
	Stage *string `field:"optional" json:"stage" yaml:"stage"`
}

