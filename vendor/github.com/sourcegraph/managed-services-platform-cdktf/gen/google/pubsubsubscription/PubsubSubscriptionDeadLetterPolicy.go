package pubsubsubscription


type PubsubSubscriptionDeadLetterPolicy struct {
	// The name of the topic to which dead letter messages should be published. Format is 'projects/{project}/topics/{topic}'.
	//
	// The Cloud Pub/Sub service account associated with the enclosing subscription's
	// parent project (i.e.,
	// service-{project_number}@gcp-sa-pubsub.iam.gserviceaccount.com) must have
	// permission to Publish() to this topic.
	//
	// The operation will fail if the topic does not exist.
	// Users should ensure that there is a subscription attached to this topic
	// since messages published to a topic with no subscriptions are lost.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#dead_letter_topic PubsubSubscription#dead_letter_topic}
	DeadLetterTopic *string `field:"optional" json:"deadLetterTopic" yaml:"deadLetterTopic"`
	// The maximum number of delivery attempts for any message. The value must be between 5 and 100.
	//
	// The number of delivery attempts is defined as 1 + (the sum of number of
	// NACKs and number of times the acknowledgement deadline has been exceeded for the message).
	//
	// A NACK is any call to ModifyAckDeadline with a 0 deadline. Note that
	// client libraries may automatically extend ack_deadlines.
	//
	// This field will be honored on a best effort basis.
	//
	// If this parameter is 0, a default value of 5 is used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#max_delivery_attempts PubsubSubscription#max_delivery_attempts}
	MaxDeliveryAttempts *float64 `field:"optional" json:"maxDeliveryAttempts" yaml:"maxDeliveryAttempts"`
}

