package monitoringuptimecheckconfig


type MonitoringUptimeCheckConfigContentMatchersJsonPathMatcher struct {
	// JSONPath within the response output pointing to the expected 'ContentMatcher::content' to match against.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#json_path MonitoringUptimeCheckConfig#json_path}
	JsonPath *string `field:"required" json:"jsonPath" yaml:"jsonPath"`
	// Options to perform JSONPath content matching. Default value: "EXACT_MATCH" Possible values: ["EXACT_MATCH", "REGEX_MATCH"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#json_matcher MonitoringUptimeCheckConfig#json_matcher}
	JsonMatcher *string `field:"optional" json:"jsonMatcher" yaml:"jsonMatcher"`
}

