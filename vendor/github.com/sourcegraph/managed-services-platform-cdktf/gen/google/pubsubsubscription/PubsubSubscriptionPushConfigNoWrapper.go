package pubsubsubscription


type PubsubSubscriptionPushConfigNoWrapper struct {
	// When true, writes the Pub/Sub message metadata to 'x-goog-pubsub-<KEY>:<VAL>' headers of the HTTP request.
	//
	// Writes the
	// Pub/Sub message attributes to '<KEY>:<VAL>' headers of the HTTP request.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#write_metadata PubsubSubscription#write_metadata}
	WriteMetadata interface{} `field:"required" json:"writeMetadata" yaml:"writeMetadata"`
}

