package pubsubsubscription

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type PubsubSubscriptionConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// Name of the subscription.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#name PubsubSubscription#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// A reference to a Topic resource, of the form projects/{project}/topics/{{name}} (as in the id property of a google_pubsub_topic), or just a topic name if the topic is in the same project as the subscription.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#topic PubsubSubscription#topic}
	Topic *string `field:"required" json:"topic" yaml:"topic"`
	// This value is the maximum time after a subscriber receives a message before the subscriber should acknowledge the message.
	//
	// After message
	// delivery but before the ack deadline expires and before the message is
	// acknowledged, it is an outstanding message and will not be delivered
	// again during that time (on a best-effort basis).
	//
	// For pull subscriptions, this value is used as the initial value for
	// the ack deadline. To override this value for a given message, call
	// subscriptions.modifyAckDeadline with the corresponding ackId if using
	// pull. The minimum custom deadline you can specify is 10 seconds. The
	// maximum custom deadline you can specify is 600 seconds (10 minutes).
	// If this parameter is 0, a default value of 10 seconds is used.
	//
	// For push delivery, this value is also used to set the request timeout
	// for the call to the push endpoint.
	//
	// If the subscriber never acknowledges the message, the Pub/Sub system
	// will eventually redeliver the message.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#ack_deadline_seconds PubsubSubscription#ack_deadline_seconds}
	AckDeadlineSeconds *float64 `field:"optional" json:"ackDeadlineSeconds" yaml:"ackDeadlineSeconds"`
	// bigquery_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#bigquery_config PubsubSubscription#bigquery_config}
	BigqueryConfig *PubsubSubscriptionBigqueryConfig `field:"optional" json:"bigqueryConfig" yaml:"bigqueryConfig"`
	// cloud_storage_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#cloud_storage_config PubsubSubscription#cloud_storage_config}
	CloudStorageConfig *PubsubSubscriptionCloudStorageConfig `field:"optional" json:"cloudStorageConfig" yaml:"cloudStorageConfig"`
	// dead_letter_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#dead_letter_policy PubsubSubscription#dead_letter_policy}
	DeadLetterPolicy *PubsubSubscriptionDeadLetterPolicy `field:"optional" json:"deadLetterPolicy" yaml:"deadLetterPolicy"`
	// If 'true', Pub/Sub provides the following guarantees for the delivery of a message with a given value of messageId on this Subscriptions':  - The message sent to a subscriber is guaranteed not to be resent before the message's acknowledgement deadline expires.
	//
	// - An acknowledged message will not be resent to a subscriber.
	//
	// Note that subscribers may still receive multiple copies of a message when 'enable_exactly_once_delivery'
	// is true if the message was published multiple times by a publisher client. These copies are considered distinct by Pub/Sub and have distinct messageId values
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#enable_exactly_once_delivery PubsubSubscription#enable_exactly_once_delivery}
	EnableExactlyOnceDelivery interface{} `field:"optional" json:"enableExactlyOnceDelivery" yaml:"enableExactlyOnceDelivery"`
	// If 'true', messages published with the same orderingKey in PubsubMessage will be delivered to the subscribers in the order in which they are received by the Pub/Sub system.
	//
	// Otherwise, they
	// may be delivered in any order.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#enable_message_ordering PubsubSubscription#enable_message_ordering}
	EnableMessageOrdering interface{} `field:"optional" json:"enableMessageOrdering" yaml:"enableMessageOrdering"`
	// expiration_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#expiration_policy PubsubSubscription#expiration_policy}
	ExpirationPolicy *PubsubSubscriptionExpirationPolicy `field:"optional" json:"expirationPolicy" yaml:"expirationPolicy"`
	// The subscription only delivers the messages that match the filter.
	//
	// Pub/Sub automatically acknowledges the messages that don't match the filter. You can filter messages
	// by their attributes. The maximum length of a filter is 256 bytes. After creating the subscription,
	// you can't modify the filter.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#filter PubsubSubscription#filter}
	Filter *string `field:"optional" json:"filter" yaml:"filter"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#id PubsubSubscription#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// A set of key/value label pairs to assign to this Subscription.
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#labels PubsubSubscription#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// How long to retain unacknowledged messages in the subscription's backlog, from the moment a message is published.
	//
	// If
	// retain_acked_messages is true, then this also configures the retention
	// of acknowledged messages, and thus configures how far back in time a
	// subscriptions.seek can be done. Defaults to 7 days. Cannot be more
	// than 7 days ('"604800s"') or less than 10 minutes ('"600s"').
	//
	// A duration in seconds with up to nine fractional digits, terminated
	// by 's'. Example: '"600.5s"'.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#message_retention_duration PubsubSubscription#message_retention_duration}
	MessageRetentionDuration *string `field:"optional" json:"messageRetentionDuration" yaml:"messageRetentionDuration"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#project PubsubSubscription#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// push_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#push_config PubsubSubscription#push_config}
	PushConfig *PubsubSubscriptionPushConfig `field:"optional" json:"pushConfig" yaml:"pushConfig"`
	// Indicates whether to retain acknowledged messages.
	//
	// If 'true', then
	// messages are not expunged from the subscription's backlog, even if
	// they are acknowledged, until they fall out of the
	// messageRetentionDuration window.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#retain_acked_messages PubsubSubscription#retain_acked_messages}
	RetainAckedMessages interface{} `field:"optional" json:"retainAckedMessages" yaml:"retainAckedMessages"`
	// retry_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#retry_policy PubsubSubscription#retry_policy}
	RetryPolicy *PubsubSubscriptionRetryPolicy `field:"optional" json:"retryPolicy" yaml:"retryPolicy"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#timeouts PubsubSubscription#timeouts}
	Timeouts *PubsubSubscriptionTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}

