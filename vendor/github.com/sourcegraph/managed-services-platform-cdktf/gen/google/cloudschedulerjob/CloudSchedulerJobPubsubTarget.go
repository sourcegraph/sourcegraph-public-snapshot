package cloudschedulerjob


type CloudSchedulerJobPubsubTarget struct {
	// The full resource name for the Cloud Pub/Sub topic to which messages will be published when a job is delivered.
	//
	// ~>**NOTE:**
	// The topic name must be in the same format as required by PubSub's
	// PublishRequest.name, e.g. 'projects/my-project/topics/my-topic'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#topic_name CloudSchedulerJob#topic_name}
	TopicName *string `field:"required" json:"topicName" yaml:"topicName"`
	// Attributes for PubsubMessage. Pubsub message must contain either non-empty data, or at least one attribute.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#attributes CloudSchedulerJob#attributes}
	Attributes *map[string]*string `field:"optional" json:"attributes" yaml:"attributes"`
	// The message payload for PubsubMessage. Pubsub message must contain either non-empty data, or at least one attribute.
	//
	// A base64-encoded string.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_scheduler_job#data CloudSchedulerJob#data}
	Data *string `field:"optional" json:"data" yaml:"data"`
}

