package directgcm

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type DirectGcmConfig struct {
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
	// Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#name DirectGcm#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#project DirectGcm#project}
	Project *string `field:"required" json:"project" yaml:"project"`
	// Source of Metrics and/or Services.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#source_of DirectGcm#source_of}
	SourceOf *[]*string `field:"required" json:"sourceOf" yaml:"sourceOf"`
	// Optional description of the resource.
	//
	// Here, you can add details about who is responsible for the integration (team/owner) or the purpose of creating it.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#description DirectGcm#description}
	Description *string `field:"optional" json:"description" yaml:"description"`
	// User-friendly display name of the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#display_name DirectGcm#display_name}
	DisplayName *string `field:"optional" json:"displayName" yaml:"displayName"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#id DirectGcm#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// [Logs documentation](https://docs.nobl9.com/Features/SLO_troubleshooting/event-logs).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#log_collection_enabled DirectGcm#log_collection_enabled}
	LogCollectionEnabled interface{} `field:"optional" json:"logCollectionEnabled" yaml:"logCollectionEnabled"`
	// query_delay block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#query_delay DirectGcm#query_delay}
	QueryDelay *DirectGcmQueryDelay `field:"optional" json:"queryDelay" yaml:"queryDelay"`
	// Release channel of the created datasource [stable/beta].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#release_channel DirectGcm#release_channel}
	ReleaseChannel *string `field:"optional" json:"releaseChannel" yaml:"releaseChannel"`
	// [required] | Service Account Key.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs/resources/direct_gcm#service_account_key DirectGcm#service_account_key}
	ServiceAccountKey *string `field:"optional" json:"serviceAccountKey" yaml:"serviceAccountKey"`
}

