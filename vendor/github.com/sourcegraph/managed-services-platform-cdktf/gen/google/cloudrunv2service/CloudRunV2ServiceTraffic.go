package cloudrunv2service


type CloudRunV2ServiceTraffic struct {
	// Specifies percent of the traffic to this Revision. This defaults to zero if unspecified.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#percent CloudRunV2Service#percent}
	Percent *float64 `field:"optional" json:"percent" yaml:"percent"`
	// Revision to which to send this portion of traffic, if traffic allocation is by revision.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#revision CloudRunV2Service#revision}
	Revision *string `field:"optional" json:"revision" yaml:"revision"`
	// Indicates a string to be part of the URI to exclusively reference this target.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#tag CloudRunV2Service#tag}
	Tag *string `field:"optional" json:"tag" yaml:"tag"`
	// The allocation type for this traffic target. Possible values: ["TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST", "TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#type CloudRunV2Service#type}
	Type *string `field:"optional" json:"type" yaml:"type"`
}

