package cloudrunv2job


type CloudRunV2JobTemplateTemplateVolumesSecretItems struct {
	// The relative path of the secret in the container.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#path CloudRunV2Job#path}
	Path *string `field:"required" json:"path" yaml:"path"`
	// The Cloud Secret Manager secret version.
	//
	// Can be 'latest' for the latest value or an integer for a specific version
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#version CloudRunV2Job#version}
	Version *string `field:"required" json:"version" yaml:"version"`
	// Integer octal mode bits to use on this file, must be a value between 01 and 0777 (octal).
	//
	// If 0 or not set, the Volume's default mode will be used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#mode CloudRunV2Job#mode}
	Mode *float64 `field:"optional" json:"mode" yaml:"mode"`
}

