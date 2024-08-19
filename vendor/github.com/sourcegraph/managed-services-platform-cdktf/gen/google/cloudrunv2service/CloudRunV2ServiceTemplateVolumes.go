package cloudrunv2service


type CloudRunV2ServiceTemplateVolumes struct {
	// Volume's name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#name CloudRunV2Service#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// cloud_sql_instance block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#cloud_sql_instance CloudRunV2Service#cloud_sql_instance}
	CloudSqlInstance *CloudRunV2ServiceTemplateVolumesCloudSqlInstance `field:"optional" json:"cloudSqlInstance" yaml:"cloudSqlInstance"`
	// gcs block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#gcs CloudRunV2Service#gcs}
	Gcs *CloudRunV2ServiceTemplateVolumesGcs `field:"optional" json:"gcs" yaml:"gcs"`
	// nfs block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#nfs CloudRunV2Service#nfs}
	Nfs *CloudRunV2ServiceTemplateVolumesNfs `field:"optional" json:"nfs" yaml:"nfs"`
	// secret block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#secret CloudRunV2Service#secret}
	Secret *CloudRunV2ServiceTemplateVolumesSecret `field:"optional" json:"secret" yaml:"secret"`
}

