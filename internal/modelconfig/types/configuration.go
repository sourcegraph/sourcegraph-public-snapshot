package types

// ========================================================
// Client-side Provider Configuration Data
// ========================================================

type ClientSideProviderConfig struct {
	// We currently do not have any known client-side provider configuration.
	// But later, if anything needs to be provided to Cody clients at the
	// provider-level it will go here.
}

// ========================================================
// Server-side Provider Configuration Data
// ========================================================

// https://sourcegraph.com/docs/cody/clients/enable-cody-enterprise#use-amazon-bedrock-aws
type AWSBedrockProviderConfig struct {
	// Access token encodes your AWS credentials one one of several ways:
	// - Leave it empty and rely on instance role bindings or other AWS configurations in the frontend service
	// - Set it to <ACCESS_KEY_ID>:<SECRET_ACCESS_KEY> if directly configuring the credentials
	// - Set it to <ACCESS_KEY_ID>:<SECRET_ACCESS_KEY>:<SESSION_TOKEN> if a session token is also required.
	AccessToken string `json:"accessToken"`
	// - For Pay-as-you-go, set it to an AWS region code (e.g., us-west-2) when using a public Amazon Bedrock endpoint
	// - For Provisioned Throughput, set it to the provisioned VPC endpoint for the bedrock-runtime API (e.g., "https://vpce-0a10b2345cd67e89f-abc0defg.bedrock-runtime.us-west-2.vpce.amazonaws.com")
	Endpoint string `json:"endpoint"`
	// Region to use when configuring API clients. (Since the `frontend` binary's container won't
	// be able to pick this up from the host OS's environment variables.)
	Region string `json:"region"`
}

// https://sourcegraph.com/docs/cody/clients/enable-cody-enterprise#use-azure-openai-service
type AzureOpenAIProviderConfig struct {
	// - As of 5.2.4 the access token can be left empty and it will rely on Environmental, Workload Identity or Managed Identity credentials configured for the frontend and worker services
	// - Set it to <API_KEY> if directly configuring the credentials using the API key specified in the Azure portal
	AccessToken string `json:"accessToken"`
	// From the Azure OpenAI Service portal.
	Endpoint string `json:"endpoint"`

	// The user field passed along to OpenAI-provided models.
	User string `json:"user"`
	// Enables the use of the older completions API for select Azure OpenAI models. This is just an escape hatch
	// for backwards compatibility, because not all Azure OpenAI models are available on the "newer" completions API.
	//
	// Moving forward, this information should be encoded in the ModelRef's APIVersionID instead.
	UseDeprecatedCompletionsAPI bool `json:"useDeprecatedCompletionsAPI"`
}

// GenericServiceProvider is an enum for describing the API provider to use
// for a GenericProviderConfig. These should generally be a subset of what
// is available via site config, `conftypes.CompletionsProviderName`.
type GenericServiceProvider string

const (
	GenericServiceProviderAnthropic GenericServiceProvider = "anthropic"
	GenericServiceProviderFireworks GenericServiceProvider = "fireworks"
	GenericServiceProviderGoogle    GenericServiceProvider = "google"
	GenericServiceProviderOpenAI    GenericServiceProvider = "openai"
)

// GenericProvider just bundles the bare-bones information needed to describe a 3rd party API.
//
// You should NEVER add new fields to this. Instead, if we wish to expose some provider-specific
// configuration knob please introduce a new data type specific to that provider. (Rather than
// adding a field to GenericProviderConfig that will not be applicable or ignored in some cases.)
type GenericProviderConfig struct {
	// ServiceName is the name of the LLM model provider or service that this generic provider
	// config applies to.
	ServiceName GenericServiceProvider `json:"serviceName"`

	AccessToken string `json:"accessToken"`
	Endpoint    string `json:"endpoint"`
}

// OpenAICompatibleServiceName is an enum for describing the names of services/software
// which provide OpenAI-compatible APIs that we are currently aware of.
type OpenAICompatibleServiceName string

const (
	OpenAICompatibleServiceNameHuggingfaceTGI OpenAICompatibleServiceName = "huggingface-tgi"
	OpenAICompatibleServiceNameNvidiaNim      OpenAICompatibleServiceName = "nvidia-nim"
	OpenAICompatibleServiceNameAwsLisa        OpenAICompatibleServiceName = "aws-lisa"
	OpenAICompatibleServiceNameAwsLisaV2      OpenAICompatibleServiceName = "aws-lisa-v2"
	OpenAICompatibleServiceNameOllama         OpenAICompatibleServiceName = "ollama"
)

// Returns (enum, ok bool) depending on whether or not s is a valid OpenAICompatibleServiceName
// constant.
func ValidateOpenAICompatibleServiceName(input string) (OpenAICompatibleServiceName, bool) {
	for _, v := range []OpenAICompatibleServiceName{
		OpenAICompatibleServiceNameHuggingfaceTGI,
		OpenAICompatibleServiceNameNvidiaNim,
		OpenAICompatibleServiceNameAwsLisa,
		OpenAICompatibleServiceNameAwsLisaV2,
		OpenAICompatibleServiceNameOllama,
	} {
		if input == string(v) {
			return v, true
		}
	}
	return "", false
}

// OpenAICompatibleProvider is a provider for connecting to OpenAI-compatible API endpoints
// supplied by various third-party software. Much of this software does not follow the OpenAI
// API protocol exactly, and so this provider configuration allows for much more extensive
// configuration and even logic aware of the service on the other side.
type OpenAICompatibleProviderConfig struct {
	// ServiceName is the name of the service being connected to, e.g. "huggingface-tgi" being the
	// name of the software actually hosting the OpenAI-compatible API endpoint. This is used to
	// handle any nuance in communicating with that service, as different software has various
	// bugs/quirks we may want to handle differently.
	//
	// This field may only be used if the service name is one we are aware of, i.e. a valid enum
	// OpenAICompatibleServiceName.
	ServiceName OpenAICompatibleServiceName `json:"serviceName,omitempty"`

	// ServiceNameCustom is exactly the same as ServiceName, except it is used if the service name
	// is not one currently known to Sourcegraph, i.e. it is not a valid OpenAICompatibleServiceName
	// enum but rather just an arbitrary string the site admin plugged in. e.g. "my-enterprise-openai-api"
	// or "brand-new-openai-compatible-software-that-is-popular"
	ServiceNameCustom string `json:"serviceNameCustom,omitempty"`

	// Endpoints where this API can be reached. If multiple are present, Sourcegraph should distribute
	// load between them.
	Endpoints []OpenAICompatibleEndpoint `json:"endpoints,omitempty"`
}

// A single API endpoint for an OpenAI-compatible API.
type OpenAICompatibleEndpoint struct {
	URL         string `json:"url"`
	AccessToken string `json:"accessToken"`
}

// SourcegraphProviderConfig is the configuration blog for configuring a provider
// to be use Sourcegraph's Cody Gateway for requests.
type SourcegraphProviderConfig struct {
	AccessToken string `json:"accessToken"`
	Endpoint    string `json:"endpoint"`
}

// The "Provider" is conceptually a namespace for models. The server-side provider configuration
// is needed to describe the API endpoint needed to serve its models.
type ServerSideProviderConfig struct {
	AWSBedrock          *AWSBedrockProviderConfig       `json:"awsBedrock,omitempty"`
	AzureOpenAI         *AzureOpenAIProviderConfig      `json:"azureOpenAi,omitempty"`
	OpenAICompatible    *OpenAICompatibleProviderConfig `json:"openAICompatible,omitempty"`
	GenericProvider     *GenericProviderConfig          `json:"genericProvider,omitempty"`
	SourcegraphProvider *SourcegraphProviderConfig      `json:"sourcegraphProvider,omitempty"`
}

// ========================================================
// Client-side Model Configuration Data
// ========================================================

type ClientSideModelConfig struct {
	// We currently do not have any known client-side model configuration.
	// But later, if anything needs to be provided to Cody clients at the
	// model-level it will go here.
	//
	// For example, allowing the server to customize/override the LLM
	// prompt used. Or describe how clients should upload context to
	// remote servers, etc. Or "hints", like "this model is great when
	// working with 'C' code.".
	OpenAICompatible *ClientSideModelConfigOpenAICompatible `json:"openAICompatible,omitempty"`
}

// Client-side model configuration used when the model is being provided by an OpenAI-compatible
// API.
type ClientSideModelConfigOpenAICompatible struct {
	// (optional) List of stop sequences to use for this model.
	StopSequences []string `json:"stopSequences,omitempty"`

	// (optional) EndOfText identifier used by the model. e.g. "<|endoftext|>", "<EOT>"
	EndOfText string `json:"endOfText,omitempty"`

	// (optional) A hint the client should use when producing context to send to the LLM.
	// The maximum length of all context (prefix + suffix + snippets), in characters.
	ContextSizeHintTotalCharacters *uint `json:"contextSizeHintTotalCharacters,omitempty"`

	// (optional) A hint the client should use when producing context to send to the LLM.
	// The maximum length of the document prefix (text before the cursor) to include, in characters.
	ContextSizeHintPrefixCharacters *uint `json:"contextSizeHintPrefixCharacters,omitempty"`

	// (optional) A hint the client should use when producing context to send to the LLM.
	// The maximum length of the document suffix (text after the cursor) to include, in characters.
	ContextSizeHintSuffixCharacters *uint `json:"contextSizeHintSuffixCharacters,omitempty"`

	// (optional) Custom instruction to be included at the start of all chat messages
	// when using this model, e.g. "Answer all questions in Spanish."
	//
	// Note: similar to Cody client config option `cody.chat.preInstruction`; if user has
	// configured that it will be used instead of this.
	ChatPreInstruction string `json:"chatPreInstruction,omitempty"`

	// (optional) Custom instruction to be included at the end of all edit commands
	// when using this model, e.g. "Write all unit tests with Jest instead of detected framework."
	//
	// Note: similar to Cody client config option `cody.edit.preInstruction`; if user has
	// configured that it will be respected instead of this.
	EditPostInstruction string `json:"editPostInstruction,omitempty"`

	// (optional) How long the client should wait for autocomplete results to come back (milliseconds),
	// before giving up and not displaying an autocomplete result at all.
	//
	// This applies on single-line completions, e.g. `var i = <completion>`
	//
	// Note: similar to hidden Cody client config option `cody.autocomplete.advanced.timeout.singleline`
	// If user has configured that, it will be respected instead of this.
	AutocompleteSinglelineTimeout uint `json:"autocompleteSinglelineTimeout,omitempty"`

	// (optional) How long the client should wait for autocomplete results to come back (milliseconds),
	// before giving up and not displaying an autocomplete result at all.
	//
	// This applies on multi-line completions, which are based on intent-detection when e.g. a code block
	// is being completed, e.g. `func parseURL(url string) {<completion>`
	//
	// Note: similar to hidden Cody client config option `cody.autocomplete.advanced.timeout.multiline`
	// If user has configured that, it will be respected instead of this.
	AutocompleteMultilineTimeout uint `json:"autocompleteMultilineTimeout,omitempty"`

	// (optional) model parameters to use for the chat feature
	ChatTopK        float32 `json:"chatTopK,omitempty"`
	ChatTopP        float32 `json:"chatTopP,omitempty"`
	ChatTemperature float32 `json:"chatTemperature,omitempty"`
	ChatMaxTokens   uint    `json:"chatMaxTokens,omitempty"`

	// (optional) model parameters to use for the autocomplete feature
	AutoCompleteTopK                float32 `json:"autoCompleteTopK,omitempty"`
	AutoCompleteTopP                float32 `json:"autoCompleteTopP,omitempty"`
	AutoCompleteTemperature         float32 `json:"autoCompleteTemperature,omitempty"`
	AutoCompleteSinglelineMaxTokens uint    `json:"autoCompleteSinglelineMaxTokens,omitempty"`
	AutoCompleteMultilineMaxTokens  uint    `json:"autoCompleteMultilineMaxTokens,omitempty"`

	// (optional) model parameters to use for the edit feature
	EditTopK        float32 `json:"editTopK,omitempty"`
	EditTopP        float32 `json:"editTopP,omitempty"`
	EditTemperature float32 `json:"editTemperature,omitempty"`
	EditMaxTokens   uint    `json:"editMaxTokens,omitempty"`
}

// ========================================================
// Server-side Model Configuration Data
// ========================================================

// https://sourcegraph.com/docs/cody/clients/enable-cody-enterprise#use-amazon-bedrock-aws
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html#prov-throughput-models
// https://docs.aws.amazon.com/bedrock/latest/userguide/prov-throughput.html
type AWSBedrockProvisionedThroughput struct {
	// ARN is the "provisioned throughput ARN" to use when sending requests to AWS Bedrock
	// for a given model.
	ARN string `json:"arn"`
}

type ServerSideModelConfigOpenAICompatible struct {
	// APIModel is value actually sent to the OpenAI-compatible API in the "model" field. THis
	// is less like a "model name" or "model identifier", and more like "an opaque, potentially
	// secret string."
	//
	// Much software that claims to 'implement the OpenAI API' actually overrides this field with
	// other information NOT related to the model name, either making it _ineffective_ as a
	// model name/identifier (e.g. you must send "tgi" or "AUTODETECT" irrespective of which model
	// you want to use) OR using it to smuggle other (potentially sensitive) information like the
	// name of the deployment, which cannot be shared with clients.
	//
	// If this field is not an empty string, we treat it as an opaque string to be sent with API
	// requests (similar to an access token) and use it for nothing else. If this field is not
	// specified, we default to the ModelName.
	//
	// Examples (these would be sent in the OpenAI /chat/completions `"model"` field):
	//
	// * Huggingface TGI: "tgi"
	// * NVIDIA NIM: "meta/llama3-70b-instruct"
	// * AWS LISA (v2): "AUTODETECT"
	// * AWS LISA (v1): "mistralai/Mistral7b-v0.3-Instruct ecs.textgen.tgi"
	// * Ollama: "llama2"
	// * Others: "<SECRET DEPLOYMENT NAME>"
	//
	APIModel string `json:"apiModel,omitempty"`
}

type ServerSideModelConfig struct {
	AWSBedrockProvisionedThroughput *AWSBedrockProvisionedThroughput       `json:"awsBedrockProvisionedThroughput,omitempty"`
	OpenAICompatible                *ServerSideModelConfigOpenAICompatible `json:"openAICompatible,omitempty"`
}
