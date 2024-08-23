package pubsubsubscription


type PubsubSubscriptionRetryPolicy struct {
	// The maximum delay between consecutive deliveries of a given message.
	//
	// Value should be between 0 and 600 seconds. Defaults to 600 seconds.
	// A duration in seconds with up to nine fractional digits, terminated by 's'. Example: "3.5s".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#maximum_backoff PubsubSubscription#maximum_backoff}
	MaximumBackoff *string `field:"optional" json:"maximumBackoff" yaml:"maximumBackoff"`
	// The minimum delay between consecutive deliveries of a given message.
	//
	// Value should be between 0 and 600 seconds. Defaults to 10 seconds.
	// A duration in seconds with up to nine fractional digits, terminated by 's'. Example: "3.5s".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#minimum_backoff PubsubSubscription#minimum_backoff}
	MinimumBackoff *string `field:"optional" json:"minimumBackoff" yaml:"minimumBackoff"`
}

