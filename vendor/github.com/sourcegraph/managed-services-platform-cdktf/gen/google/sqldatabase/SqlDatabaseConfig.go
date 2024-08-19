package sqldatabase

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type SqlDatabaseConfig struct {
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
	// The name of the Cloud SQL instance. This does not include the project ID.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database#instance SqlDatabase#instance}
	Instance *string `field:"required" json:"instance" yaml:"instance"`
	// The name of the database in the Cloud SQL instance. This does not include the project ID or instance name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database#name SqlDatabase#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The charset value.
	//
	// See MySQL's
	// [Supported Character Sets and Collations](https://dev.mysql.com/doc/refman/5.7/en/charset-charsets.html)
	// and Postgres' [Character Set Support](https://www.postgresql.org/docs/9.6/static/multibyte.html)
	// for more details and supported values. Postgres databases only support
	// a value of 'UTF8' at creation time.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database#charset SqlDatabase#charset}
	Charset *string `field:"optional" json:"charset" yaml:"charset"`
	// The collation value.
	//
	// See MySQL's
	// [Supported Character Sets and Collations](https://dev.mysql.com/doc/refman/5.7/en/charset-charsets.html)
	// and Postgres' [Collation Support](https://www.postgresql.org/docs/9.6/static/collation.html)
	// for more details and supported values. Postgres databases only support
	// a value of 'en_US.UTF8' at creation time.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database#collation SqlDatabase#collation}
	Collation *string `field:"optional" json:"collation" yaml:"collation"`
	// The deletion policy for the database.
	//
	// Setting ABANDON allows the resource
	// to be abandoned rather than deleted. This is useful for Postgres, where databases cannot be
	// deleted from the API if there are users other than cloudsqlsuperuser with access. Possible
	// values are: "ABANDON", "DELETE". Defaults to "DELETE".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database#deletion_policy SqlDatabase#deletion_policy}
	DeletionPolicy *string `field:"optional" json:"deletionPolicy" yaml:"deletionPolicy"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database#id SqlDatabase#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database#project SqlDatabase#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/sql_database#timeouts SqlDatabase#timeouts}
	Timeouts *SqlDatabaseTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}

