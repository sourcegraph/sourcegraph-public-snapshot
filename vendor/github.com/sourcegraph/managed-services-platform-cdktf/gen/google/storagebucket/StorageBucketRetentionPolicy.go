package storagebucket


type StorageBucketRetentionPolicy struct {
	// The period of time, in seconds, that objects in the bucket must be retained and cannot be deleted, overwritten, or archived.
	//
	// The value must be less than 3,155,760,000 seconds.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#retention_period StorageBucket#retention_period}
	RetentionPeriod *float64 `field:"required" json:"retentionPeriod" yaml:"retentionPeriod"`
	// If set to true, the bucket will be locked and permanently restrict edits to the bucket's retention policy.
	//
	// Caution: Locking a bucket is an irreversible action.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#is_locked StorageBucket#is_locked}
	IsLocked interface{} `field:"optional" json:"isLocked" yaml:"isLocked"`
}

