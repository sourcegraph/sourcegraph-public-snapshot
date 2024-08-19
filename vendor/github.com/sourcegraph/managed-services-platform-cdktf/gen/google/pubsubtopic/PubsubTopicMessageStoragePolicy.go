package pubsubtopic


type PubsubTopicMessageStoragePolicy struct {
	// A list of IDs of GCP regions where messages that are published to the topic may be persisted in storage.
	//
	// Messages published by
	// publishers running in non-allowed GCP regions (or running outside
	// of GCP altogether) will be routed for storage in one of the
	// allowed regions. An empty list means that no regions are allowed,
	// and is not a valid configuration.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#allowed_persistence_regions PubsubTopic#allowed_persistence_regions}
	AllowedPersistenceRegions *[]*string `field:"required" json:"allowedPersistenceRegions" yaml:"allowedPersistenceRegions"`
}

