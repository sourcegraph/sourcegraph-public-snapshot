package storagebucket


type StorageBucketAutoclass struct {
	// While set to true, autoclass automatically transitions objects in your bucket to appropriate storage classes based on each object's access pattern.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#enabled StorageBucket#enabled}
	Enabled interface{} `field:"required" json:"enabled" yaml:"enabled"`
	// The storage class that objects in the bucket eventually transition to if they are not read for a certain length of time.
	//
	// Supported values include: NEARLINE, ARCHIVE.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/storage_bucket#terminal_storage_class StorageBucket#terminal_storage_class}
	TerminalStorageClass *string `field:"optional" json:"terminalStorageClass" yaml:"terminalStorageClass"`
}

