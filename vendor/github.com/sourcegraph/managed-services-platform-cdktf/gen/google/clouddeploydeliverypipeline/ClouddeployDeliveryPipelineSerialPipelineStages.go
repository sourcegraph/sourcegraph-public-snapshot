package clouddeploydeliverypipeline


type ClouddeployDeliveryPipelineSerialPipelineStages struct {
	// deploy_parameters block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#deploy_parameters ClouddeployDeliveryPipeline#deploy_parameters}
	DeployParameters interface{} `field:"optional" json:"deployParameters" yaml:"deployParameters"`
	// Skaffold profiles to use when rendering the manifest for this stage's `Target`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#profiles ClouddeployDeliveryPipeline#profiles}
	Profiles *[]*string `field:"optional" json:"profiles" yaml:"profiles"`
	// strategy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#strategy ClouddeployDeliveryPipeline#strategy}
	Strategy *ClouddeployDeliveryPipelineSerialPipelineStagesStrategy `field:"optional" json:"strategy" yaml:"strategy"`
	// The target_id to which this stage points.
	//
	// This field refers exclusively to the last segment of a target name. For example, this field would just be `my-target` (rather than `projects/project/locations/location/targets/my-target`). The location of the `Target` is inferred to be the same as the location of the `DeliveryPipeline` that contains this `Stage`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#target_id ClouddeployDeliveryPipeline#target_id}
	TargetId *string `field:"optional" json:"targetId" yaml:"targetId"`
}

