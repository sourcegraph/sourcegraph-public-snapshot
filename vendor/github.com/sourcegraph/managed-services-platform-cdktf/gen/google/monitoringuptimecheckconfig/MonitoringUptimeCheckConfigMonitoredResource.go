package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigMonitoredResource struct {
	// Values for all of the labels listed in the associated monitored resource descriptor.
	//
	// For example, Compute Engine VM instances use the labels 'project_id', 'instance_id', and 'zone'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#labels MonitoringUptimeCheckConfig#labels}
	Labels *map[string]*string `field:"required" json:"labels" yaml:"labels"`
	// The monitored resource type.
	//
	// This field must match the type field of a ['MonitoredResourceDescriptor'](https://cloud.google.com/monitoring/api/ref_v3/rest/v3/projects.monitoredResourceDescriptors#MonitoredResourceDescriptor) object. For example, the type of a Compute Engine VM instance is 'gce_instance'. For a list of types, see [Monitoring resource types](https://cloud.google.com/monitoring/api/resources) and [Logging resource types](https://cloud.google.com/logging/docs/api/v2/resource-list).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#type MonitoringUptimeCheckConfig#type}
	Type *string `field:"required" json:"type" yaml:"type"`
}

