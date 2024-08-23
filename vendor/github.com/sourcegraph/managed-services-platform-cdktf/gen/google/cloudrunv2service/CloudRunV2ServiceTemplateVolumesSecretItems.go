package cloudrunv2service


type CloudRunV2ServiceTemplateVolumesSecretItems struct {
	// The relative path of the secret in the container.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#path CloudRunV2Service#path}
	Path *string `field:"required" json:"path" yaml:"path"`
	// Integer octal mode bits to use on this file, must be a value between 01 and 0777 (octal).
	//
	// If 0 or not set, the Volume's default mode will be used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#mode CloudRunV2Service#mode}
	Mode *float64 `field:"optional" json:"mode" yaml:"mode"`
	// The Cloud Secret Manager secret version.
	//
	// Can be 'latest' for the latest value or an integer for a specific version
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#version CloudRunV2Service#version}
	Version *string `field:"optional" json:"version" yaml:"version"`
}

