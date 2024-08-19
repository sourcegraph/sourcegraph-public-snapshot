package pubsubsubscription


type PubsubSubscriptionCloudStorageConfig struct {
	// User-provided name for the Cloud Storage bucket.
	//
	// The bucket must be created by the user. The bucket name must be without any prefix like "gs://".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#bucket PubsubSubscription#bucket}
	Bucket *string `field:"required" json:"bucket" yaml:"bucket"`
	// avro_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#avro_config PubsubSubscription#avro_config}
	AvroConfig *PubsubSubscriptionCloudStorageConfigAvroConfig `field:"optional" json:"avroConfig" yaml:"avroConfig"`
	// User-provided prefix for Cloud Storage filename.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#filename_prefix PubsubSubscription#filename_prefix}
	FilenamePrefix *string `field:"optional" json:"filenamePrefix" yaml:"filenamePrefix"`
	// User-provided suffix for Cloud Storage filename. Must not end in "/".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#filename_suffix PubsubSubscription#filename_suffix}
	FilenameSuffix *string `field:"optional" json:"filenameSuffix" yaml:"filenameSuffix"`
	// The maximum bytes that can be written to a Cloud Storage file before a new file is created.
	//
	// Min 1 KB, max 10 GiB.
	// The maxBytes limit may be exceeded in cases where messages are larger than the limit.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#max_bytes PubsubSubscription#max_bytes}
	MaxBytes *float64 `field:"optional" json:"maxBytes" yaml:"maxBytes"`
	// The maximum duration that can elapse before a new Cloud Storage file is created.
	//
	// Min 1 minute, max 10 minutes, default 5 minutes.
	// May not exceed the subscription's acknowledgement deadline.
	// A duration in seconds with up to nine fractional digits, ending with 's'. Example: "3.5s".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#max_duration PubsubSubscription#max_duration}
	MaxDuration *string `field:"optional" json:"maxDuration" yaml:"maxDuration"`
}

