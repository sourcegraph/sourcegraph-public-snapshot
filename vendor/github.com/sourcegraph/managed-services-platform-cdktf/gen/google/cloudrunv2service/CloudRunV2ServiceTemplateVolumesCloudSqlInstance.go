package cloudrunv2service


type CloudRunV2ServiceTemplateVolumesCloudSqlInstance struct {
	// The Cloud SQL instance connection names, as can be found in https://console.cloud.google.com/sql/instances. Visit https://cloud.google.com/sql/docs/mysql/connect-run for more information on how to connect Cloud SQL and Cloud Run. Format: {project}:{location}:{instance}.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#instances CloudRunV2Service#instances}
	Instances *[]*string `field:"optional" json:"instances" yaml:"instances"`
}

