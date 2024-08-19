package pubsubtopic


type PubsubTopicIngestionDataSourceSettings struct {
	// aws_kinesis block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#aws_kinesis PubsubTopic#aws_kinesis}
	AwsKinesis *PubsubTopicIngestionDataSourceSettingsAwsKinesis `field:"optional" json:"awsKinesis" yaml:"awsKinesis"`
}

