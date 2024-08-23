package clouddeploydeliverypipeline


type ClouddeployDeliveryPipelineSerialPipelineStagesStrategyStandard struct {
	// postdeploy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#postdeploy ClouddeployDeliveryPipeline#postdeploy}
	Postdeploy *ClouddeployDeliveryPipelineSerialPipelineStagesStrategyStandardPostdeploy `field:"optional" json:"postdeploy" yaml:"postdeploy"`
	// predeploy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#predeploy ClouddeployDeliveryPipeline#predeploy}
	Predeploy *ClouddeployDeliveryPipelineSerialPipelineStagesStrategyStandardPredeploy `field:"optional" json:"predeploy" yaml:"predeploy"`
	// Whether to verify a deployment.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#verify ClouddeployDeliveryPipeline#verify}
	Verify interface{} `field:"optional" json:"verify" yaml:"verify"`
}

