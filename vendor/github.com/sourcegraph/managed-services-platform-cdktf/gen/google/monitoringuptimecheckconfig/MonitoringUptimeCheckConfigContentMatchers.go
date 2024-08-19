package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigContentMatchers struct {
	// String or regex content to match (max 1024 bytes).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#content MonitoringUptimeCheckConfig#content}
	Content *string `field:"required" json:"content" yaml:"content"`
	// json_path_matcher block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#json_path_matcher MonitoringUptimeCheckConfig#json_path_matcher}
	JsonPathMatcher *MonitoringUptimeCheckConfigContentMatchersJsonPathMatcher `field:"optional" json:"jsonPathMatcher" yaml:"jsonPathMatcher"`
	// The type of content matcher that will be applied to the server output, compared to the content string when the check is run.
	//
	// Default value: "CONTAINS_STRING" Possible values: ["CONTAINS_STRING", "NOT_CONTAINS_STRING", "MATCHES_REGEX", "NOT_MATCHES_REGEX", "MATCHES_JSON_PATH", "NOT_MATCHES_JSON_PATH"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#matcher MonitoringUptimeCheckConfig#matcher}
	Matcher *string `field:"optional" json:"matcher" yaml:"matcher"`
}

