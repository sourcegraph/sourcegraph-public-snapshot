package pubsubsubscription


type PubsubSubscriptionPushConfig struct {
	// A URL locating the endpoint to which messages should be pushed. For example, a Webhook endpoint might use "https://example.com/push".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#push_endpoint PubsubSubscription#push_endpoint}
	PushEndpoint *string `field:"required" json:"pushEndpoint" yaml:"pushEndpoint"`
	// Endpoint configuration attributes.
	//
	// Every endpoint has a set of API supported attributes that can
	// be used to control different aspects of the message delivery.
	//
	// The currently supported attribute is x-goog-version, which you
	// can use to change the format of the pushed message. This
	// attribute indicates the version of the data expected by
	// the endpoint. This controls the shape of the pushed message
	// (i.e., its fields and metadata). The endpoint version is
	// based on the version of the Pub/Sub API.
	//
	// If not present during the subscriptions.create call,
	// it will default to the version of the API used to make
	// such call. If not present during a subscriptions.modifyPushConfig
	// call, its value will not be changed. subscriptions.get
	// calls will always return a valid version, even if the
	// subscription was created without this attribute.
	//
	// The possible values for this attribute are:
	//
	// - v1beta1: uses the push format defined in the v1beta1 Pub/Sub API.
	// - v1 or v1beta2: uses the push format defined in the v1 Pub/Sub API.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#attributes PubsubSubscription#attributes}
	Attributes *map[string]*string `field:"optional" json:"attributes" yaml:"attributes"`
	// no_wrapper block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#no_wrapper PubsubSubscription#no_wrapper}
	NoWrapper *PubsubSubscriptionPushConfigNoWrapper `field:"optional" json:"noWrapper" yaml:"noWrapper"`
	// oidc_token block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/pubsub_subscription#oidc_token PubsubSubscription#oidc_token}
	OidcToken *PubsubSubscriptionPushConfigOidcToken `field:"optional" json:"oidcToken" yaml:"oidcToken"`
}

