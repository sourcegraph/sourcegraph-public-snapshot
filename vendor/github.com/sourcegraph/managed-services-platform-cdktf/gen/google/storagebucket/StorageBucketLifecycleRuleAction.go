package storagebucket


type StorageBucketLifecycleRuleAction struct {
	// The type of the action of this Lifecycle Rule. Supported values include: Delete, SetStorageClass and AbortIncompleteMultipartUpload.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#type StorageBucket#type}
	Type *string `field:"required" json:"type" yaml:"type"`
	// The target Storage Class of objects affected by this Lifecycle Rule. Supported values include: MULTI_REGIONAL, REGIONAL, NEARLINE, COLDLINE, ARCHIVE.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#storage_class StorageBucket#storage_class}
	StorageClass *string `field:"optional" json:"storageClass" yaml:"storageClass"`
}

