package cloudrunv2service


type CloudRunV2ServiceTemplateVolumesGcs struct {
	// GCS Bucket name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#bucket CloudRunV2Service#bucket}
	Bucket *string `field:"required" json:"bucket" yaml:"bucket"`
	// If true, mount the GCS bucket as read-only.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#read_only CloudRunV2Service#read_only}
	ReadOnly interface{} `field:"optional" json:"readOnly" yaml:"readOnly"`
}

