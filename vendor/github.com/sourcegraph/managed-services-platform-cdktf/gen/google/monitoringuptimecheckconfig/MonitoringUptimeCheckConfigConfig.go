package monitoringuptimecheckconfig

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type MonitoringUptimeCheckConfigConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// A human-friendly name for the uptime check configuration.
	//
	// The display name should be unique within a Stackdriver Workspace in order to make it easier to identify; however, uniqueness is not enforced.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#display_name MonitoringUptimeCheckConfig#display_name}
	DisplayName *string `field:"required" json:"displayName" yaml:"displayName"`
	// The maximum amount of time to wait for the request to complete (must be between 1 and 60 seconds).
	//
	// [See the accepted formats]( https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.Duration)
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#timeout MonitoringUptimeCheckConfig#timeout}
	Timeout *string `field:"required" json:"timeout" yaml:"timeout"`
	// The checker type to use for the check.
	//
	// If the monitored resource type is 'servicedirectory_service', 'checker_type' must be set to 'VPC_CHECKERS'. Possible values: ["STATIC_IP_CHECKERS", "VPC_CHECKERS"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#checker_type MonitoringUptimeCheckConfig#checker_type}
	CheckerType *string `field:"optional" json:"checkerType" yaml:"checkerType"`
	// content_matchers block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#content_matchers MonitoringUptimeCheckConfig#content_matchers}
	ContentMatchers interface{} `field:"optional" json:"contentMatchers" yaml:"contentMatchers"`
	// http_check block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#http_check MonitoringUptimeCheckConfig#http_check}
	HttpCheck *MonitoringUptimeCheckConfigHttpCheck `field:"optional" json:"httpCheck" yaml:"httpCheck"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#id MonitoringUptimeCheckConfig#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// monitored_resource block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#monitored_resource MonitoringUptimeCheckConfig#monitored_resource}
	MonitoredResource *MonitoringUptimeCheckConfigMonitoredResource `field:"optional" json:"monitoredResource" yaml:"monitoredResource"`
	// How often, in seconds, the uptime check is performed.
	//
	// Currently, the only supported values are 60s (1 minute), 300s (5 minutes), 600s (10 minutes), and 900s (15 minutes). Optional, defaults to 300s.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#period MonitoringUptimeCheckConfig#period}
	Period *string `field:"optional" json:"period" yaml:"period"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#project MonitoringUptimeCheckConfig#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// resource_group block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#resource_group MonitoringUptimeCheckConfig#resource_group}
	ResourceGroup *MonitoringUptimeCheckConfigResourceGroup `field:"optional" json:"resourceGroup" yaml:"resourceGroup"`
	// The list of regions from which the check will be run.
	//
	// Some regions contain one location, and others contain more than one. If this field is specified, enough regions to include a minimum of 3 locations must be provided, or an error message is returned. Not specifying this field will result in uptime checks running from all regions.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#selected_regions MonitoringUptimeCheckConfig#selected_regions}
	SelectedRegions *[]*string `field:"optional" json:"selectedRegions" yaml:"selectedRegions"`
	// synthetic_monitor block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#synthetic_monitor MonitoringUptimeCheckConfig#synthetic_monitor}
	SyntheticMonitor *MonitoringUptimeCheckConfigSyntheticMonitor `field:"optional" json:"syntheticMonitor" yaml:"syntheticMonitor"`
	// tcp_check block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#tcp_check MonitoringUptimeCheckConfig#tcp_check}
	TcpCheck *MonitoringUptimeCheckConfigTcpCheck `field:"optional" json:"tcpCheck" yaml:"tcpCheck"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#timeouts MonitoringUptimeCheckConfig#timeouts}
	Timeouts *MonitoringUptimeCheckConfigTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// User-supplied key/value data to be used for organizing and identifying the 'UptimeCheckConfig' objects.
	//
	// The field can contain up to 64 entries. Each key and value is limited to 63 Unicode characters or 128 bytes, whichever is smaller. Labels and values can contain only lowercase letters, numerals, underscores, and dashes. Keys must begin with a letter.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/monitoring_uptime_check_config#user_labels MonitoringUptimeCheckConfig#user_labels}
	UserLabels *map[string]*string `field:"optional" json:"userLabels" yaml:"userLabels"`
}

