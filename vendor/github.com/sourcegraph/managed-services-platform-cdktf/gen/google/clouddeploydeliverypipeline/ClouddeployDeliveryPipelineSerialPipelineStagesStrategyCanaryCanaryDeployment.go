package clouddeploydeliverypipeline


type ClouddeployDeliveryPipelineSerialPipelineStagesStrategyCanaryCanaryDeployment struct {
	// Required.
	//
	// The percentage based deployments that will occur as a part of a `Rollout`. List is expected in ascending order and each integer n is 0 <= n < 100.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#percentages ClouddeployDeliveryPipeline#percentages}
	Percentages *[]*float64 `field:"required" json:"percentages" yaml:"percentages"`
	// postdeploy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#postdeploy ClouddeployDeliveryPipeline#postdeploy}
	Postdeploy *ClouddeployDeliveryPipelineSerialPipelineStagesStrategyCanaryCanaryDeploymentPostdeploy `field:"optional" json:"postdeploy" yaml:"postdeploy"`
	// predeploy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#predeploy ClouddeployDeliveryPipeline#predeploy}
	Predeploy *ClouddeployDeliveryPipelineSerialPipelineStagesStrategyCanaryCanaryDeploymentPredeploy `field:"optional" json:"predeploy" yaml:"predeploy"`
	// Whether to run verify tests after each percentage deployment.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#verify ClouddeployDeliveryPipeline#verify}
	Verify interface{} `field:"optional" json:"verify" yaml:"verify"`
}

