package cloudschedulerjob


type CloudSchedulerJobAppEngineHttpTargetAppEngineRouting struct {
	// App instance. By default, the job is sent to an instance which is available when the job is attempted.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#instance CloudSchedulerJob#instance}
	Instance *string `field:"optional" json:"instance" yaml:"instance"`
	// App service.
	//
	// By default, the job is sent to the service which is the default service when the job is attempted.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#service CloudSchedulerJob#service}
	Service *string `field:"optional" json:"service" yaml:"service"`
	// App version.
	//
	// By default, the job is sent to the version which is the default version when the job is attempted.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#version CloudSchedulerJob#version}
	Version *string `field:"optional" json:"version" yaml:"version"`
}

