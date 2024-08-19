package grant

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type GrantConfig struct {
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
	// The database to grant privileges on for this role.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#database Grant#database}
	Database *string `field:"required" json:"database" yaml:"database"`
	// The PostgreSQL object type to grant the privileges on (one of: database, function, procedure, routine, schema, sequence, table, foreign_data_wrapper, foreign_server, column).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#object_type Grant#object_type}
	ObjectType *string `field:"required" json:"objectType" yaml:"objectType"`
	// The list of privileges to grant.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#privileges Grant#privileges}
	Privileges *[]*string `field:"required" json:"privileges" yaml:"privileges"`
	// The name of the role to grant privileges on.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#role Grant#role}
	Role *string `field:"required" json:"role" yaml:"role"`
	// The specific columns to grant privileges on for this role.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#columns Grant#columns}
	Columns *[]*string `field:"optional" json:"columns" yaml:"columns"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#id Grant#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The specific objects to grant privileges on for this role (empty means all objects of the requested type).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#objects Grant#objects}
	Objects *[]*string `field:"optional" json:"objects" yaml:"objects"`
	// The database schema to grant privileges on for this role.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#schema Grant#schema}
	Schema *string `field:"optional" json:"schema" yaml:"schema"`
	// Permit the grant recipient to grant it to others.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/grant#with_grant_option Grant#with_grant_option}
	WithGrantOption interface{} `field:"optional" json:"withGrantOption" yaml:"withGrantOption"`
}

