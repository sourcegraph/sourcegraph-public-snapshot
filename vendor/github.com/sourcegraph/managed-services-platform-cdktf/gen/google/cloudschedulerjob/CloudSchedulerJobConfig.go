package cloudschedulerjob

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type CloudSchedulerJobConfig struct {
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
	// The name of the job.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#name CloudSchedulerJob#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// app_engine_http_target block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#app_engine_http_target CloudSchedulerJob#app_engine_http_target}
	AppEngineHttpTarget *CloudSchedulerJobAppEngineHttpTarget `field:"optional" json:"appEngineHttpTarget" yaml:"appEngineHttpTarget"`
	// The deadline for job attempts.
	//
	// If the request handler does not respond by this deadline then the request is
	// cancelled and the attempt is marked as a DEADLINE_EXCEEDED failure. The failed attempt can be viewed in
	// execution logs. Cloud Scheduler will retry the job according to the RetryConfig.
	// The allowed duration for this deadline is:
	// For HTTP targets, between 15 seconds and 30 minutes.
	// For App Engine HTTP targets, between 15 seconds and 24 hours.
	// **Note**: For PubSub targets, this field is ignored - setting it will introduce an unresolvable diff.
	// A duration in seconds with up to nine fractional digits, terminated by 's'. Example: "3.5s"
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#attempt_deadline CloudSchedulerJob#attempt_deadline}
	AttemptDeadline *string `field:"optional" json:"attemptDeadline" yaml:"attemptDeadline"`
	// A human-readable description for the job. This string must not contain more than 500 characters.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#description CloudSchedulerJob#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// http_target block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#http_target CloudSchedulerJob#http_target}
	HttpTarget *CloudSchedulerJobHttpTarget `field:"optional" json:"httpTarget" yaml:"httpTarget"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#id CloudSchedulerJob#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Sets the job to a paused state. Jobs default to being enabled when this property is not set.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#paused CloudSchedulerJob#paused}
	Paused interface{} `field:"optional" json:"paused" yaml:"paused"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#project CloudSchedulerJob#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// pubsub_target block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#pubsub_target CloudSchedulerJob#pubsub_target}
	PubsubTarget *CloudSchedulerJobPubsubTarget `field:"optional" json:"pubsubTarget" yaml:"pubsubTarget"`
	// Region where the scheduler job resides. If it is not provided, Terraform will use the provider default.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#region CloudSchedulerJob#region}
	Region *string `field:"optional" json:"region" yaml:"region"`
	// retry_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#retry_config CloudSchedulerJob#retry_config}
	RetryConfig *CloudSchedulerJobRetryConfig `field:"optional" json:"retryConfig" yaml:"retryConfig"`
	// Describes the schedule on which the job will be executed.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#schedule CloudSchedulerJob#schedule}
	Schedule *string `field:"optional" json:"schedule" yaml:"schedule"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#timeouts CloudSchedulerJob#timeouts}
	Timeouts *CloudSchedulerJobTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// Specifies the time zone to be used in interpreting schedule.
	//
	// The value of this field must be a time zone name from the tz database.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#time_zone CloudSchedulerJob#time_zone}
	TimeZone *string `field:"optional" json:"timeZone" yaml:"timeZone"`
}

