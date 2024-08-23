package cloudrunv2job


type CloudRunV2JobTemplateTemplateContainersEnv struct {
	// Name of the environment variable. Must be a C_IDENTIFIER, and mnay not exceed 32768 characters.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#name CloudRunV2Job#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any route environment variables.
	//
	// If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to "", and the maximum length is 32768 bytes
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#value CloudRunV2Job#value}
	Value *string `field:"optional" json:"value" yaml:"value"`
	// value_source block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#value_source CloudRunV2Job#value_source}
	ValueSource *CloudRunV2JobTemplateTemplateContainersEnvValueSource `field:"optional" json:"valueSource" yaml:"valueSource"`
}

