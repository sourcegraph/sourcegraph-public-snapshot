package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthentication struct {
	// The type of authentication to use. Possible values: ["SERVICE_AGENT_AUTHENTICATION_TYPE_UNSPECIFIED", "OIDC_TOKEN"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#type MonitoringUptimeCheckConfig#type}
	Type *string `field:"optional" json:"type" yaml:"type"`
}

