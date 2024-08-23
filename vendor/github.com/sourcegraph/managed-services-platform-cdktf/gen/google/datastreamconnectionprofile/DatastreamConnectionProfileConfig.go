package datastreamconnectionprofile

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type DatastreamConnectionProfileConfig struct {
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
	// The connection profile identifier.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#connection_profile_id DatastreamConnectionProfile#connection_profile_id}
	ConnectionProfileId *string `field:"required" json:"connectionProfileId" yaml:"connectionProfileId"`
	// Display name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#display_name DatastreamConnectionProfile#display_name}
	DisplayName *string `field:"required" json:"displayName" yaml:"displayName"`
	// The name of the location this connection profile is located in.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#location DatastreamConnectionProfile#location}
	Location *string `field:"required" json:"location" yaml:"location"`
	// bigquery_profile block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#bigquery_profile DatastreamConnectionProfile#bigquery_profile}
	BigqueryProfile *DatastreamConnectionProfileBigqueryProfile `field:"optional" json:"bigqueryProfile" yaml:"bigqueryProfile"`
	// forward_ssh_connectivity block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#forward_ssh_connectivity DatastreamConnectionProfile#forward_ssh_connectivity}
	ForwardSshConnectivity *DatastreamConnectionProfileForwardSshConnectivity `field:"optional" json:"forwardSshConnectivity" yaml:"forwardSshConnectivity"`
	// gcs_profile block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#gcs_profile DatastreamConnectionProfile#gcs_profile}
	GcsProfile *DatastreamConnectionProfileGcsProfile `field:"optional" json:"gcsProfile" yaml:"gcsProfile"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#id DatastreamConnectionProfile#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Labels.
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#labels DatastreamConnectionProfile#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// mysql_profile block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#mysql_profile DatastreamConnectionProfile#mysql_profile}
	MysqlProfile *DatastreamConnectionProfileMysqlProfile `field:"optional" json:"mysqlProfile" yaml:"mysqlProfile"`
	// oracle_profile block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#oracle_profile DatastreamConnectionProfile#oracle_profile}
	OracleProfile *DatastreamConnectionProfileOracleProfile `field:"optional" json:"oracleProfile" yaml:"oracleProfile"`
	// postgresql_profile block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#postgresql_profile DatastreamConnectionProfile#postgresql_profile}
	PostgresqlProfile *DatastreamConnectionProfilePostgresqlProfile `field:"optional" json:"postgresqlProfile" yaml:"postgresqlProfile"`
	// private_connectivity block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#private_connectivity DatastreamConnectionProfile#private_connectivity}
	PrivateConnectivity *DatastreamConnectionProfilePrivateConnectivity `field:"optional" json:"privateConnectivity" yaml:"privateConnectivity"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#project DatastreamConnectionProfile#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/datastream_connection_profile#timeouts DatastreamConnectionProfile#timeouts}
	Timeouts *DatastreamConnectionProfileTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}

