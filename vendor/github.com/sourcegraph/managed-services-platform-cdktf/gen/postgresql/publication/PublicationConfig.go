package publication

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type PublicationConfig struct {
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
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#name Publication#name}.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Sets the tables list to publish to ALL tables.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#all_tables Publication#all_tables}
	AllTables interface{} `field:"optional" json:"allTables" yaml:"allTables"`
	// Sets the database to add the publication for.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#database Publication#database}
	Database *string `field:"optional" json:"database" yaml:"database"`
	// When true, will also drop all the objects that depend on the publication, and in turn all objects that depend on those objects.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#drop_cascade Publication#drop_cascade}
	DropCascade interface{} `field:"optional" json:"dropCascade" yaml:"dropCascade"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#id Publication#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Sets the owner of the publication.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#owner Publication#owner}
	Owner *string `field:"optional" json:"owner" yaml:"owner"`
	// Sets which DML operations will be published.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#publish_param Publication#publish_param}
	PublishParam *[]*string `field:"optional" json:"publishParam" yaml:"publishParam"`
	// Sets whether changes in a partitioned table using the identity and schema of the partitioned table.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#publish_via_partition_root_param Publication#publish_via_partition_root_param}
	PublishViaPartitionRootParam interface{} `field:"optional" json:"publishViaPartitionRootParam" yaml:"publishViaPartitionRootParam"`
	// Sets the tables list to publish.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/publication#tables Publication#tables}
	Tables *[]*string `field:"optional" json:"tables" yaml:"tables"`
}

