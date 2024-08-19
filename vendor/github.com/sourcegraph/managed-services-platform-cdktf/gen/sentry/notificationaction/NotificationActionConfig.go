package notificationaction

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type NotificationActionConfig struct {
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
	// The slug of the organization the project belongs to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/notification_action#organization NotificationAction#organization}
	Organization *string `field:"required" json:"organization" yaml:"organization"`
	// The list of project slugs that the Notification Action is created for.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/notification_action#projects NotificationAction#projects}
	Projects *[]*string `field:"required" json:"projects" yaml:"projects"`
	// The service that is used for sending the notification.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/notification_action#service_type NotificationAction#service_type}
	ServiceType *string `field:"required" json:"serviceType" yaml:"serviceType"`
	// The type of trigger that will activate this action. Valid values are `spike-protection`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/notification_action#trigger_type NotificationAction#trigger_type}
	TriggerType *string `field:"required" json:"triggerType" yaml:"triggerType"`
	// The ID of the integration that is used for sending the notification.
	//
	// Use the `sentry_organization_integration` data source to retrieve an integration. Required if `service_type` is `slack`, `pagerduty` or `opsgenie`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/notification_action#integration_id NotificationAction#integration_id}
	IntegrationId *string `field:"optional" json:"integrationId" yaml:"integrationId"`
	// The display name of the target that is used for sending the notification (e.g. Slack channel name). Required if `service_type` is `slack` or `opsgenie`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/notification_action#target_display NotificationAction#target_display}
	TargetDisplay *string `field:"optional" json:"targetDisplay" yaml:"targetDisplay"`
	// The identifier of the target that is used for sending the notification (e.g. Slack channel ID). Required if `service_type` is `slack` or `opsgenie`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/notification_action#target_identifier NotificationAction#target_identifier}
	TargetIdentifier *string `field:"optional" json:"targetIdentifier" yaml:"targetIdentifier"`
}

