package clouddeploydeliverypipeline


type ClouddeployDeliveryPipelineSerialPipelineStagesStrategyCanaryRuntimeConfigKubernetesServiceNetworking struct {
	// Required. Name of the Kubernetes Deployment whose traffic is managed by the specified Service.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#deployment ClouddeployDeliveryPipeline#deployment}
	Deployment *string `field:"required" json:"deployment" yaml:"deployment"`
	// Required. Name of the Kubernetes Service.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#service ClouddeployDeliveryPipeline#service}
	Service *string `field:"required" json:"service" yaml:"service"`
	// Optional.
	//
	// Whether to disable Pod overprovisioning. If Pod overprovisioning is disabled then Cloud Deploy will limit the number of total Pods used for the deployment strategy to the number of Pods the Deployment has on the cluster.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/clouddeploy_delivery_pipeline#disable_pod_overprovisioning ClouddeployDeliveryPipeline#disable_pod_overprovisioning}
	DisablePodOverprovisioning interface{} `field:"optional" json:"disablePodOverprovisioning" yaml:"disablePodOverprovisioning"`
}

