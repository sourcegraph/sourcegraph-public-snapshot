package provider


type SentryProviderConfig struct {
	// Alias name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs#alias SentryProvider#alias}
	Alias *string `field:"optional" json:"alias" yaml:"alias"`
	// The target Sentry Base API URL in the format `https://[hostname]/api/`.
	//
	// The default value is `https://sentry.io/api/`. The value must be provided when working with Sentry On-Premise. The value can be sourced from the `SENTRY_BASE_URL` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs#base_url SentryProvider#base_url}
	BaseUrl *string `field:"optional" json:"baseUrl" yaml:"baseUrl"`
	// The authentication token used to connect to Sentry. The value can be sourced from the `SENTRY_AUTH_TOKEN` environment variable.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/jianyuan/sentry/0.12.3/docs#token SentryProvider#token}
	Token *string `field:"optional" json:"token" yaml:"token"`
}

