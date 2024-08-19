package provider


type CloudflareProviderConfig struct {
	// Alias name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#alias CloudflareProvider#alias}
	Alias *string `field:"optional" json:"alias" yaml:"alias"`
	// Configure the base path used by the API client. Alternatively, can be configured using the `CLOUDFLARE_API_BASE_PATH` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#api_base_path CloudflareProvider#api_base_path}
	ApiBasePath *string `field:"optional" json:"apiBasePath" yaml:"apiBasePath"`
	// Whether to print logs from the API client (using the default log library logger).
	//
	// Alternatively, can be configured using the `CLOUDFLARE_API_CLIENT_LOGGING` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#api_client_logging CloudflareProvider#api_client_logging}
	ApiClientLogging interface{} `field:"optional" json:"apiClientLogging" yaml:"apiClientLogging"`
	// Configure the hostname used by the API client. Alternatively, can be configured using the `CLOUDFLARE_API_HOSTNAME` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#api_hostname CloudflareProvider#api_hostname}
	ApiHostname *string `field:"optional" json:"apiHostname" yaml:"apiHostname"`
	// The API key for operations.
	//
	// Alternatively, can be configured using the `CLOUDFLARE_API_KEY` environment variable. API keys are [now considered legacy by Cloudflare](https://developers.cloudflare.com/fundamentals/api/get-started/keys/#limitations), API tokens should be used instead. Must provide only one of `api_key`, `api_token`, `api_user_service_key`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#api_key CloudflareProvider#api_key}
	ApiKey *string `field:"optional" json:"apiKey" yaml:"apiKey"`
	// The API Token for operations.
	//
	// Alternatively, can be configured using the `CLOUDFLARE_API_TOKEN` environment variable. Must provide only one of `api_key`, `api_token`, `api_user_service_key`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#api_token CloudflareProvider#api_token}
	ApiToken *string `field:"optional" json:"apiToken" yaml:"apiToken"`
	// A special Cloudflare API key good for a restricted set of endpoints.
	//
	// Alternatively, can be configured using the `CLOUDFLARE_API_USER_SERVICE_KEY` environment variable. Must provide only one of `api_key`, `api_token`, `api_user_service_key`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#api_user_service_key CloudflareProvider#api_user_service_key}
	ApiUserServiceKey *string `field:"optional" json:"apiUserServiceKey" yaml:"apiUserServiceKey"`
	// A registered Cloudflare email address.
	//
	// Alternatively, can be configured using the `CLOUDFLARE_EMAIL` environment variable. Required when using `api_key`. Conflicts with `api_token`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#email CloudflareProvider#email}
	Email *string `field:"optional" json:"email" yaml:"email"`
	// Maximum backoff period in seconds after failed API calls. Alternatively, can be configured using the `CLOUDFLARE_MAX_BACKOFF` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#max_backoff CloudflareProvider#max_backoff}
	MaxBackoff *float64 `field:"optional" json:"maxBackoff" yaml:"maxBackoff"`
	// Minimum backoff period in seconds after failed API calls. Alternatively, can be configured using the `CLOUDFLARE_MIN_BACKOFF` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#min_backoff CloudflareProvider#min_backoff}
	MinBackoff *float64 `field:"optional" json:"minBackoff" yaml:"minBackoff"`
	// Maximum number of retries to perform when an API request fails.
	//
	// Alternatively, can be configured using the `CLOUDFLARE_RETRIES` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#retries CloudflareProvider#retries}
	Retries *float64 `field:"optional" json:"retries" yaml:"retries"`
	// RPS limit to apply when making calls to the API. Alternatively, can be configured using the `CLOUDFLARE_RPS` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs#rps CloudflareProvider#rps}
	Rps *float64 `field:"optional" json:"rps" yaml:"rps"`
}

