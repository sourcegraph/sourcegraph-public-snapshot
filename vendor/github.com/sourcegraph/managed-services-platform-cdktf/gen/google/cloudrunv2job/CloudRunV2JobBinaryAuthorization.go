package cloudrunv2job


type CloudRunV2JobBinaryAuthorization struct {
	// If present, indicates to use Breakglass using this justification.
	//
	// If useDefault is False, then it must be empty. For more information on breakglass, see https://cloud.google.com/binary-authorization/docs/using-breakglass
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#breakglass_justification CloudRunV2Job#breakglass_justification}
	BreakglassJustification *string `field:"optional" json:"breakglassJustification" yaml:"breakglassJustification"`
	// If True, indicates to use the default project's binary authorization policy. If False, binary authorization will be disabled.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#use_default CloudRunV2Job#use_default}
	UseDefault interface{} `field:"optional" json:"useDefault" yaml:"useDefault"`
}

