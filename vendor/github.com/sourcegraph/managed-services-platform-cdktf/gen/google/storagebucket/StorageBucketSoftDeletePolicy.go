package storagebucket


type StorageBucketSoftDeletePolicy struct {
	// The duration in seconds that soft-deleted objects in the bucket will be retained and cannot be permanently deleted.
	//
	// Default value is 604800.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#retention_duration_seconds StorageBucket#retention_duration_seconds}
	RetentionDurationSeconds *float64 `field:"optional" json:"retentionDurationSeconds" yaml:"retentionDurationSeconds"`
}

