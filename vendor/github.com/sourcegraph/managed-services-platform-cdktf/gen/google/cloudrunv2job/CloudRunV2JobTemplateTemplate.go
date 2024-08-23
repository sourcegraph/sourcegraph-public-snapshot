package cloudrunv2job


type CloudRunV2JobTemplateTemplate struct {
	// containers block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#containers CloudRunV2Job#containers}
	Containers interface{} `field:"optional" json:"containers" yaml:"containers"`
	// A reference to a customer managed encryption key (CMEK) to use to encrypt this container image.
	//
	// For more information, go to https://cloud.google.com/run/docs/securing/using-cmek
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#encryption_key CloudRunV2Job#encryption_key}
	EncryptionKey *string `field:"optional" json:"encryptionKey" yaml:"encryptionKey"`
	// The execution environment being used to host this Task. Possible values: ["EXECUTION_ENVIRONMENT_GEN1", "EXECUTION_ENVIRONMENT_GEN2"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#execution_environment CloudRunV2Job#execution_environment}
	ExecutionEnvironment *string `field:"optional" json:"executionEnvironment" yaml:"executionEnvironment"`
	// Number of retries allowed per Task, before marking this Task failed.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#max_retries CloudRunV2Job#max_retries}
	MaxRetries *float64 `field:"optional" json:"maxRetries" yaml:"maxRetries"`
	// Email address of the IAM service account associated with the Task of a Job.
	//
	// The service account represents the identity of the running task, and determines what permissions the task has. If not provided, the task will use the project's default service account.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#service_account CloudRunV2Job#service_account}
	ServiceAccount *string `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	// Max allowed time duration the Task may be active before the system will actively try to mark it failed and kill associated containers.
	//
	// This applies per attempt of a task, meaning each retry can run for the full timeout.
	//
	// A duration in seconds with up to nine fractional digits, ending with 's'. Example: "3.5s".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#timeout CloudRunV2Job#timeout}
	Timeout *string `field:"optional" json:"timeout" yaml:"timeout"`
	// volumes block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#volumes CloudRunV2Job#volumes}
	Volumes interface{} `field:"optional" json:"volumes" yaml:"volumes"`
	// vpc_access block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_job#vpc_access CloudRunV2Job#vpc_access}
	VpcAccess *CloudRunV2JobTemplateTemplateVpcAccess `field:"optional" json:"vpcAccess" yaml:"vpcAccess"`
}

