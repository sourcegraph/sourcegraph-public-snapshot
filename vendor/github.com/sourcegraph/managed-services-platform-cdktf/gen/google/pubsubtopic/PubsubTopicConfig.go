package pubsubtopic

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type PubsubTopicConfig struct {
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
	// Name of the topic.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#name PubsubTopic#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#id PubsubTopic#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// ingestion_data_source_settings block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#ingestion_data_source_settings PubsubTopic#ingestion_data_source_settings}
	IngestionDataSourceSettings *PubsubTopicIngestionDataSourceSettings `field:"optional" json:"ingestionDataSourceSettings" yaml:"ingestionDataSourceSettings"`
	// The resource name of the Cloud KMS CryptoKey to be used to protect access to messages published on this topic.
	//
	// Your project's PubSub service account
	// ('service-{{PROJECT_NUMBER}}@gcp-sa-pubsub.iam.gserviceaccount.com') must have
	// 'roles/cloudkms.cryptoKeyEncrypterDecrypter' to use this feature.
	// The expected format is 'projects/*\/locations/*\/keyRings/*\/cryptoKeys/*'
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#kms_key_name PubsubTopic#kms_key_name}
	KmsKeyName *string `field:"optional" json:"kmsKeyName" yaml:"kmsKeyName"`
	// A set of key/value label pairs to assign to this Topic.
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#labels PubsubTopic#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// Indicates the minimum duration to retain a message after it is published to the topic.
	//
	// If this field is set, messages published to the topic in
	// the last messageRetentionDuration are always available to subscribers.
	// For instance, it allows any attached subscription to seek to a timestamp
	// that is up to messageRetentionDuration in the past. If this field is not
	// set, message retention is controlled by settings on individual subscriptions.
	// The rotation period has the format of a decimal number, followed by the
	// letter 's' (seconds). Cannot be more than 31 days or less than 10 minutes.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#message_retention_duration PubsubTopic#message_retention_duration}
	MessageRetentionDuration *string `field:"optional" json:"messageRetentionDuration" yaml:"messageRetentionDuration"`
	// message_storage_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#message_storage_policy PubsubTopic#message_storage_policy}
	MessageStoragePolicy *PubsubTopicMessageStoragePolicy `field:"optional" json:"messageStoragePolicy" yaml:"messageStoragePolicy"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#project PubsubTopic#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// schema_settings block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#schema_settings PubsubTopic#schema_settings}
	SchemaSettings *PubsubTopicSchemaSettings `field:"optional" json:"schemaSettings" yaml:"schemaSettings"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#timeouts PubsubTopic#timeouts}
	Timeouts *PubsubTopicTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
}

