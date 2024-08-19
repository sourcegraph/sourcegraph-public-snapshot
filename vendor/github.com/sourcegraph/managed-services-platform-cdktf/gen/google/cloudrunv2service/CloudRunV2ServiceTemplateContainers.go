package cloudrunv2service


type CloudRunV2ServiceTemplateContainers struct {
	// URL of the Container image in Google Container Registry or Google Artifact Registry. More info: https://kubernetes.io/docs/concepts/containers/images.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#image CloudRunV2Service#image}
	Image *string `field:"required" json:"image" yaml:"image"`
	// Arguments to the entrypoint.
	//
	// The docker image's CMD is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#args CloudRunV2Service#args}
	Args *[]*string `field:"optional" json:"args" yaml:"args"`
	// Entrypoint array.
	//
	// Not executed within a shell. The docker image's ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME) are expanded using the container's environment. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#command CloudRunV2Service#command}
	Command *[]*string `field:"optional" json:"command" yaml:"command"`
	// Containers which should be started before this container.
	//
	// If specified the container will wait to start until all containers with the listed names are healthy.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#depends_on CloudRunV2Service#depends_on}
	DependsOn *[]*string `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// env block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#env CloudRunV2Service#env}
	Env interface{} `field:"optional" json:"env" yaml:"env"`
	// liveness_probe block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#liveness_probe CloudRunV2Service#liveness_probe}
	LivenessProbe *CloudRunV2ServiceTemplateContainersLivenessProbe `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	// Name of the container specified as a DNS_LABEL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#name CloudRunV2Service#name}
	Name *string `field:"optional" json:"name" yaml:"name"`
	// ports block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#ports CloudRunV2Service#ports}
	Ports *CloudRunV2ServiceTemplateContainersPorts `field:"optional" json:"ports" yaml:"ports"`
	// resources block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#resources CloudRunV2Service#resources}
	Resources *CloudRunV2ServiceTemplateContainersResources `field:"optional" json:"resources" yaml:"resources"`
	// startup_probe block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#startup_probe CloudRunV2Service#startup_probe}
	StartupProbe *CloudRunV2ServiceTemplateContainersStartupProbe `field:"optional" json:"startupProbe" yaml:"startupProbe"`
	// volume_mounts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#volume_mounts CloudRunV2Service#volume_mounts}
	VolumeMounts interface{} `field:"optional" json:"volumeMounts" yaml:"volumeMounts"`
	// Container's working directory.
	//
	// If not specified, the container runtime's default will be used, which might be configured in the container image.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#working_dir CloudRunV2Service#working_dir}
	WorkingDir *string `field:"optional" json:"workingDir" yaml:"workingDir"`
}

