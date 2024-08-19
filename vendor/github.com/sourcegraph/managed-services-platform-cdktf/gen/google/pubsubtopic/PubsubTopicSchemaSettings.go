package pubsubtopic


type PubsubTopicSchemaSettings struct {
	// The name of the schema that messages published should be validated against.
	//
	// Format is projects/{project}/schemas/{schema}.
	// The value of this field will be _deleted-schema_
	// if the schema has been deleted.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#schema PubsubTopic#schema}
	Schema *string `field:"required" json:"schema" yaml:"schema"`
	// The encoding of messages validated against schema. Default value: "ENCODING_UNSPECIFIED" Possible values: ["ENCODING_UNSPECIFIED", "JSON", "BINARY"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_topic#encoding PubsubTopic#encoding}
	Encoding *string `field:"optional" json:"encoding" yaml:"encoding"`
}

