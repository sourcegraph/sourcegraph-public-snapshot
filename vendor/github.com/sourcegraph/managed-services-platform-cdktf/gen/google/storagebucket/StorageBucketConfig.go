package storagebucket

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type StorageBucketConfig struct {
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
	// The Google Cloud Storage location.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#location StorageBucket#location}
	Location *string `field:"required" json:"location" yaml:"location"`
	// The name of the bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#name StorageBucket#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// autoclass block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#autoclass StorageBucket#autoclass}
	Autoclass *StorageBucketAutoclass `field:"optional" json:"autoclass" yaml:"autoclass"`
	// cors block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#cors StorageBucket#cors}
	Cors interface{} `field:"optional" json:"cors" yaml:"cors"`
	// custom_placement_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#custom_placement_config StorageBucket#custom_placement_config}
	CustomPlacementConfig *StorageBucketCustomPlacementConfig `field:"optional" json:"customPlacementConfig" yaml:"customPlacementConfig"`
	// Whether or not to automatically apply an eventBasedHold to new objects added to the bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#default_event_based_hold StorageBucket#default_event_based_hold}
	DefaultEventBasedHold interface{} `field:"optional" json:"defaultEventBasedHold" yaml:"defaultEventBasedHold"`
	// Enables each object in the bucket to have its own retention policy, which prevents deletion until stored for a specific length of time.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#enable_object_retention StorageBucket#enable_object_retention}
	EnableObjectRetention interface{} `field:"optional" json:"enableObjectRetention" yaml:"enableObjectRetention"`
	// encryption block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#encryption StorageBucket#encryption}
	Encryption *StorageBucketEncryption `field:"optional" json:"encryption" yaml:"encryption"`
	// When deleting a bucket, this boolean option will delete all contained objects.
	//
	// If you try to delete a bucket that contains objects, Terraform will fail that run.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#force_destroy StorageBucket#force_destroy}
	ForceDestroy interface{} `field:"optional" json:"forceDestroy" yaml:"forceDestroy"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#id StorageBucket#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// A set of key/value label pairs to assign to the bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#labels StorageBucket#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// lifecycle_rule block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#lifecycle_rule StorageBucket#lifecycle_rule}
	LifecycleRule interface{} `field:"optional" json:"lifecycleRule" yaml:"lifecycleRule"`
	// logging block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#logging StorageBucket#logging}
	Logging *StorageBucketLogging `field:"optional" json:"logging" yaml:"logging"`
	// The ID of the project in which the resource belongs.
	//
	// If it is not provided, the provider project is used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#project StorageBucket#project}
	Project *string `field:"optional" json:"project" yaml:"project"`
	// Prevents public access to a bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#public_access_prevention StorageBucket#public_access_prevention}
	PublicAccessPrevention *string `field:"optional" json:"publicAccessPrevention" yaml:"publicAccessPrevention"`
	// Enables Requester Pays on a storage bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#requester_pays StorageBucket#requester_pays}
	RequesterPays interface{} `field:"optional" json:"requesterPays" yaml:"requesterPays"`
	// retention_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#retention_policy StorageBucket#retention_policy}
	RetentionPolicy *StorageBucketRetentionPolicy `field:"optional" json:"retentionPolicy" yaml:"retentionPolicy"`
	// Specifies the RPO setting of bucket.
	//
	// If set 'ASYNC_TURBO', The Turbo Replication will be enabled for the dual-region bucket. Value 'DEFAULT' will set RPO setting to default. Turbo Replication is only for buckets in dual-regions.See the docs for more details.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#rpo StorageBucket#rpo}
	Rpo *string `field:"optional" json:"rpo" yaml:"rpo"`
	// soft_delete_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#soft_delete_policy StorageBucket#soft_delete_policy}
	SoftDeletePolicy *StorageBucketSoftDeletePolicy `field:"optional" json:"softDeletePolicy" yaml:"softDeletePolicy"`
	// The Storage Class of the new bucket. Supported values include: STANDARD, MULTI_REGIONAL, REGIONAL, NEARLINE, COLDLINE, ARCHIVE.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#storage_class StorageBucket#storage_class}
	StorageClass *string `field:"optional" json:"storageClass" yaml:"storageClass"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#timeouts StorageBucket#timeouts}
	Timeouts *StorageBucketTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// Enables uniform bucket-level access on a bucket.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#uniform_bucket_level_access StorageBucket#uniform_bucket_level_access}
	UniformBucketLevelAccess interface{} `field:"optional" json:"uniformBucketLevelAccess" yaml:"uniformBucketLevelAccess"`
	// versioning block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#versioning StorageBucket#versioning}
	Versioning *StorageBucketVersioning `field:"optional" json:"versioning" yaml:"versioning"`
	// website block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#website StorageBucket#website}
	Website *StorageBucketWebsite `field:"optional" json:"website" yaml:"website"`
}

