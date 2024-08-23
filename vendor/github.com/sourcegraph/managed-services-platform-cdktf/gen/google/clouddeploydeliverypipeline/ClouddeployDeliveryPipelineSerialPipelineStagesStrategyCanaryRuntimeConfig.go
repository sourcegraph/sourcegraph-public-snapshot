package clouddeploydeliverypipeline


type ClouddeployDeliveryPipelineSerialPipelineStagesStrategyCanaryRuntimeConfig struct {
	// cloud_run block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#cloud_run ClouddeployDeliveryPipeline#cloud_run}
	CloudRun *ClouddeployDeliveryPipelineSerialPipelineStagesStrategyCanaryRuntimeConfigCloudRun `field:"optional" json:"cloudRun" yaml:"cloudRun"`
	// kubernetes block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#kubernetes ClouddeployDeliveryPipeline#kubernetes}
	Kubernetes *ClouddeployDeliveryPipelineSerialPipelineStagesStrategyCanaryRuntimeConfigKubernetes `field:"optional" json:"kubernetes" yaml:"kubernetes"`
}

