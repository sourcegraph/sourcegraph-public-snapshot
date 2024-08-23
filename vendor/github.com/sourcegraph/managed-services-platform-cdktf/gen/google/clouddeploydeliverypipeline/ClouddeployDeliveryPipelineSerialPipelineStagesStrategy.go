package clouddeploydeliverypipeline


type ClouddeployDeliveryPipelineSerialPipelineStagesStrategy struct {
	// canary block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#canary ClouddeployDeliveryPipeline#canary}
	Canary *ClouddeployDeliveryPipelineSerialPipelineStagesStrategyCanary `field:"optional" json:"canary" yaml:"canary"`
	// standard block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#standard ClouddeployDeliveryPipeline#standard}
	Standard *ClouddeployDeliveryPipelineSerialPipelineStagesStrategyStandard `field:"optional" json:"standard" yaml:"standard"`
}

