package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigResourceGroup struct {
	// The group of resources being monitored. Should be the 'name' of a group.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#group_id MonitoringUptimeCheckConfig#group_id}
	GroupId *string `field:"optional" json:"groupId" yaml:"groupId"`
	// The resource type of the group members. Possible values: ["RESOURCE_TYPE_UNSPECIFIED", "INSTANCE", "AWS_ELB_LOAD_BALANCER"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#resource_type MonitoringUptimeCheckConfig#resource_type}
	ResourceType *string `field:"optional" json:"resourceType" yaml:"resourceType"`
}

