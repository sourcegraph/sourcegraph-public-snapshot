package sqluser

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type SqlUserConfig struct {
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
	// The name of the Cloud SQL instance. Changing this forces a new resource to be created.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#instance SqlUser#instance}
	Instance *string `field:"required" json:"instance" yaml:"instance"`
	// The name of the user. Changing this forces a new resource to be created.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#name SqlUser#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The deletion policy for the user.
	//
	// Setting ABANDON allows the resource
	// to be abandoned rather than deleted. This is useful for Postgres, where users cannot be deleted from the API if they
	// have been granted SQL roles. Possible values are: "ABANDON".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#deletion_policy SqlUser#deletion_policy}
	DeletionPolicy *string `field:"optional" json:"deletionPolicy" yaml:"deletionPolicy"`
	// The host the user can connect from.
	//
	// This is only supported for MySQL instances. Don't set this field for PostgreSQL instances. Can be an IP address. Changing this forces a new resource to be created.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#host SqlUser#host}
	Host *string `field:"optional" json:"host" yaml:"host"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#id SqlUser#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The password for the user.
	//
	// Can be updated. For Postgres instances this is a Required field, unless type is set to
	//             either CLOUD_IAM_USER or CLOUD_IAM_SERVICE_ACCOUNT.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#password SqlUser#password}
	Password *string `field:"optional" json:"password" yaml:"password"`
	// password_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#password_policy SqlUser#password_policy}
	PasswordPolicy *SqlUserPasswordPolicy `field:"optional" json:"passwordPolicy" yaml:"passwordPolicy"`
	// The ID of the project in which the resource belongs.
	//
	// If it is not provided, the provider project is used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#project SqlUser#project}
	Project *string `field:"optional" json:"project" yaml:"project"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#timeouts SqlUser#timeouts}
	Timeouts *SqlUserTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// The user type.
	//
	// It determines the method to authenticate the user during login.
	//             The default is the database's built-in user type. Flags include "BUILT_IN", "CLOUD_IAM_USER", "CLOUD_IAM_GROUP" or "CLOUD_IAM_SERVICE_ACCOUNT".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_user#type SqlUser#type}
	Type *string `field:"optional" json:"type" yaml:"type"`
}

