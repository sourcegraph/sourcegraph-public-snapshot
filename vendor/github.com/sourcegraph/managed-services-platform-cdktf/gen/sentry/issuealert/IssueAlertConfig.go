package issuealert

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type IssueAlertConfig struct {
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
	// Trigger actions when an event is captured by Sentry and `any` or `all` of the specified conditions happen.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#action_match IssueAlert#action_match}
	ActionMatch *string `field:"required" json:"actionMatch" yaml:"actionMatch"`
	// List of actions. In JSON string format.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#actions IssueAlert#actions}
	Actions *string `field:"required" json:"actions" yaml:"actions"`
	// List of conditions. In JSON string format.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#conditions IssueAlert#conditions}
	Conditions *string `field:"required" json:"conditions" yaml:"conditions"`
	// Perform actions at most once every `X` minutes for this issue.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#frequency IssueAlert#frequency}
	Frequency *float64 `field:"required" json:"frequency" yaml:"frequency"`
	// The issue alert name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#name IssueAlert#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The slug of the organization the resource belongs to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#organization IssueAlert#organization}
	Organization *string `field:"required" json:"organization" yaml:"organization"`
	// The slug of the project the resource belongs to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#project IssueAlert#project}
	Project *string `field:"required" json:"project" yaml:"project"`
	// Perform issue alert in a specific environment.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#environment IssueAlert#environment}
	Environment *string `field:"optional" json:"environment" yaml:"environment"`
	// A string determining which filters need to be true before any actions take place.
	//
	// Required when a value is provided for `filters`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#filter_match IssueAlert#filter_match}
	FilterMatch *string `field:"optional" json:"filterMatch" yaml:"filterMatch"`
	// A list of filters that determine if a rule fires after the necessary conditions have been met.
	//
	// In JSON string format.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#filters IssueAlert#filters}
	Filters *string `field:"optional" json:"filters" yaml:"filters"`
	// The ID of the team or user that owns the rule.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs/resources/issue_alert#owner IssueAlert#owner}
	Owner *string `field:"optional" json:"owner" yaml:"owner"`
}

